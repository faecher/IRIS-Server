package mcpcontrol

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// ErrMCPConnectionFailed indicates that the connection test to the MCP system failed
var ErrMCPConnectionFailed = errors.New("MCP connection failed")

// TestMCPConnection checks connectivity to the MCP system by requesting its version endpoint.
func TestMCPConnection() error {
	resp, err := mcpRequest("GET", "/api/version", nil)
	if err != nil {
		return fmt.Errorf("failed to request MCP version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("MCP connection test failed", "status", resp.Status)
		return fmt.Errorf("%w: %s", ErrMCPConnectionFailed, resp.Status)
	}

	return nil
}
