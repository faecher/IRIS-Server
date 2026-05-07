// SPDX-License-Identifier: EUPL-1.2

// Package traccar provides utilities for processing messages from Traccar
package traccar

import "context"

// RunTraccarWebsocketListener connects to the Traccar websocket and handles incomping messages.
// This function will run until the given context is Done
func RunTraccarWebsocketListener(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// TODO: obtain / set session auth cookie(?)
			// TODO. obtain server config
			// TODO: start socket connection
			// TODO: handle incoming messages
			return
		}
	}
}
