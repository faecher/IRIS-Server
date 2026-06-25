// SPDX-License-Identifier: EUPL-1.2

package models

// ChirpstackTracker represents a tracker device using Chirpstack LoRaWAN
type ChirpstackTracker struct {
	BaseTracker

	DevEUI string `db:"dev_eui" json:"deviceEUI"`
}

// ChirpstackUplink represents the structure of an uplink event received from Chirpstack
type ChirpstackUplink struct {
	DeviceInfo ChirpstackMessage `json:"deviceInfo"`
	Object     ChirpstackData    `json:"object"`
}

// ChirpstackMessage represents the device information in a Chirpstack uplink event
type ChirpstackMessage struct {
	DeviceName string `json:"deviceName"`
	DevEui     string `json:"devEui"`
}

// ChirpstackData represents the data payload in a Chirpstack uplink event
type ChirpstackData struct {
	Timestamp float64 `json:"timestamp"`
	Battery   float64 `json:"battery"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Invalid   bool    `json:"unsupported,omitempty"` // default false, true if the message is unsupported
}

// NewChirpstackTracker creates a new Chirpstack tracker instance
func NewChirpstackTracker(devEUI, name string) *ChirpstackTracker {
	return &ChirpstackTracker{
		BaseTracker: BaseTracker{
			Name: name,
		},
		DevEUI: devEUI,
	}
}
