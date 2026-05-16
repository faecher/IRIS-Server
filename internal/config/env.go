// SPDX-License-Identifier: EUPL-1.2

// Package config provides functionality to load and manage application configuration from environment variables.
package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

// UpdateConfig holds the configuration for periodic updates
type UpdateConfig struct {
	ResourceUpdate uint16 `env:"MCP_RESOURCE_UPDATE" envDefault:"5"` // in seconds
}

// MCPConfig holds the MCP integration configuration
type MCPConfig struct {
	EnableSSLVerification bool `env:"MCP_ENABLE_SSL_VERIFICATION" envDefault:"true"`
	RequestTimeout        int  `env:"MCP_TIMEOUT"                 envDefault:"10"` // in seconds
}

// SQLConfig holds the database configuration
type SQLConfig struct {
	Host             string `env:"DB_HOST"                    envDefault:"localhost"`
	Port             uint16 `env:"DB_PORT"                    envDefault:"5432"`
	User             string `env:"DB_USER"                    envDefault:"postgres"`
	Password         string `env:"DB_PASSWORD"` // will get populated from file if unset and DB_PASSWORD_FROM_FILE is set
	PasswordFromFile string `env:"DB_PASSWORD_FROM_FILE,file"`
	DBName           string `env:"DB_NAME"                    envDefault:"iris"`
	SSLMode          string `env:"DB_SSLMODE"                 envDefault:"disable"`
}

// WebServerConfig holds the web server configuration
type WebServerConfig struct {
	Address string `env:"SERVER_ADDRESS" envDefault:"0.0.0.0"` // port is statically set to 8080 by the server startup code

	ReadTimeout    int `env:"SERVER_READ_TIMEOUT"     envDefault:"10"`      // in seconds
	WriteTimeout   int `env:"SERVER_WRITE_TIMEOUT"    envDefault:"10"`      // in seconds
	IdleTimeout    int `env:"SERVER_IDLE_TIMEOUT"     envDefault:"2"`       // in minutes
	MaxHeaderBytes int `env:"SERVER_MAX_HEADER_BYTES" envDefault:"1048576"` // 1 MB default
}

// Config is the main configuration struct
type Config struct {
	Update UpdateConfig
	MCP    MCPConfig
	SQL    SQLConfig
	Server WebServerConfig
}

// Load reads environment variables and populates the Config struct
func Load() (*Config, error) {
	// Load .env file if it exists (silently ignore if not found)
	_ = godotenv.Load()

	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	if cfg.SQL.Password == "" && cfg.SQL.PasswordFromFile != "" {
		cfg.SQL.Password = cfg.SQL.PasswordFromFile
	}
	return cfg, nil
}

// ConnectionString returns a PostgreSQL connection string from the SQL configuration
func (c *SQLConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}
