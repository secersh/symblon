package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"undercover/internal/ingestor/normalize"
	"undercover/pkg/messaging"

	"github.com/gin-gonic/gin"
)

type GitHubHandler struct {
	mgg    messaging.MessagingService
	secret []byte // HMAC-SHA256 webhook secret; empty = verification disabled
}

func NewGitHubHandler(mgg messaging.MessagingService, secret string) *GitHubHandler {
	return &GitHubHandler{
		mgg:    mgg,
		secret: []byte(secret),
	}
}

func (h *GitHubHandler) Handle(ctx *gin.Context) {
	defer ctx.Request.Body.Close()

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(h.secret) > 0 {
		sig := ctx.GetHeader("X-Hub-Signature-256")
		if !verifySignature(h.secret, body, sig) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
	}

	eventType := ctx.GetHeader("X-GitHub-Event")
	deliveryID := ctx.GetHeader("X-GitHub-Delivery")

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

// verifySignature compares the GitHub-provided HMAC-SHA256 signature against
// the one computed from the raw body and the shared secret.
func verifySignature(secret, body []byte, sigHeader string) bool {
	const prefix = "sha256="
	if !strings.HasPrefix(sigHeader, prefix) {
		return false
	}
	got, err := hex.DecodeString(strings.TrimPrefix(sigHeader, prefix))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	expected := mac.Sum(nil)
	return hmac.Equal(expected, got)
}
