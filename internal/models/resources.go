// Package models defines data models used throughout the application
package models

import "github.com/gofrs/uuid/v5"

// Resource represents an MCP resource that can be assigned to trackers
type Resource struct {
	ID     uuid.UUID `db:"resource_id" json:"id"`
	Name   string    `db:"name"        json:"name"`
	Type   string    `db:"type"        json:"type"`
	Status uint16    `db:"status"      json:"status"`
}

// ResourceMarker associates a resource with an MCP marker on a specific siteplan
type ResourceMarker struct {
	ResourceID uuid.UUID `db:"resource_id" json:"resourceId"`
	MarkerID   uuid.UUID `db:"marker_id"   json:"markerId"`
	SiteplanID uuid.UUID `db:"siteplan_id" json:"siteplanId"`
}
