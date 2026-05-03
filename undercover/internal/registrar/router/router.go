package router

import (
	"github.com/gin-gonic/gin"

	"undercover/internal/registrar/handler"
	"undercover/internal/registrar/store"
	"undercover/internal/registrar/upload"
	"undercover/pkg/auth"
	"undercover/pkg/messaging"
)

// SetupRouter wires up all registrar HTTP routes.
//
// API surface:
//
//	POST   /api/v1/agents                                     — publish an agent package
//	GET    /api/v1/agents                                     — list marketplace agents
//	GET    /api/v1/agents/:publisher/:handle/:version         — agent detail
//	POST   /api/v1/agents/:publisher/:handle/:version/install — install an agent
//	DELETE /api/v1/agents/:publisher/:handle/:version/install — uninstall an agent
func SetupRouter(agents store.AgentStore, installs store.InstallStore, symbols store.SymbolStore, uploader upload.Uploader, mq messaging.MessagingService, jwks *auth.JWKS) *gin.Engine {
	r := gin.Default()

	publish := handler.NewPublishHandler(agents, uploader)
	market := handler.NewMarketplaceHandler(agents)
	install := handler.NewInstallHandler(agents, installs, mq)
	sym := handler.NewSymbolsHandler(symbols)

	v1 := r.Group("/registrar/v1")
	{
		// Public — no auth required
		v1.GET("/agents", market.ListAgents)
		v1.GET("/agents/:publisher/:handle/:version", market.GetAgent)

		// Authenticated
		protected := v1.Group("", jwks.RequireAuth())
		{
			protected.POST("/agents", publish.Publish)
			protected.GET("/me/agents", install.ListInstalled)
			protected.GET("/me/owned", install.ListOwned)
			protected.GET("/me/symbols", sym.ListIssued)

			agent := protected.Group("/agents/:publisher/:handle/:version")
			{
				agent.POST("/install", install.Install)
				agent.DELETE("/install", install.Uninstall)
			}
		}
	}

	return r
}
