package chirpstack

import (
	"IRIS-Server/internal/models"
	"errors"

	"github.com/chirpstack/chirpstack/api/go/v4/integration"
)

var (
	ErrNoMessages = errors.New("no messages in uplink event")
)

func ParseChirpstackTrackerMessage(msg *integration.UplinkEvent, tracker *models.BaseTracker) error {
	objectMap := msg.Object.AsMap()

	messages, ok := objectMap["messages"].([]interface{})
	if !ok || len(messages) <= 0 {
		return ErrNoMessages
	}

	messageList, ok := messages[0].([]interface{})
	if !ok || len(messageList) <= 0 {
		return ErrNoMessages
	}

	for _, rawMsg := range messageList {
		msgMap, ok := rawMsg.(map[string]interface{})
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

	return nil
}
