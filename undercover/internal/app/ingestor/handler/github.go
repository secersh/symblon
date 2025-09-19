package handler

import (
	"github.com/gin-gonic/gin"
	"undercover/internal/pkg/messaging"
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

	ctx.JSON(200, gin.H{
		"message": "GitHub webhook received",
	})
}
