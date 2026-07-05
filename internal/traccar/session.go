package traccar

import (
	"IRIS-Server/internal/mcpcontrol"
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/websocket"
)

func runSocketSession(ctx context.Context, conn *websocket.Conn) error {
	slog.Info("Connected to Traccar websocket, listening for messages...")

	// set up ping/pong handlers to detect dead connections
	err := conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	pingErr := make(chan error, 1)
	go handlePingPong(ctx, pingTicker, pingErr, conn)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		case err := <-pingErr:
			return fmt.Errorf("ping failed: %w", err)
		default:
			_, payload, err := conn.ReadMessage()
			if err != nil {
				return fmt.Errorf("read failed: %w", err)
			}

			err = handleTraccarMessage(ctx, payload)
			if err != nil {
				return fmt.Errorf("failed to handle traccar payload: %w", err)
			}
		}
	}
}

func handlePingPong(ctx context.Context, pingTicker *time.Ticker, pingErr chan error, conn *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			_ = conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "shutdown"),
				time.Now().Add(writeWait),
			)
			return
		case <-pingTicker.C:
			err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait))
			if err != nil {
				pingErr <- err
				return
			}
		}
	}
}

func handleTraccarMessage(ctx context.Context, payload []byte) error {
	var message traccarMessage

	err := json.Unmarshal(payload, &message)
	if err != nil {
		return fmt.Errorf("failed to parse websocket payload as traccar message: %w", err)
	}

	updateTrackers(ctx, message.Devices, message.Positions)

	return nil
}

func updateTrackers(ctx context.Context, devices []device, positions []position) {
	for _, device := range devices {
		slog.Debug("Traccar device update received", "device_id", device.ID, "status", device.Status)

		trackerID, err := repository.GetTrackerIDByTraccarID(ctx, device.ID)
		if err != nil {
			return
		}

		devicePosition := models.Position{
			Latitude:  math.Inf(-1),
			Longitude: math.Inf(-1),
		}
		if device.PositionID != nil {
			for _, position := range positions {
				if position.ID == *device.PositionID {
					devicePosition.Latitude = position.Latitude
					devicePosition.Longitude = position.Longitude
					break
				}
			}
		}

		// Parse Chirpstack uplink message into tracker model
		tracker := models.BaseTracker{
			ID:       trackerID,
			Battery:  -1, // TODO: parse battery from attributes, might be key batteryLevel or battery, to be tested
			Position: devicePosition,
		}

		err = updateTrackerPositionInDB(ctx, &tracker, device.ID)
		if err != nil {
			return
		}

		// TODO: this code is cloned from handlers/chirpstackGateway.go
		// TODO: this should be a generic function that can be called from both places to avoid code duplication
		// Skip MCP update if no resource is assigned
		trackerData, err := repository.GetTrackerByID(ctx, tracker.ID)
		if err != nil {
			// c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tracker: " + err.Error()})
			continue
		}

		if trackerData.TableauResource == nil {
			// No resource assigned, skip MCP update
			// c.JSON(http.StatusOK, gin.H{"status": "success", "note": "no resource assigned, MCP update skipped"})
			continue
		}

		// Resource assigned, proceed with MCP update
		// Update tracker marker in MCP
		err = mcpcontrol.UpdateMarkerInMCP(ctx, tracker.ID)
		if err != nil {
			slog.Error("Failed to update tracker marker in MCP", "trackerID", tracker.ID, "error", err)
			// c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tracker marker in MCP: " + err.Error()})
			continue
		}

		// c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func updateTrackerPositionInDB(ctx context.Context, tracker *models.BaseTracker, traccarID int64) error {
	if tracker.ID == uuid.Nil { // Tracker unknown
		chirpstackTracker := models.TraccarTracker{
			BaseTracker: *tracker, TraccarID: traccarID,
		}

		err := repository.CreateTraccarTracker(ctx, &chirpstackTracker)
		if err != nil {
			return fmt.Errorf("failed to create tracker: %w", err)
		}

		// Propagate the newly created tracker ID back to the caller's BaseTracker
		tracker.ID = chirpstackTracker.ID
		return nil
	}

	// Tracker known, update existing record
	err := repository.UpdateTracker(ctx, *tracker)
	if err != nil {
		return fmt.Errorf("failed to update tracker: %w", err)
	}

	return nil
}
