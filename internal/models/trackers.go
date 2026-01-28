package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// Tracker is the interface that all tracker types must implement
type Tracker interface {
	GetPosition() Position
	GetLastUpdate() time.Time
	GetBattery() int16
}

// Position represents a GPS coordinate
type Position struct {
	Longitude float64 `json:"lon" db:"position_longitude"`
	Latitude  float64 `json:"lat" db:"position_latitude"`
}

// BaseTracker contains common fields for all tracker types
type BaseTracker struct {
	Position         `json:"position"`
	Resource         `json:"resource"`
	ID               uuid.UUID `json:"id" db:"tracker_id"`
	AssignedResource uuid.UUID `json:"resourceId" db:"resource_id"`
	Name             string    `json:"name" db:"name"`
	Battery          int16     `json:"battery" db:"battery"`
	LastUpdate       time.Time `json:"lastUpdate" db:"updated_at"`
}

// GetPosition returns the tracker's GPS position
func (b *BaseTracker) GetPosition() Position {
	return b.Position
}

// GetLastUpdate returns the timestamp of the last update
func (b *BaseTracker) GetLastUpdate() time.Time {
	return b.LastUpdate
}

// GetBattery returns the battery level
func (b *BaseTracker) GetBattery() int16 {
	return b.Battery
}
