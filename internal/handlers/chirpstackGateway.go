package handlers

import (
	"IRIS-Server/internal/chirpstack"
	"IRIS-Server/internal/mcpcontrol"
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/chirpstack/chirpstack/api/go/v4/integration"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
)

// ErrMissingDevEUI indicates that the DevEUI is missing in the Chirpstack payload
var ErrMissingDevEUI = errors.New("missing DevEUI in payload")

// GatewayHandler registers all Chirpstack gateway HTTP endpoints
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

	upMessage, trackerID, eui, err := getTrackerAndEuiFromContext(c)
	if err != nil {
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
	err = chirpstack.ParseChirpstackTrackerMessage(upMessage, &tracker)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse tracker message: " + err.Error()})
		return
	}

	err = updateTrackerPositionInDB(c, tracker, eui)
	if err != nil {
		return
	}

	err = mcpcontrol.UpdateTrackerInMCP(tracker.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tracker in MCP: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// getTrackerAndEuiFromContext extracts the tracker and DevEUI from the Chirpstack uplink event in the request context
// This function also sets the appropriate HTTP error responses if extraction fails
func getTrackerAndEuiFromContext(c *gin.Context) (*integration.UplinkEvent, uuid.UUID, string, error) {
	var upMessage *integration.UplinkEvent
	err := c.BindJSON(&upMessage)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload: " + err.Error()})
		return upMessage, uuid.Nil, "", fmt.Errorf("invalid JSON payload: %w", err)
	}

	eui := upMessage.GetDeviceInfo().GetDevEui()
	if eui == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing DevEUI in payload"})
		return upMessage, uuid.Nil, "", ErrMissingDevEUI
	}

	trackerID, err := repository.GetTrackerByDevEUI(eui)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return upMessage, uuid.Nil, "", fmt.Errorf("database error: %w", err)
	}

	return upMessage, trackerID, eui, nil
}

// updateTrackerPositionInDB creates or updates the tracker record in the database
// This function also sets the appropriate HTTP error responses if database operations fail
func updateTrackerPositionInDB(c *gin.Context, tracker models.BaseTracker, eui string) error {
	if tracker.ID == uuid.Nil { // Tracker unknown
		chirpstackTracker := models.ChirpstackTracker{
			BaseTracker: tracker, DevEUI: eui,
		}

		err := repository.CreateChirpstackTracker(&chirpstackTracker)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tracker: " + err.Error()})
			return fmt.Errorf("failed to create tracker: %w", err)
		}
	} else {
		err := repository.UpdateTracker(tracker)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tracker: " + err.Error()})
			return fmt.Errorf("failed to update tracker: %w", err)
		}
	}

	return nil
}
