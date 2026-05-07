// SPDX-License-Identifier: EUPL-1.2

package repository

import (
	"IRIS-Server/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

// GetMCPConfig retrieves the MCP integration configuration. Returns empty config if none exists.
func GetMCPConfig() (models.MCPConfig, error) {
	var config models.MCPConfig

	SQL := `SELECT url, api_key, enabled, operation_id, siteplan_id FROM mcp_config WHERE id = 1`

	err := DBConnPool.QueryRow(context.Background(), SQL).Scan(
		&config.URL,
		&config.APIKey,
		&config.Enabled,
		&config.OperationID,
		&config.SiteplanID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.MCPConfig{}, nil // No config found
	}
	if err != nil {
		return models.MCPConfig{}, fmt.Errorf("failed to query MCP config: %w", err)
	}

	return config, nil
}

// UpdateMCPConfig replaces the entire MCP configuration (singleton)
func UpdateMCPConfig(newConfig models.MCPConfig) error {
	// Treat as singleton Config
	SQL := `INSERT INTO mcp_config (id, url, api_key, enabled) VALUES (1, $1, $2, $3)
	        ON CONFLICT (id) DO UPDATE SET url = $1, api_key = $2, enabled = $3`

	_, err := DBConnPool.Exec(context.Background(), SQL, newConfig.URL, newConfig.APIKey, newConfig.Enabled)
	if err != nil {
		return fmt.Errorf("failed to insert new MCP config: %w", err)
	}

	return nil
}

// UpdateMCPOperation sets the selected MCP operation ID
func UpdateMCPOperation(operationID uuid.UUID) error {
	// Ensure mcp_config row exists, then update operation_id
	SQL := `INSERT INTO mcp_config (id, url, api_key, enabled, operation_id) 
	        VALUES (1, '', '', false, $1)
	        ON CONFLICT (id) DO UPDATE SET operation_id = $1`

	_, err := DBConnPool.Exec(context.Background(), SQL, operationID)
	if err != nil {
		return fmt.Errorf("failed to update MCP operation: %w", err)
	}

	return nil
}

// GetMCPOperation retrieves the currently selected MCP operation ID
func GetMCPOperation() (*uuid.UUID, error) {
	SQL := `SELECT operation_id FROM mcp_config WHERE id = 1`

	var operationID *uuid.UUID
	err := DBConnPool.QueryRow(context.Background(), SQL).Scan(&operationID)
	if err != nil {
		return nil, fmt.Errorf("failed to query MCP operation: %w", err)
	}

	return operationID, nil
}

// UpdateMCPSiteplan sets the selected MCP siteplan ID
func UpdateMCPSiteplan(siteplanID uuid.UUID) error {
	// Ensure mcp_config row exists, then update siteplan_id
	SQL := `INSERT INTO mcp_config (id, url, api_key, enabled, siteplan_id) 
	        VALUES (1, '', '', false, $1)
	        ON CONFLICT (id) DO UPDATE SET siteplan_id = $1`
	_, err := DBConnPool.Exec(context.Background(), SQL, siteplanID)
	if err != nil {
		return fmt.Errorf("failed to update MCP siteplan: %w", err)
	}

	return nil
}

// GetMCPSiteplan retrieves the currently selected MCP siteplan ID
func GetMCPSiteplan() (*uuid.UUID, error) {
	SQL := `SELECT siteplan_id FROM mcp_config WHERE id = 1`

	var siteplanID *uuid.UUID
	err := DBConnPool.QueryRow(context.Background(), SQL).Scan(&siteplanID)
	if err != nil {
		return nil, fmt.Errorf("failed to query MCP siteplan: %w", err)
	}

	return siteplanID, nil
}
