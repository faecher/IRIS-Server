package models

import "github.com/gofrs/uuid/v5"

// Resource represents an MCP resource that can be assigned to trackers
type Resource struct {
	ID       uuid.UUID  `json:"id" db:"resource_id"`
	MarkerID *uuid.UUID `json:"markerId" db:"marker_id"`
	Name     string     `json:"name" db:"name"`
	Type     string     `json:"type" db:"type"`
	Status   uint16     `json:"status" db:"status"`
}
