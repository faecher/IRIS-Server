package models

type MCPConfig struct {
	Enabled bool   `json:"enabled" db:"enabled"`
	URL     string `json:"url" db:"url"`
	APIKey  string `json:"api_key" db:"api_key"`
}
