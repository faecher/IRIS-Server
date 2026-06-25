// SPDX-License-Identifier: EUPL-1.2

package models

import "github.com/gofrs/uuid/v5"

// MCPConfig represents the MCP integration configuration stored in the database
type MCPConfig struct {
	SentNotEnabledWarning bool   // internal flag to track if the "not enabled" warning has been logged
	Enabled               bool   `db:"enabled" json:"enabled"`
	URL                   string `db:"url"     json:"url"`
	APIKey                string `db:"api_key" json:"api_key"`

	OperationID *uuid.UUID `db:"operation_id" json:"operation_id"`
	SiteplanID  *uuid.UUID `db:"siteplan_id"  json:"siteplan_id"`
}

// MCPOperation represents an operation in the MCP system
type MCPOperation struct {
	ID       uuid.UUID `json:"id,omitempty"`
	Name     string    `json:"title,omitempty"`
	Active   bool      `json:"active,omitempty"`
	Archived bool      `json:"archived,omitempty"`
	Place    *MCPPlace `json:"place,omitempty"`
}

// MCPSiteplan represents a siteplan in the MCP system
type MCPSiteplan struct {
	ID   uuid.UUID `json:"id,omitempty"`
	Name string    `json:"name,omitempty"`
}

// MCPPlace represents a geographic location in the MCP system
type MCPPlace struct {
	ID   uuid.UUID `json:"id,omitempty"`
	Name string    `json:"name,omitempty"`
}
