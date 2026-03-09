package router

import (
	"github.com/gin-gonic/gin"
	"undercover/internal/ingestor/handler"
	"undercover/pkg/messaging"
)

func SetupRouter(mgg messaging.MessagingService) *gin.Engine {
	router := gin.Default()
	githubHandler := handler.NewGitHubHandler(mgg)

	apiV1 := router.Group("api/v1")
	{
		apiV1.POST("/github", githubHandler.Handle)
	}

	return router
}
