package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
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

	w, err := writer.NewWriter(dataDir)
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

	log.Printf("Paquetier started — consuming from queue %q, writing parquet to %s", queueName, dataDir)

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
