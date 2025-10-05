package main

import (
	"log"

	"undercover/internal/app/ingestor/router"
	"undercover/internal/pkg/messaging"
)

func main() {
	mgg, mggErr := messaging.NewRabbitMQService("amqp://guest:guest@localhost:5672/")

	if mggErr != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", mggErr)
	}

	router := router.SetupRouter(mgg)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
