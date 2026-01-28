package mcp_control

import (
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"encoding/json"
	"errors"
	"io"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

var ErrNoOperationSelected = errors.New("no MCP operation selected")
var ErrNoPlaceAssociated = errors.New("no place associated with MCP operation")

func GetMCPOperations() ([]models.MCPOperation, error) {
	resp, err := mcpRequest("GET", "/api/operations", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, ErrMCPRequestFailed
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var operations []models.MCPOperation
	err = json.Unmarshal(body, &operations)
	if err != nil {
		return nil, err
	}

	return operations, nil
}

func GetMCPSiteplans() ([]models.MCPSiteplan, error) {
	// get current operation from DB
	operationID, err := repository.GetMCPOperation()
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoOperationSelected
	} else if err != nil {
		return nil, err
	}

	placeID, err := getMCPPlaceFromOrg(operationID)
	if errors.Is(err, ErrNoPlaceAssociated) {
		return nil, ErrNoPlaceAssociated
	} else if err != nil {
		return nil, err
	}

	resp, err := mcpRequest("GET", "/api/siteplan/template?placeId="+placeID.String(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, ErrMCPRequestFailed
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var siteplans []models.MCPSiteplan
	err = json.Unmarshal(body, &siteplans)
	if err != nil {
		return nil, err
	}

	return siteplans, nil
}

func getMCPPlaceFromOrg(orgID uuid.UUID) (uuid.UUID, error) {
	resp, err := mcpRequest("GET", "/api/organizations/"+orgID.String(), nil)
	if err != nil {
		return uuid.Nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return uuid.Nil, ErrMCPRequestFailed
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return uuid.Nil, err
	}

	var operation models.MCPOperation
	err = json.Unmarshal(body, &operation)
	if err != nil {
		return uuid.Nil, err
	}

	if operation.Place == nil {
		return uuid.Nil, ErrNoPlaceAssociated
	}

	return operation.Place.ID, nil
}
