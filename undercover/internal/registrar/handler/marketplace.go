package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"undercover/internal/registrar/store"
)

// MarketplaceHandler handles marketplace listing and agent detail requests.
type MarketplaceHandler struct {
	agents store.AgentStore
}

// NewMarketplaceHandler returns a MarketplaceHandler wired to the given store.
func NewMarketplaceHandler(agents store.AgentStore) *MarketplaceHandler {
	return &MarketplaceHandler{agents: agents}
}

// ListAgents handles GET /api/v1/agents.
// Optional query param: ?publisher=<handle>
// Returns agents with their symbols included.
func (h *MarketplaceHandler) ListAgents(c *gin.Context) {
	agents, err := h.agents.ListAgents(c.Request.Context(), c.Query("publisher"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ids := make([]string, len(agents))
	for i, a := range agents {
		ids[i] = a.ID
	}
	symbolMap, err := h.agents.SymbolsByAgentIDs(c.Request.Context(), ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type agentWithSymbols struct {
		*store.Agent
		Symbols []store.Symbol `json:"symbols"`
	}
	result := make([]agentWithSymbols, len(agents))
	for i, a := range agents {
		syms := symbolMap[a.ID]
		if syms == nil {
			syms = []store.Symbol{}
		}
		result[i] = agentWithSymbols{Agent: a, Symbols: syms}
	}

	c.JSON(http.StatusOK, result)
}

// GetAgent handles GET /api/v1/agents/:publisher/:handle/:version.
func (h *MarketplaceHandler) GetAgent(c *gin.Context) {
	agent, err := h.agents.GetAgent(
		c.Request.Context(),
		c.Param("publisher"),
		c.Param("handle"),
		c.Param("version"),
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	symbols, err := h.agents.SymbolsByAgent(c.Request.Context(), agent.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"agent": agent, "symbols": symbols})
}
