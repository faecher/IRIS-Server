// SPDX-License-Identifier: EUPL-1.2

package mcpcontrol

import (
	"IRIS-Server/internal/models"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// ErrMCPConnectionFailed indicates that the connection test to the MCP system failed
var ErrMCPConnectionFailed = errors.New("MCP connection failed")

// TestMCPConnection checks connectivity to the MCP system by requesting its version endpoint.
func TestMCPConnection(newConfig models.MCPConfig) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, newConfig.URL+"/api/version", nil)
	if err != nil {
		return fmt.Errorf("failed to create MCP request: %w", err)
	}

	req.Header.Set("Api-Key", newConfig.APIKey)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")

	resp, err := mcpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform MCP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("MCP connection test failed", "status", resp.Status)
		return fmt.Errorf("%w: %s", ErrMCPConnectionFailed, resp.Status)
	}

	return nil
}
