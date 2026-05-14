package traccar

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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

			// TODO: unmarshal and handle your Traccar payload
			// if err := handleTraccarMessage(payload); err != nil { ... }
			_ = payload
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
