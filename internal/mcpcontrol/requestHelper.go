// Package mcpcontrol provides functions to interact with the MCP (Medical Care Platform) system.
package mcpcontrol

import (
	"IRIS-Server/internal/models"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	mcpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // WARNING: Disables certificate verification (TODO: make configurable)
			},
		},
	}

	// MCPConfig holds the configuration for MCP integration
	MCPConfig models.MCPConfig

	// ErrMCPDisabled indicates that MCP integration is disabled
	ErrMCPDisabled = errors.New("MCP integration is disabled")
)

// mcpRequest performs an HTTP request to the MCP server
// with the given method, endpoint, and body.
// It returns the HTTP response or an error.
// BaseURL and APIKey are read from the global MCPConfig variable.
func mcpRequest(method, endpoint string, body io.ReadCloser) (*http.Response, error) {
	if !MCPConfig.Enabled {
		return nil, ErrMCPDisabled
	}

	req, err := http.NewRequestWithContext(context.Background(), method, MCPConfig.URL+endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP request: %w", err)
	}

	req.Header.Set("Api-Key", MCPConfig.APIKey)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")

	req.Body = body

	resp, err := mcpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform MCP request: %w", err)
	}

	return resp, nil
}
