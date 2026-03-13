package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"undercover/internal/ingestor/normalize"
	"undercover/pkg/messaging"

	"github.com/gin-gonic/gin"
)

type GitHubHandler struct {
	mgg messaging.MessagingService
}

func NewGitHubHandler(mgg messaging.MessagingService) *GitHubHandler {
	return &GitHubHandler{
		mgg: mgg,
	}
}

func (h *GitHubHandler) Handle(ctx *gin.Context) {
	defer ctx.Request.Body.Close()

	eventType := ctx.GetHeader("X-GitHub-Event")
	deliveryID := ctx.GetHeader("X-GitHub-Delivery")

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	webhookEvent, err := normalize.GitHub(eventType, deliveryID, body)
	if err != nil {
		log.Printf("Failed to normalize GitHub webhook: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
		return
	}

	eventJSON, err := json.Marshal(webhookEvent)
	if err != nil {
		log.Printf("Failed to marshal webhook event: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
		return
	}

	if err = h.mgg.Publish(string(eventJSON)); err != nil {
		log.Printf("Failed to publish message: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "GitHub webhook received"})
}
