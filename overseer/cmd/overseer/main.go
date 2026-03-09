package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"overseer/internal/overseer/router"
	"overseer/internal/overseer/store"
	"overseer/internal/overseer/trigger"
	"overseer/pkg/messaging"
)

func main() {
	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// ── Storage (in-memory placeholder; swap in EventStoreDB/Postgres later) ──
	agentStore := store.NewInMemoryAgentStore()
	evalStore := store.NewInMemoryEvaluationStore()

	// ── Messaging ────────────────────────────────────────────────────────────
	mgg, err := messaging.NewRabbitMQService(amqpURL)
	if err != nil {
		log.Fatalf("overseer: connect RabbitMQ: %v", err)
	}
	defer func() {
		if err := mgg.Close(); err != nil {
			log.Printf("overseer: close RabbitMQ: %v", err)
		}
	}()

	// ── Trigger consumer ─────────────────────────────────────────────────────
	// StubEvaluator is used until DuckDB/Parquet persistence is wired up.
	// Replace with trigger.DuckDBEvaluator(dataDir) once storage is ready.
	trig := trigger.New(mgg, agentStore, evalStore, trigger.StubEvaluator{})
	if err := trig.Start(); err != nil {
		log.Fatalf("overseer: start trigger: %v", err)
	}
	log.Printf("overseer: trigger consuming activity.# → queue=overseer")

	// ── HTTP API ─────────────────────────────────────────────────────────────
	r := router.SetupRouter(agentStore, evalStore)

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

	// ── Graceful shutdown ────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("overseer: shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("overseer: forced shutdown: %v", err)
	}

	log.Println("overseer: exited")
}
