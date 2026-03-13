package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"undercover/internal/paquetier/writer"
	"undercover/pkg/event"
	"undercover/pkg/messaging"
)

const (
	queueName  = "paquetier"
	routingKey = "activity.#"
	flushEvery = 30 * time.Second
)

func main() {
	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/data"
	}

	// S3 config — all fields required to enable S3 backend.
	// If any are empty the writer falls back to local dataDir.
	var s3cfg *writer.S3Config
	if ep := os.Getenv("S3_ENDPOINT"); ep != "" {
		// Strip scheme — DuckDB httpfs expects host only; ssl is controlled via SET s3_use_ssl
		ep = strings.TrimPrefix(ep, "https://")
		ep = strings.TrimPrefix(ep, "http://")
		s3cfg = &writer.S3Config{
			Endpoint:  ep,
			Region:    getEnvOr("S3_REGION", "auto"),
			Bucket:    os.Getenv("S3_BUCKET"),
			Prefix:    getEnvOr("S3_PREFIX", "events"),
			AccessKey: os.Getenv("S3_ACCESS_KEY"),
			SecretKey: os.Getenv("S3_SECRET_KEY"),
		}
		log.Printf("Using S3 backend: %s/%s/%s", s3cfg.Endpoint, s3cfg.Bucket, s3cfg.Prefix)
	}

	mgg, err := messaging.NewRabbitMQService(amqpURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer func() {
		if err := mgg.Close(); err != nil {
			log.Printf("Failed to close RabbitMQ connection: %v", err)
		}
	}()

	if err = mgg.BindQueue(queueName, routingKey); err != nil {
		log.Fatalf("Failed to bind queue: %v", err)
	}

	w, err := writer.NewWriter(dataDir, s3cfg)
	if err != nil {
		log.Fatalf("Failed to create parquet writer: %v", err)
	}
	defer func() {
		if err := w.Close(); err != nil {
			log.Printf("Failed to close parquet writer: %v", err)
		}
	}()

	if err = mgg.Subscribe(queueName, func(message string) {
		var e event.WebhookEvent
		if err := json.Unmarshal([]byte(message), &e); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			return
		}
		if err := w.Insert(&e); err != nil {
			log.Printf("Failed to insert event: %v", err)
		}
	}); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	dest := dataDir
	if s3cfg != nil {
		dest = fmt.Sprintf("s3://%s/%s", s3cfg.Bucket, s3cfg.Prefix)
	}
	log.Printf("Paquetier started — consuming from queue %q, writing parquet to %s", queueName, dest)

	ticker := time.NewTicker(flushEvery)
	defer ticker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			if err := w.Flush(); err != nil {
				log.Printf("Periodic flush failed: %v", err)
			}
		case <-quit:
			log.Println("Shutting down paquetier...")
			return
		}
	}
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
