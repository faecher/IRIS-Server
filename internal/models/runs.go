// SPDX-License-Identifier: EUPL-1.2

package models

import "github.com/gofrs/uuid/v5"

// Run represents an MCP run that is stored locally and returned by the API.
type Run struct {
	OperationID uuid.UUID    `db:"operation_id"     json:"-"`
	Operation   MCPOperation `json:"operation"`

	ID            uuid.UUID `db:"run_id"              json:"id"`
	House         string    `db:"house_object"       json:"obj,omitempty"`
	City          string    `db:"place"              json:"place,omitempty"`
	Street        string    `db:"address_field"      json:"address,omitempty"`
	Latitude      float64   `db:"position_latitude"  json:"lat"`
	Longitude     float64   `db:"position_longitude" json:"long"`
	UnsetPosition bool      `db:"unset_position"     json:"unsetPosition"`
	HasPatient    bool      `db:"has_patient"        json:"hasPatient"`
	Active        bool      `db:"active"             json:"active"`
	Nr            int       `db:"nr"                 json:"nr"`
	Text          string    `db:"notes"              json:"text,omitempty"`
	Deleted       bool      `db:"deleted"            json:"deleted"`
}
