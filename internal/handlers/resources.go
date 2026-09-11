// SPDX-License-Identifier: EUPL-1.2

package handlers

import (
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResourcesHandler registers resource-related HTTP endpoints
func ResourcesHandler(router *gin.Engine) {
	resourcesGroup := router.Group("/resources")

	resourcesGroup.GET("/", listResources)
	resourcesGroup.PUT("/:resourceId/position", setResourcePosition)
}

// listResources returns all MCP resources
// @Summary Get all resources
// @Description Returns a list of all resources from the MCP system
// @Tags resources
// @Produce json
// @Success 200 {array} models.TableauResource "List of Tableau resources"
// @Failure 500 {object} map[string]string "Failed to fetch resources"
// @Router /resources/ [get]
func listResources(c *gin.Context) {
	var resources []models.TableauResource

	resources, err := repository.GetAllResources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch resources"})
		return
	}

	c.JSON(http.StatusOK, resources)
}

// setResourcePosition updates the position of a specific resource
// @Summary Set resource position
// @Description Updates the position (longitude and latitude) of a specific resource
// @Tags resources
// @Accept json
// @Produce json
// @Param resourceId path string true "Resource ID"
// @Param position body PositionRequest true "Position data"
// @Success 200 {object} map[string]string "Resource position updated successfully"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 500 {object} map[string]string "Failed to update resource position"
// @Router /resources/{resourceId}/position [put]
func setResourcePosition(c *gin.Context) {
	var position PositionRequest

	err := c.ShouldBindJSON(&position)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidRequestBody})
		return
	}

	resourceID := c.Param("resourceId")
	err = repository.UpdateResourcePosition(resourceID, position.Longitude, position.Latitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update resource position"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Resource position updated successfully"})
}
