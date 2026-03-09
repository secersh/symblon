package handler

import (
	"io"
	"log"
	"net/http"
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

	body, err := io.ReadAll(ctx.Request.Body)

	if err != nil {
		log.Printf("Failed to read request body: %v", err)

		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})

		return
	}

	err = h.mgg.Publish(string(body))

	if err != nil {
		log.Printf("Failed to publish message: %v", err)

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process webhook",
		})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "GitHub webhook received",
	})
}
