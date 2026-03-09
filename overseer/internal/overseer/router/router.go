package router

import (
	"github.com/gin-gonic/gin"

	"overseer/internal/overseer/handler"
	"overseer/internal/overseer/store"
)

// SetupRouter wires up all HTTP routes and returns a configured Gin engine.
//
// API surface:
//
//	GET    /api/v1/agents                      — list all registered agents
//	POST   /api/v1/agents                      — register a new agent
//	GET    /api/v1/agents/:id                  — get agent by ID
//	PUT    /api/v1/agents/:id                  — update agent definition
//	DELETE /api/v1/agents/:id                  — remove an agent
//	GET    /api/v1/agents/:id/evaluations      — list evaluations for an agent
//	GET    /api/v1/actors/:login/evaluations   — list evaluations for an actor
func SetupRouter(agents store.AgentStore, evals store.EvaluationStore) *gin.Engine {
	r := gin.Default()

	h := handler.NewAgentHandler(agents, evals)

	v1 := r.Group("/api/v1")
	{
		agentRoutes := v1.Group("/agents")
		{
			agentRoutes.GET("", h.ListAgents)
			agentRoutes.POST("", h.RegisterAgent)
			agentRoutes.GET("/:id", h.GetAgent)
			agentRoutes.PUT("/:id", h.UpdateAgent)
			agentRoutes.DELETE("/:id", h.DeleteAgent)
			agentRoutes.GET("/:id/evaluations", h.ListAgentEvaluations)
		}

		actorRoutes := v1.Group("/actors")
		{
			actorRoutes.GET("/:login/evaluations", h.ListActorEvaluations)
		}
	}

	return r
}
