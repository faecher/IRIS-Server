package traccar

import "time"

type traccarMessage struct {
	Devices   []device   `json:"devices,omitempty"`
	Positions []position `json:"positions,omitempty"`
	Events    []event    `json:"events,omitempty"`
}

// commented out fields are not needed for our use case, but can be added later if needed
type device struct {
	ID         int64      `json:"id"` // Traccar device ID
	Name       string     `json:"name"`
	UniqueID   string     `json:"uniqueId"` // HW or protocol-based unique identifier
	Status     string     `json:"status"`   // online|offline|unknown
	LastUpdate *time.Time `json:"lastUpdate,omitempty"`
	PositionID *int64     `json:"positionId,omitempty"`
	// GroupID    *int64         `json:"groupId,omitempty"`
	// Phone      *string        `json:"phone,omitempty"`
	// Model      *string        `json:"model,omitempty"`
	// Contact    *string        `json:"contact,omitempty"`
	// Category   *string        `json:"category,omitempty"`
	// Attributes map[string]any `json:"attributes,omitempty"`
	Disabled bool `json:"disabled"`
}

type position struct {
	ID         int64     `json:"id"`
	DeviceID   int64     `json:"deviceId"`
	Protocol   string    `json:"protocol"`
	DeviceTime time.Time `json:"deviceTime"`
	// FixTime     time.Time      `json:"fixTime"`
	ServerTime time.Time `json:"serverTime"`
	Valid      bool      `json:"valid"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Altitude   float64   `json:"altitude"` // meters
	// Speed      float64   `json:"speed"`    // knots
	// Course     float64   `json:"course"`   // degrees
	// Address     string         `json:"address"`
	Accuracy float64 `json:"accuracy"` // meters
	// Network     map[string]any `json:"network,omitempty"`
	GeofenceIDs []int64 `json:"geofenceIds,omitempty"`
	// Attributes  map[string]any `json:"attributes,omitempty"`
}

type event struct {
	ID         int64     `json:"id"`
	Type       string    `json:"type"`
	EventTime  time.Time `json:"eventTime"`
	DeviceID   int64     `json:"deviceId"`
	PositionID int64     `json:"positionId"`
	GeofenceID int64     `json:"geofenceId"`
	// MaintenanceID int64          `json:"maintenanceId"`
	// Attributes    map[string]any `json:"attributes,omitempty"`
}
