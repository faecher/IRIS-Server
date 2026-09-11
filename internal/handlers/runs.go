package handlers

import (
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RunsHandler registers run-related HTTP endpoints
func RunsHandler(router *gin.Engine) {
	runsGroup := router.Group("/runs")

	runsGroup.GET("/", listRuns)
	runsGroup.POST("/:runId/position", setRunPosition)
}

// listRuns returns all MCP runs
// @Summary Get all runs
// @Description Returns a list of all runs from the MCP system
// @Tags runs
// @Produce json
// @Success 200 {array} models.MCPRun "List of MCP runs"
// @Failure 500 {object} map[string]string "Failed to fetch runs"
// @Router /runs/ [get]
func listRuns(c *gin.Context) {
	var runs []models.MCPRun
	runs, err := repository.GetAllRuns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch runs"})
		return
	}

	c.JSON(http.StatusOK, runs)
}

// setRunPosition sets the position of a run
// @Summary Set run position
// @Description Sets the position of a specific run on the map.
// The position is provided in the request body as a JSON object containing latitude and longitude.
// @Tags runs
// @Accept json
// @Produce json
// @Param runId path string true "Run ID"
// @Param position body PositionRequest true "Position data"
// @Success 200 {object} map[string]string "Position set successfully"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Failed to set run position"
// @Router /runs/{runId}/position [post]
func setRunPosition(c *gin.Context) {
	// Get position from request body
	var position PositionRequest
	err := c.ShouldBindJSON(&position)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidRequestBody})
		return
	}

	runID := c.Param("runId")

	err = repository.UpdateRunPosition(runID, position.Latitude, position.Longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set run position"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Position set successfully"})
}
