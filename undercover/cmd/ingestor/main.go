package main

import (
	"log"

	"undercover/internal/app/ingestor/router"
	"undercover/internal/pkg/messaging"
)

func main() {
	mgg, nil := messaging.NewRabbitMQService("amqp://guest:guest@localhost:5672/")
	router := router.SetupRouter(mgg)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
