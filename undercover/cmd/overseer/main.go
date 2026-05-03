package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"undercover/internal/overseer/evaluator"
	"undercover/internal/overseer/installconsumer"
	"undercover/internal/overseer/resolvedconsumer"
	"undercover/internal/overseer/router"
	"undercover/internal/overseer/store"
	"undercover/internal/overseer/trigger"
	"undercover/pkg/messaging"
)

func main() {
	ctx := context.Background()

	postgresDSN := os.Getenv("POSTGRES_DSN")
	if postgresDSN == "" {
		postgresDSN = "postgres://postgres:postgres@localhost:5432/overseer?sslmode=disable"
	}

	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// ── Storage ───────────────────────────────────────────────────────────────
	rulesStore, err := store.NewPostgresInstalledRuleStore(ctx, postgresDSN)
	if err != nil {
		log.Fatalf("overseer: connect postgres: %v", err)
	}
	defer rulesStore.Close()

	if err := rulesStore.Migrate(ctx); err != nil {
		log.Fatalf("overseer: migrate: %v", err)
	}
	log.Println("overseer: schema ready")

	evalStore := store.NewInMemoryEvaluationStore()

	// ── Messaging ─────────────────────────────────────────────────────────────
	mq, err := messaging.NewRabbitMQService(amqpURL)
	if err != nil {
		log.Fatalf("overseer: connect RabbitMQ: %v", err)
	}
	defer func() {
		if err := mq.Close(); err != nil {
			log.Printf("overseer: close RabbitMQ: %v", err)
		}
	}()

	// ── DuckDB evaluator ──────────────────────────────────────────────────────
	s3cfg := evaluator.S3Config{
		Endpoint:     os.Getenv("S3_ENDPOINT"),
		Region:       envOr("S3_REGION", "auto"),
		Bucket:       os.Getenv("S3_BUCKET"),
		EventsPrefix: "events",
		AgentsPrefix: "agents",
		AccessKey:    os.Getenv("S3_ACCESS_KEY"),
		SecretKey:    os.Getenv("S3_SECRET_KEY"),
	}

	eval, err := evaluator.New(s3cfg)
	if err != nil {
		log.Fatalf("overseer: init evaluator: %v", err)
	}
	defer eval.Close()
	log.Println("overseer: DuckDB evaluator ready")

	// ── Install consumer ──────────────────────────────────────────────────────
	ic := installconsumer.New(rulesStore, mq)
	if err := ic.Start(); err != nil {
		log.Fatalf("overseer: start install consumer: %v", err)
	}
	log.Println("overseer: install consumer listening on agent.installed / agent.uninstalled")

	// ── Trigger consumer ──────────────────────────────────────────────────────
	trig := trigger.New(mq, rulesStore, evalStore, eval)
	if err := trig.Start(); err != nil {
		log.Fatalf("overseer: start trigger: %v", err)
	}
	log.Println("overseer: trigger consuming parquet.flushed")

	// ── Resolved consumer ─────────────────────────────────────────────────────
	rc := resolvedconsumer.New(rulesStore, mq)
	if err := rc.Start(); err != nil {
		log.Fatalf("overseer: start resolved consumer: %v", err)
	}
	log.Println("overseer: resolved consumer listening on agent.resolved.#")

	// ── HTTP API ──────────────────────────────────────────────────────────────
	r := router.SetupRouter()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("overseer: http server: %v", err)
		}
	}()
	log.Printf("overseer: HTTP API listening on :%s", port)

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("overseer: shutting down...")

	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		log.Fatalf("overseer: forced shutdown: %v", err)
	}

	log.Println("overseer: exited")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
