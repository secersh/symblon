package router

import (
	"github.com/gin-gonic/gin"
	"undercover/internal/app/ingestor/handler"
	"undercover/internal/pkg/messaging"
)

func SetupRouter(mgg messaging.MessagingService) *gin.Engine {
	router := gin.Default()
	githubHandler := handler.NewGitHubHandler(mgg)

	apiV1 := router.Group("api/v1")
	{
		apiV1.GET("/github", githubHandler.Handle)
	}

	return router
}
