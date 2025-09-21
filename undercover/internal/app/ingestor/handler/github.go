package handler

import (
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
	h.mgg.Publish("GitHub webhook received")

	log.Print(ctx.Request.Body)

	ctx.JSON(200, gin.H{
		"message": "GitHub webhook received",
	})
}
