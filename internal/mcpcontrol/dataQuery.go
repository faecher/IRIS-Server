// SPDX-License-Identifier: EUPL-1.2

package mcpcontrol

import (
	"IRIS-Server/internal/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ErrNoOperationSelected indicates that no MCP operation has been configured
var ErrNoOperationSelected = errors.New("no MCP operation selected")

// ErrNoPlaceAssociated indicates that the configured MCP operation has no associated place
var ErrNoPlaceAssociated = errors.New("no place associated with MCP operation")

// GetMCPOperations fetches all available operations from the MCP API
func GetMCPOperations() ([]models.MCPOperation, error) {
	resp, err := mcpRequest(context.Background(), http.MethodGet, "/api/operations", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request MCP operations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrMCPRequestFailed
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Response Body: %w", err)
	}

	var operations []models.MCPOperation
	err = json.Unmarshal(body, &operations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal MCP operations: %w", err)
	}

	return operations, nil
}

// GetMCPSiteplans fetches all siteplans for the currently selected operation
// Throws ErrNoOperationSelected if no operation is selected
// Throws ErrNoPlaceAssociated if the selected operation has no associated place
// Throws ErrMCPRequestFailed if the MCP request fails
func GetMCPSiteplans() ([]models.MCPSiteplan, error) {
	body, err := mcpRequestFromEndpointWithCurrentOperation(http.MethodGet, "/api/siteplan/snapshot")
	if err != nil {
		return nil, fmt.Errorf("failed to request MCP siteplans: %w", err)
	}

	var siteplans []models.MCPSiteplan
	err = json.Unmarshal(body, &siteplans)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal MCP siteplans: %w", err)
	}

	return siteplans, nil
}

func getMCPResources() ([]models.TableauResource, error) {
	body, err := mcpRequestFromEndpointWithCurrentOperation(http.MethodGet, "/api/tableau/resources")
	if err != nil {
		return nil, fmt.Errorf("failed to request MCP resources: %w", err)
	}

	var resources []models.TableauResource
	err = json.Unmarshal(body, &resources)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal MCP resources: %w", err)
	}

	return resources, nil
}
