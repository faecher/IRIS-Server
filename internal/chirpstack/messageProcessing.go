// SPDX-License-Identifier: EUPL-1.2

// Package chirpstack provides utilities for processing messages from Chirpstack
package chirpstack

import (
	"IRIS-Server/internal/models"
	"errors"
	"time"
)

var (
	// ErrNoMessages indicates that there are no messages in the uplink event
	ErrNoMessages = errors.New("no messages in uplink event")
)

// ParseChirpstackTrackerMessage extracts battery, position data, and timestamp from a Chirpstack uplink event
func ParseChirpstackTrackerMessage(msg *models.ChirpstackUplink, tracker *models.BaseTracker) error {
	tracker.Battery = int16(msg.Object.Battery)
	tracker.Position.Latitude = msg.Object.Latitude
	tracker.Position.Longitude = msg.Object.Longitude
	tracker.Name = msg.DeviceInfo.DeviceName

	// Extract timestamp from uplink event
	if msg.Object.Timestamp != 0 {
		tracker.LastUpdate = time.Unix(int64(msg.Object.Timestamp), 0).UTC()
	}
	return nil
}
