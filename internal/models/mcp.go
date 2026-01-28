package models

import "github.com/gofrs/uuid/v5"

type MCPConfig struct {
	Enabled bool   `json:"enabled" db:"enabled"`
	URL     string `json:"url" db:"url"`
	APIKey  string `json:"api_key" db:"api_key"`

	OperationID *uuid.UUID `json:"operation_id" db:"operation_id"`
	SiteplanID  *uuid.UUID `json:"siteplan_id" db:"siteplan_id"`
}

type MCPOperation struct {
	ID       uuid.UUID `json:"id,omitempty"`
	Name     string    `json:"title,omitempty"`
	Active   bool      `json:"active,omitempty"`
	Archived bool      `json:"archived,omitempty"`
	Place    *MCPPlace `json:"place,omitempty"`
}

type MCPSiteplan struct {
	ID   uuid.UUID `json:"id,omitempty"`
	Name string    `json:"name,omitempty"`
}

type MCPPlace struct {
	ID   uuid.UUID `json:"id,omitempty"`
	Name string    `json:"name,omitempty"`
}
