package repository

import (
	"IRIS-Server/internal/models"
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func GetMCPConfig() (models.MCPConfig, error) {
	var config models.MCPConfig

	SQL := `SELECT url, api_key, enabled FROM mcp_config`

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
	// Use DELETE FROM (not DELETE *) and handle as singleton config
	SQL := `DELETE FROM mcp_config;
			INSERT INTO mcp_config (id, url, api_key, enabled) VALUES (1, $1, $2, $3)`

	_, err := DBConnPool.Exec(context.Background(), SQL, newConfig.URL, newConfig.APIKey, newConfig.Enabled)

	return err
}

func UpdateMCPOperation(operationID uuid.UUID) error {
	// Assuming mcp_config is a singleton table with id=1
	SQL := `UPDATE mcp_config SET operation_id = $1 WHERE id = 1`

	_, err := DBConnPool.Exec(context.Background(), SQL, operationID)

	return err
}

func GetMCPOperation() (uuid.UUID, error) {
	SQL := `SELECT operation_id FROM mcp_config WHERE id = 1`

	var operationID uuid.UUID
	err := DBConnPool.QueryRow(context.Background(), SQL).Scan(&operationID)
	if err != nil {
		return uuid.Nil, err
	}

	return operationID, nil
}

func UpdateMCPSiteplan(siteplanID uuid.UUID) error {
	// Assuming mcp_config is a singleton table with id=1
	SQL := `UPDATE mcp_config SET siteplan_id = $1 WHERE id = 1`
	_, err := DBConnPool.Exec(context.Background(), SQL, siteplanID)

	return err
}

func GetMCPSiteplan() (uuid.UUID, error) {
	SQL := `SELECT siteplan_id FROM mcp_config WHERE id = 1`

	var siteplanID uuid.UUID
	err := DBConnPool.QueryRow(context.Background(), SQL).Scan(&siteplanID)
	if err != nil {
		return uuid.Nil, err
	}

	return siteplanID, nil
}
