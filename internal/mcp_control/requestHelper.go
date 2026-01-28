package mcp_control

import (
	"IRIS-Server/internal/models"
	"errors"
	"io"
	"net/http"
	"time"
)

var (
	mcpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	MCPConfig models.MCPConfig

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

	req, err := http.NewRequest(method, MCPConfig.URL+endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Api-Key", MCPConfig.APIKey)
	req.Header.Set("accept", "*/*")

	req.Body = body

	return mcpClient.Do(req)
}
