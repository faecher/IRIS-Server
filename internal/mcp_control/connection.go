package mcp_control

import (
	"errors"
	"net/http"
)

func TestMCPConnection() error {
	resp, err := mcpRequest("GET", "/api/version", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to connect to MCP: " + resp.Status)
	}

	return nil
}
