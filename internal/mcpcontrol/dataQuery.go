// SPDX-License-Identifier: EUPL-1.2

package mcpcontrol

import (
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

// ErrNoOperationSelected indicates that no MCP operation has been configured
var ErrNoOperationSelected = errors.New("no MCP operation selected")

// ErrNoPlaceAssociated indicates that the configured MCP operation has no associated place
var ErrNoPlaceAssociated = errors.New("no place associated with MCP operation")

// GetMCPOperations fetches all available operations from the MCP API
func GetMCPOperations() ([]models.MCPOperation, error) {
	resp, err := mcpRequest(http.MethodGet, "/api/operations", nil)
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
	// get current operation from DB
	operationID, err := repository.GetMCPOperation()
	if errors.Is(err, pgx.ErrNoRows) || operationID == nil {
		return nil, ErrNoOperationSelected
	} else if err != nil {
		return nil, fmt.Errorf("failed to get MCP operation: %w", err)
	}

	placeID, err := getMCPPlaceFromOperation(*operationID)
	if errors.Is(err, ErrNoPlaceAssociated) {
		return nil, ErrNoPlaceAssociated
	} else if err != nil {
		return nil, fmt.Errorf("failed to get MCP place: %w", err)
	}

	resp, err := mcpRequest(http.MethodGet, "/api/siteplan/template?placeId="+placeID.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request MCP siteplans: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrMCPRequestFailed
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read MCP siteplans response: %w", err)
	}

	var siteplans []models.MCPSiteplan
	err = json.Unmarshal(body, &siteplans)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal MCP siteplans: %w", err)
	}

	return siteplans, nil
}

func getMCPPlaceFromOperation(operationID uuid.UUID) (uuid.UUID, error) {
	resp, err := mcpRequest(http.MethodGet, "/api/operations/"+operationID.String(), nil)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to request MCP organization: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("MCP organization request failed", "status", resp.Status, "resp", resp.Body)
		return uuid.Nil, ErrMCPRequestFailed
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to read MCP organization response: %w", err)
	}

	var operation models.MCPOperation
	err = json.Unmarshal(body, &operation)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal MCP organization: %w", err)
	}

	if operation.Place == nil {
		return uuid.Nil, ErrNoPlaceAssociated
	}

	return operation.Place.ID, nil
}

func getMCPResources() ([]models.Resource, error) {
	resp, err := mcpRequest(http.MethodGet, "/api/resources", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request MCP resources: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrMCPRequestFailed
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read MCP resources response: %w", err)
	}

	var resources []models.Resource
	err = json.Unmarshal(body, &resources)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal MCP resources: %w", err)
	}

	return resources, nil
}
