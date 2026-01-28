package handlers

import (
	"IRIS-Server/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ResourcesHandler(router *gin.Engine) {
	resourcesGroup := router.Group("/resources")

	resourcesGroup.GET("/", listResources)
}

// listResources returns all MCP resources
// GET /resources/
func listResources(c *gin.Context) {
	resources, err := repository.GetAllResources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch resources"})
		return
	}

	c.JSON(http.StatusOK, resources)
}
