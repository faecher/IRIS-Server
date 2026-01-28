package repository

import (
	"IRIS-Server/internal/models"
	"context"

	"github.com/jackc/pgx/v5"
)

func GetMCPConfig() (models.MCPConfig, error) {
	var config models.MCPConfig

	SQL := `SELECT url, api_key FROM mcp_config`

	err := DBConnPool.QueryRow(context.Background(), SQL).Scan(
		&config.URL,
		&config.APIKey,
	)
	if err == pgx.ErrNoRows {
		return models.MCPConfig{}, nil // No config found
	}
	if err != nil {
		return models.MCPConfig{}, err
	}

	return config, nil
}

func UpdateMCPConfig(newConfig models.MCPConfig) error {
	SQL := `DELETE * FROM mcp_config;
			INSERT INTO mcp_config (url, api_key, enabled) VALUES ($1, $2, $3)`

	_, err := DBConnPool.Exec(context.Background(), SQL, newConfig.URL, newConfig.APIKey, newConfig.Enabled)

	return err
}
