// SPDX-License-Identifier: EUPL-1.2

// Package chirpstack provides utilities for processing messages from Chirpstack
package chirpstack

import (
	"IRIS-Server/internal/models"
	"errors"

	"github.com/chirpstack/chirpstack/api/go/v4/integration"
)

var (
	// ErrNoMessages indicates that there are no messages in the uplink event
	ErrNoMessages = errors.New("no messages in uplink event")
)

// ParseChirpstackTrackerMessage extracts battery and position data from a Chirpstack uplink event
func ParseChirpstackTrackerMessage(msg *integration.UplinkEvent, tracker *models.BaseTracker) error {
	objectMap := msg.GetObject().AsMap()

	messages, ok := objectMap["messages"].([]any)
	if !ok || len(messages) == 0 {
		return ErrNoMessages
	}

	messageList, ok := messages[0].([]any)
	if !ok || len(messageList) == 0 {
		return ErrNoMessages
	}

	parseAllMessagesInMessageList(messageList, tracker)

	return nil
}

func parseAllMessagesInMessageList(messageList []any, tracker *models.BaseTracker) {
	for _, rawMsg := range messageList {
		msgMap, ok := rawMsg.(map[string]any)
		if !ok {
			continue
		}

		msgType, ok1 := msgMap["type"].(string)
		value, ok2 := msgMap["measurementValue"].(float64)
		if !ok1 || !ok2 {
			continue
		}

		switch msgType {
		case "Battery":
			tracker.Battery = int16(value)
		case "Latitude":
			tracker.Position.Latitude = value

		case "Longitude":
			tracker.Position.Longitude = value
		}
	}
}
