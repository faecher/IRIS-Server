// Package config provides functionality to load and manage application configuration from environment variables.
package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

// MCPConfig holds the configuration for MCP integration
type MCPConfig struct {
	ServerURL    string `env:"MCP_SERVER_URL"`
	APIToken     string `env:"MCP_API_TOKEN"`
	OperationUID string `env:"MCP_OPERATION_UID"`
}

// SQLConfig holds the database configuration
type SQLConfig struct {
	Host     string `env:"DB_HOST"     envDefault:"localhost"`
	Port     uint16 `env:"DB_PORT"     envDefault:"5432"`
	User     string `env:"DB_USER"     envDefault:"postgres"`
	Password string `env:"DB_PASSWORD"`
	DBName   string `env:"DB_NAME"     envDefault:"iris"`
	SSLMode  string `env:"DB_SSLMODE"  envDefault:"disable"`
}

// WebServerConfig holds the web server configuration
type WebServerConfig struct {
	Address string `env:"SERVER_ADDRESS" envDefault:"0.0.0.0:8080"`

	ReadTimeout    int `env:"SERVER_READ_TIMEOUT"     envDefault:"10"`      // in seconds
	WriteTimeout   int `env:"SERVER_WRITE_TIMEOUT"    envDefault:"10"`      // in seconds
	IdleTimeout    int `env:"SERVER_IDLE_TIMEOUT"     envDefault:"2"`       // in minutes
	MaxHeaderBytes int `env:"SERVER_MAX_HEADER_BYTES" envDefault:"1048576"` // 1 MB default
}

// SentryConfig holds the configuration for Sentry error tracking
// type SentryConfig struct {
// 	DSN         string `env:"SENTRY_DSN"`
// 	Environment string `env:"SENTRY_ENVIRONMENT" envDefault:"production"`
// }

// Config is the main configuration struct
type Config struct {
	MCP    MCPConfig
	SQL    SQLConfig
	Server WebServerConfig
	// Sentry SentryConfig
}

// Load reads environment variables and populates the Config struct
func Load() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}
	return cfg, nil
}

// ConnectionString returns a PostgreSQL connection string from the SQL configuration
func (c *SQLConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}
