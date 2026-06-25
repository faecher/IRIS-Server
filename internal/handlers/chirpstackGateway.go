// SPDX-License-Identifier: EUPL-1.2

package handlers

import (
	"IRIS-Server/internal/chirpstack"
	"IRIS-Server/internal/mcpcontrol"
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
)

// ErrMissingDevEUI indicates that the DevEUI is missing in the Chirpstack payload
var ErrMissingDevEUI = errors.New("missing DevEUI in payload")

const statusSuccess = "success"

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
		slog.Info("Received unsupported event type", "event", eventType)
		return
	}

	upMessage, trackerID, eui, err := getTrackerAndEuiFromContext(c)
	if err != nil {
		return
	} else if upMessage.Object.Invalid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported message type"})
		slog.Info("Received unsupported message type", "DevEUI", eui)
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

	// Check if message is newer than existing data (ignore buffered old messages)
	if !verifyTrackerMessageFreshness(c, &tracker) {
		return
	}

	err = updateTrackerPositionInDB(c, &tracker, eui)
	if err != nil {
		slog.Error("Failed to update position in DB", "error", err)
		return
	}

	// Skip MCP update if no resource is assigned
	trackerData, err := repository.GetTrackerByID(tracker.ID)
	if err != nil {
		slog.Error("Failed to get tracker after update")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tracker: " + err.Error()})
		return
	}

	if trackerData.TableauResource == nil {
		// No resource assigned, skip MCP update
		slog.Debug("No resource assigned to tracker, skipping MCP update", "trackerID", tracker.ID, "trackerName", tracker.Name)
		c.JSON(http.StatusOK, gin.H{"status": statusSuccess, "note": "no resource assigned, MCP update skipped"})
		return
	}

	// Resource assigned, proceed with MCP update
	// Update tracker marker in MCP
	err = mcpcontrol.UpdateMarkerInMCP(tracker.ID)
	if errors.Is(err, mcpcontrol.ErrMCPDisabled) {
		slog.Debug("MCP is disabled, skipping update", "trackerID", tracker.ID)
		c.JSON(http.StatusOK, gin.H{"status": statusSuccess, "note": "MCP is disabled, update skipped"})
		return
	}
	if err != nil {
		slog.Error("Failed to update tracker marker in MCP", "trackerID", tracker.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tracker marker in MCP: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": statusSuccess})
}

// getTrackerAndEuiFromContext extracts the tracker and DevEUI from the Chirpstack uplink event in the request context
// This function also sets the appropriate HTTP error responses if extraction fails
// if no tracker is found, trackerID will be uuid.Nil
func getTrackerAndEuiFromContext(c *gin.Context) (*models.ChirpstackUplink, uuid.UUID, string, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body: " + err.Error()})
		slog.Error("Failed to read request body", "error", err)
		return nil, uuid.Nil, "", fmt.Errorf("read body: %w", err)
	}

	upMessage := &models.ChirpstackUplink{}
	err = json.Unmarshal(body, upMessage)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid protobuf JSON payload: " + err.Error()})
		slog.Error("Failed to bind protobuf JSON payload", "error", err)
		return upMessage, uuid.Nil, "", fmt.Errorf("invalid protobuf JSON payload: %w", err)
	}

	eui := upMessage.DeviceInfo.DevEui
	if eui == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing DevEUI in payload"})
		slog.Error("Missing DevEUI in payload")
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
// the returned ID is the tracker's UUID. Upon error, uuid.Nil is returned
func updateTrackerPositionInDB(c *gin.Context, tracker *models.BaseTracker, eui string) error {
	if tracker.ID == uuid.Nil { // Tracker unknown
		chirpstackTracker := models.ChirpstackTracker{
			BaseTracker: *tracker, DevEUI: eui,
		}

		err := repository.CreateChirpstackTracker(&chirpstackTracker)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tracker: " + err.Error()})
			return fmt.Errorf("failed to create tracker: %w", err)
		}

		// Propagate the newly created tracker ID back to the caller's BaseTracker
		tracker.ID = chirpstackTracker.ID
		return nil
	}

	// Tracker known, update existing record
	err := repository.UpdateTracker(*tracker)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tracker: " + err.Error()})
		return fmt.Errorf("failed to update tracker: %w", err)
	}

	return nil
}

func verifyTrackerMessageFreshness(c *gin.Context, newData *models.BaseTracker) bool {
	if newData.ID != uuid.Nil && !newData.LastUpdate.IsZero() {
		existingTracker, err := repository.GetTrackerByID(newData.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get existing tracker: " + err.Error()})
			return false
		}
		if !existingTracker.LastUpdate.Before(newData.LastUpdate) {
			slog.Debug("Received outdated message", "newData", newData, "existingData", existingTracker)
			// Existing data is newer or same, ignore this message
			c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "message older than existing data"})
			return false
		}
	}

	return true
}
