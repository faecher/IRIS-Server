// Package mcpcontrol provides functions to interact with the MCP (Medical Care Platform) system.
// SPDX-License-Identifier: EUPL-1.2
package mcpcontrol

import (
	"IRIS-Server/internal/config"
	"IRIS-Server/internal/models"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	mcpClient = &http.Client{
		Transport: &http.Transport{},
	}

	// MCPConfig holds the configuration for MCP integration
	MCPConfig models.MCPConfig

	// ErrMCPDisabled indicates that MCP integration is disabled
	ErrMCPDisabled = errors.New("MCP integration is disabled")
)

// InitMCPClient initializes the MCP HTTP client with the given configuration.
func InitMCPClient(config config.MCPConfig) {
	clientTransport, ok := mcpClient.Transport.(*http.Transport)
	if ok {
		clientTransport.TLSClientConfig.InsecureSkipVerify = !config.EnableSSLVerification
	}

	mcpClient.Timeout = time.Duration(config.RequestTimeout) * time.Second
}

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

	resp, err := mcpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform MCP request: %w", err)
	}

	return resp, nil
}
