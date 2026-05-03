package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"undercover/internal/registrar/store"
	"undercover/pkg/auth"
)

// SymbolsHandler handles issued symbol queries.
type SymbolsHandler struct {
	symbols store.SymbolStore
}

// NewSymbolsHandler returns a SymbolsHandler.
func NewSymbolsHandler(symbols store.SymbolStore) *SymbolsHandler {
	return &SymbolsHandler{symbols: symbols}
}

// ListIssued handles GET /registrar/v1/me/symbols.
func (h *SymbolsHandler) ListIssued(c *gin.Context) {
	userID := auth.UserID(c)
	symbols, err := h.symbols.ListIssuedSymbols(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, symbols)
}
