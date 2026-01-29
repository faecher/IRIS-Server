package repository

import (
	"IRIS-Server/internal/models"
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func GetMCPConfig() (models.MCPConfig, error) {
	var config models.MCPConfig

	SQL := `SELECT url, api_key, enabled, operation_id, siteplan_id FROM mcp_config`

	err := DBConnPool.QueryRow(context.Background(), SQL).Scan(
		&config.URL,
		&config.APIKey,
		&config.Enabled,
		&config.OperationID,
		&config.SiteplanID,
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
	// Treat as singleton Config
	SQL1 := `DELETE FROM mcp_config;`
	SQL2 := `INSERT INTO mcp_config (id, url, api_key, enabled) VALUES (1, $1, $2, $3)`

	_, err := DBConnPool.Exec(context.Background(), SQL1)
	if err != nil {
		return err
	}

	_, err = DBConnPool.Exec(context.Background(), SQL2, newConfig.URL, newConfig.APIKey, newConfig.Enabled)

	return err
}

func UpdateMCPOperation(operationID uuid.UUID) error {
	// Ensure mcp_config row exists, then update operation_id
	SQL := `INSERT INTO mcp_config (id, url, api_key, enabled, operation_id) 
	        VALUES (1, '', '', false, $1)
	        ON CONFLICT (id) DO UPDATE SET operation_id = $1`

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
	// Ensure mcp_config row exists, then update siteplan_id
	SQL := `INSERT INTO mcp_config (id, url, api_key, enabled, siteplan_id) 
	        VALUES (1, '', '', false, $1)
	        ON CONFLICT (id) DO UPDATE SET siteplan_id = $1`
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
