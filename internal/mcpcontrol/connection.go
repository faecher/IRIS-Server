package mcpcontrol

import (
	"fmt"
	"net/http"
)

// ErrMCPConnectionFailed indicates that the connection test to the MCP system failed
var ErrMCPConnectionFailed = fmt.Errorf("MCP connection failed")

// TestMCPConnection checks connectivity to the MCP system by requesting its version endpoint.
func TestMCPConnection() error {
	resp, err := mcpRequest("GET", "/api/version", nil)
	if err != nil {
		return fmt.Errorf("failed to request MCP version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: "+resp.Status, ErrMCPConnectionFailed)
	}

	return nil
}
