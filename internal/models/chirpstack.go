package models

// ChirpstackTracker represents a tracker device using Chirpstack LoRaWAN
type ChirpstackTracker struct {
	BaseTracker
	DevEUI string `json:"deviceEUI" db:"dev_eui"`
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
