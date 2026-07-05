package traccar

import (
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
	slog.Info("Traccar message handled successfully", "message", message)

	updateTrackers(ctx, message)

	return nil
}

func updateTrackers(ctx context.Context, message traccarMessage) {
	// Go through all devices in list and update the database registration
	for _, device := range message.Devices {
		slog.Info("Traccar device update received", "device_id", device.ID, "status", device.Status)

		trackerID, err := repository.GetTrackerIDByTraccarID(ctx, device.ID)
		if err != nil {
			slog.Error("Failed to get tracker ID by Traccar ID", "error", err, "traccar_id", device.ID)
			continue
		}

		if trackerID == uuid.Nil {
			err = repository.CreateTraccarTracker(ctx, &models.TraccarTracker{
				BaseTracker: models.BaseTracker{
					Name:    device.Name,
					Battery: -1,
					Position: models.Position{
						Latitude:  0.0,
						Longitude: 0.0,
					},
				},
				TraccarID: device.ID,
			})
			slog.Info("Created new tracker for Traccar device", "traccar_id", device.ID)

			if err != nil {
				slog.Error("Failed to create tracker for Traccar device", "error", err, "traccar_id", device.ID)
			}
		}
	}

	for _, position := range message.Positions {
		trackerID, err := repository.GetTrackerIDByTraccarID(ctx, position.DeviceID)
		if err != nil {
			slog.Error("Failed to get tracker ID by Traccar ID", "error", err, "traccar_id", position.DeviceID)
			continue
		}

		if trackerID == uuid.Nil {
			slog.Warn("Received position update for unknown Traccar device", "traccar_id", position.DeviceID)
			continue
		}

		batteryLevel, ok := position.Attributes["batteryLevel"].(float64)
		if !ok {
			batteryLevel = -1
		}
		tracker := models.BaseTracker{
			ID: trackerID,
			Position: models.Position{
				Latitude:  position.Latitude,
				Longitude: position.Longitude,
			},
			Battery:    int16(batteryLevel),
			LastUpdate: position.ServerTime,
		}

		err = repository.UpdateTracker(ctx, tracker)
		if err != nil {
			slog.Error("Failed to update tracker position in DB", "error", err, "tracker_id", trackerID)
			continue
		}

		slog.Debug("Updated tracker position in DB",
			"tracker_id", trackerID,
			"latitude", position.Latitude,
			"longitude", position.Longitude,
			"battery", batteryLevel,
		)
	}
}
