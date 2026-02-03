// SPDX-License-Identifier: EUPL-1.2

package handlers

import (
	"IRIS-Server/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResourcesHandler registers resource-related HTTP endpoints
func ResourcesHandler(router *gin.Engine) {
	resourcesGroup := router.Group("/resources")

	resourcesGroup.GET("/", listResources)
}

// listResources returns all MCP resources
// @Summary Get all resources
// @Description Returns a list of all resources from the MCP system
// @Tags resources
// @Produce json
// @Success 200 {array} models.Resource "List of resources"
// @Failure 500 {object} map[string]string "Failed to fetch resources"
// @Router /resources/ [get]
func listResources(c *gin.Context) {
	resources, err := repository.GetAllResources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch resources"})
		return
	}

	c.JSON(http.StatusOK, resources)
}
