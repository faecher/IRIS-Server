package handlers

import (
	"IRIS-Server/internal/chirpstack"
	"IRIS-Server/internal/mcp_control"
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"math"
	"net/http"

	"github.com/chirpstack/chirpstack/api/go/v4/integration"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
)

func GatewayHandler(router *gin.Engine) {
	gatewayGroup := router.Group("/chirpstackGateway")

	gatewayGroup.POST("/data", handleChirpstackWebhook)
}

// handleChirpstackWebhook handles incoming Chirpstack HTTP integration webhooks
// @Summary Chirpstack webhook endpoint
// @Description Receives uplink events from Chirpstack HTTP integration, parses tracker data, and updates database and MCP
// @Tags gateway
// @Accept json
// @Produce json
// @Param event query string true "Event type (must be 'up')"
// @Param payload body object true "Chirpstack uplink event payload"
// @Success 200 {object} map[string]string "Event processed successfully"
// @Failure 400 {object} map[string]string "Unsupported event type, invalid JSON, or missing DevEUI"
// @Failure 500 {object} map[string]string "Database or MCP update error"
// @Router /chirpstackGateway/data [post]
func handleChirpstackWebhook(c *gin.Context) {
	eventType := c.Query("event")
	if eventType != "up" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported event type"})
		return
	}

	var upMessage integration.UplinkEvent
	err := c.BindJSON(&upMessage)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload: " + err.Error()})
		return
	}

	eui := upMessage.GetDeviceInfo().GetDevEui()
	if eui == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing DevEUI in payload"})
		return
	}

	trackerID, err := repository.GetTrackerByDevEUI(eui)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	// Parse Chirpstack uplink message into tracker model
	tracker := models.BaseTracker{
		ID:      trackerID,
		Battery: -1,
		Position: models.Position{
			Latitude:  math.Inf(-1),
			Longitude: math.Inf(-1),
		},
	}
	chirpstack.ParseChirpstackTrackerMessage(&upMessage, &tracker)

	if trackerID == uuid.Nil { // Tracker unknown
		chirpstackTracker := models.ChirpstackTracker{
			BaseTracker: tracker, DevEUI: eui,
		}

		err := repository.CreateChirpstackTracker(&chirpstackTracker)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tracker: " + err.Error()})
			return
		}
	} else {
		err := repository.UpdateTracker(tracker)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tracker: " + err.Error()})
			return
		}
	}

	err = mcp_control.UpdateTrackerInMCP(tracker.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tracker in MCP: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
