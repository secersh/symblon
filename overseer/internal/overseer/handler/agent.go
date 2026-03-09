package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"overseer/internal/overseer/agent"
	"overseer/internal/overseer/store"
)

// AgentHandler handles HTTP requests for the agent registry.
type AgentHandler struct {
	agents store.AgentStore
	evals  store.EvaluationStore
}

// NewAgentHandler creates a new AgentHandler.
func NewAgentHandler(agents store.AgentStore, evals store.EvaluationStore) *AgentHandler {
	return &AgentHandler{agents: agents, evals: evals}
}

// registerAgentRequest is the request body for POST /api/v1/agents.
type registerAgentRequest struct {
	Name        string        `json:"name"        binding:"required"`
	Description string        `json:"description"`
	Scope       agent.Scope   `json:"scope"       binding:"required"`
	OwnerID     string        `json:"owner_id"`
	SymbolPath  string        `json:"symbol_path" binding:"required"`
	Trigger     agent.Trigger `json:"trigger"     binding:"required"`
	Rule        agent.Rule    `json:"rule"        binding:"required"`
}

// ListAgents handles GET /api/v1/agents.
func (h *AgentHandler) ListAgents(c *gin.Context) {
	agents, err := h.agents.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, agents)
}

// RegisterAgent handles POST /api/v1/agents.
func (h *AgentHandler) RegisterAgent(c *gin.Context) {
	var req registerAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	a := &agent.Agent{
		ID:          uuid.NewString(),
		Name:        req.Name,
		Description: req.Description,
		Scope:       req.Scope,
		OwnerID:     req.OwnerID,
		SymbolPath:  req.SymbolPath,
		Trigger:     req.Trigger,
		Rule:        req.Rule,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.agents.Save(c.Request.Context(), a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, a)
}

// GetAgent handles GET /api/v1/agents/:id.
func (h *AgentHandler) GetAgent(c *gin.Context) {
	a, err := h.agents.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

// UpdateAgent handles PUT /api/v1/agents/:id.
func (h *AgentHandler) UpdateAgent(c *gin.Context) {
	existing, err := h.agents.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req registerAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.Scope = req.Scope
	existing.OwnerID = req.OwnerID
	existing.SymbolPath = req.SymbolPath
	existing.Trigger = req.Trigger
	existing.Rule = req.Rule
	existing.UpdatedAt = time.Now().UTC()

	if err := h.agents.Save(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, existing)
}

// DeleteAgent handles DELETE /api/v1/agents/:id.
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	if _, err := h.agents.Get(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err := h.agents.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ListAgentEvaluations handles GET /api/v1/agents/:id/evaluations.
func (h *AgentHandler) ListAgentEvaluations(c *gin.Context) {
	evals, err := h.evals.ListByAgent(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, evals)
}

// ListActorEvaluations handles GET /api/v1/actors/:login/evaluations.
func (h *AgentHandler) ListActorEvaluations(c *gin.Context) {
	evals, err := h.evals.ListByActor(c.Request.Context(), c.Param("login"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, evals)
}
