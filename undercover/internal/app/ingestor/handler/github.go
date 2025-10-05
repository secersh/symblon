package handler

import (
	"io"
	"log"
	"undercover/internal/pkg/messaging"

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
	body, err := io.ReadAll(ctx.Request.Body)

	if err != nil {
		log.Printf("Failed to read request body: %v", err)

		ctx.JSON(400, gin.H{
			"error": "Invalid request body",
		})

		return
	}

	err = h.mgg.Publish(string(body))

	if err != nil {
		log.Printf("Failed to publish message: %v", err)

		ctx.JSON(500, gin.H{
			"error": "Failed to process webhook",
		})

		return
	}

	ctx.JSON(200, gin.H{
		"message": "GitHub webhook received",
	})
}
