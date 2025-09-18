package main

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()

	v1 := router.Group("api/v1")
	{
		v1.GET("/echo", handleEcho)
	}

	router.Run(":8080")
}

func handleEcho(c *gin.Context) {
}
