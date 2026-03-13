package router

import (
	"github.com/gin-gonic/gin"

	"undercover/internal/overseer/handler"
	"undercover/internal/overseer/store"
)

// SetupRouter wires up all HTTP routes and returns a configured Gin engine.
//
// API surface:
//
//	GET    /api/v1/agents        — list all registered agents
//	POST   /api/v1/agents        — register a new agent
//	GET    /api/v1/agents/:id    — get agent by ID
//	PUT    /api/v1/agents/:id    — update agent definition
//	DELETE /api/v1/agents/:id    — remove an agent
func SetupRouter(agents store.AgentStore) *gin.Engine {
	r := gin.Default()

	h := handler.NewAgentHandler(agents)

	v1 := r.Group("/api/v1")
	{
		agentRoutes := v1.Group("/agents")
		{
			agentRoutes.GET("", h.ListAgents)
			agentRoutes.POST("", h.RegisterAgent)
			agentRoutes.GET("/:id", h.GetAgent)
			agentRoutes.PUT("/:id", h.UpdateAgent)
			agentRoutes.DELETE("/:id", h.DeleteAgent)
		}
	}

	return r
}
