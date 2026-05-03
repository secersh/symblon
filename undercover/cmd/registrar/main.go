package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"undercover/internal/registrar/router"
	"undercover/internal/registrar/store"
	"undercover/internal/registrar/upload"
	"undercover/pkg/auth"
	"undercover/pkg/messaging"
)

func main() {
	ctx := context.Background()

	// ── Auth (Supabase JWKS) — read first so PublicBaseURL can use supabaseURL ──
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		log.Fatal("registrar: SUPABASE_URL is required")
	}

	postgresDSN := os.Getenv("POSTGRES_DSN")
	if postgresDSN == "" {
		postgresDSN = "postgres://postgres:postgres@localhost:5432/registrar?sslmode=disable"
	}

	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	// ── Storage ──────────────────────────────────────────────────────────────
	db, err := store.NewPostgresStore(ctx, postgresDSN)
	if err != nil {
		log.Fatalf("registrar: connect postgres: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(ctx); err != nil {
		log.Fatalf("registrar: migrate: %v", err)
	}
	log.Println("registrar: schema ready")

	// ── Object storage ───────────────────────────────────────────────────────
	s3Bucket := os.Getenv("S3_BUCKET")
	s3PublicBucket := envOr("S3_PUBLIC_BUCKET", s3Bucket)
	s3cfg := upload.S3Config{
		Endpoint:      os.Getenv("S3_ENDPOINT"),
		PublicBaseURL: strings.TrimRight(supabaseURL, "/") + "/storage/v1/object/public/" + s3PublicBucket,
		Region:        envOr("S3_REGION", "auto"),
		Bucket:        s3Bucket,
		PublicBucket:  s3PublicBucket,
		Prefix:        "agents",
		AccessKey:     os.Getenv("S3_ACCESS_KEY"),
		SecretKey:     os.Getenv("S3_SECRET_KEY"),
	}
	uploader := upload.NewS3Uploader(s3cfg)

	// ── Messaging ────────────────────────────────────────────────────────────
	mq, err := messaging.NewRabbitMQService(amqpURL)
	if err != nil {
		log.Fatalf("registrar: connect RabbitMQ: %v", err)
	}
	defer func() {
		if err := mq.Close(); err != nil {
			log.Printf("registrar: close RabbitMQ: %v", err)
		}
	}()

	// ── Auth (Supabase JWKS) ──────────────────────────────────────────────────
	jwks, err := auth.NewJWKS(ctx, supabaseURL)
	if err != nil {
		log.Fatalf("registrar: init JWKS: %v", err)
	}
	log.Println("registrar: JWKS ready")

	// ── HTTP API ─────────────────────────────────────────────────────────────
	r := router.SetupRouter(db, db, db, uploader, mq, jwks)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("registrar: http server: %v", err)
		}
	}()
	log.Printf("registrar: HTTP API listening on :%s", port)

	// ── Graceful shutdown ────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("registrar: shutting down...")

	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		log.Fatalf("registrar: forced shutdown: %v", err)
	}

	log.Println("registrar: exited")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
