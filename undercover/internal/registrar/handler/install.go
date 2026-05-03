package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"undercover/internal/registrar/store"
	"undercover/pkg/auth"
	"undercover/pkg/messaging"
)

// InstallHandler handles agent install and uninstall requests.
type InstallHandler struct {
	agents   store.AgentStore
	installs store.InstallStore
	mq       messaging.MessagingService
}

// NewInstallHandler returns an InstallHandler.
func NewInstallHandler(agents store.AgentStore, installs store.InstallStore, mq messaging.MessagingService) *InstallHandler {
	return &InstallHandler{agents: agents, installs: installs, mq: mq}
}

// Install handles POST /api/v1/agents/:publisher/:handle/:version/install.
func (h *InstallHandler) Install(c *gin.Context) {
	userID := auth.UserID(c)
	actorLogin := auth.PublisherName(c)

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

	inst, err := h.installs.Install(c.Request.Context(), userID, agent.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	symbols, err := h.agents.SymbolsByAgent(c.Request.Context(), agent.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.emitInstalled(userID, actorLogin, agent, symbols, inst.InstalledAt); err != nil {
		// Non-fatal: install is recorded, overseer will eventually catch up.
		// In production this should be retried or written to an outbox.
		_ = err
	}

	c.JSON(http.StatusCreated, inst)
}

// Uninstall handles DELETE /api/v1/agents/:publisher/:handle/:version/install.
func (h *InstallHandler) Uninstall(c *gin.Context) {
	userID := auth.UserID(c)

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

	if err := h.installs.Uninstall(c.Request.Context(), userID, agent.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	event := messaging.AgentUninstalledEvent{
		UserID:        userID,
		AgentID:       agent.Publisher + "/" + agent.Handle + "/" + agent.Version,
		UninstalledAt: time.Now().UTC(),
	}
	if body, err := json.Marshal(event); err == nil {
		_ = h.mq.PublishTo(messaging.RoutingKeyAgentUninstalled, string(body))
	}

	c.Status(http.StatusNoContent)
}

// ListInstalled handles GET /registrar/v1/me/agents — returns installed agents for the current user.
func (h *InstallHandler) ListInstalled(c *gin.Context) {
	userID := auth.UserID(c)
	agents, err := h.installs.ListInstalledAgents(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if agents == nil {
		agents = []*store.Agent{}
	}
	c.JSON(http.StatusOK, agents)
}

func (h *InstallHandler) emitInstalled(userID, actorLogin string, agent *store.Agent, symbols []store.Symbol, installedAt time.Time) error {
	installed := make([]messaging.InstalledSymbol, 0, len(symbols))
	for _, s := range symbols {
		installed = append(installed, messaging.InstalledSymbol{
			SymbolID:    s.SymbolID,
			SQLPath:     s.SQLPath,
			Type:        s.Type,
			WindowHours: s.WindowHours,
		})
	}

	event := messaging.AgentInstalledEvent{
		UserID:      userID,
		ActorLogin:  actorLogin,
		AgentID:     agent.Publisher + "/" + agent.Handle + "/" + agent.Version,
		Symbols:     installed,
		InstalledAt: installedAt,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return h.mq.PublishTo(messaging.RoutingKeyAgentInstalled, string(body))
}
