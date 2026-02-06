// SPDX-License-Identifier: EUPL-1.2

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
	Longitude float64 `db:"position_longitude" json:"lon"`
	Latitude  float64 `db:"position_latitude"  json:"lat"`
}

// BaseTracker contains common fields for all tracker types
type BaseTracker struct {
	Position `json:"position"`

	TableauResource *TableauResource `json:"resource"`

	ID         uuid.UUID `db:"tracker_id" json:"id"`
	Name       string    `db:"name"       json:"name"`
	Battery    int16     `db:"battery"    json:"battery"`
	LastUpdate time.Time `db:"updated_at" json:"lastUpdate"`
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
