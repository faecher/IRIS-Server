package mcp_control

import (
	"IRIS-Server/internal/repository"
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/gofrs/uuid/v5"
)

var ErrMCPRequestFailed = errors.New("MCP request failed")

// UpdateTrackerInMCP sends the updated tracker information to the MCP system.
// This is done by identifying the resource the tracker belongs to and updating its position marker in MCP.
// If no resource is found, no action is taken.
func UpdateTrackerInMCP(trackerID uuid.UUID) error {
	tracker, err := repository.GetTrackerByID(trackerID)
	if err != nil {
		return err
	}

	requestBody := map[string]any{
		"name": tracker.Resource.Name,
		"position": map[string]any{
			"lat": tracker.Position.Latitude,
			"lng": tracker.Position.Longitude,
		},
		"id":         tracker.Resource.MarkerID,
		"icon":       "BASIC_PIN", // TODO: what icon to use here? read from resource type?
		"entityType": "TEMPLATE",
		"siteplanId": nil, // TODO: where tf do we get siteplanId from? do we really need to make a ui page to select this?
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	body := io.NopCloser(bytes.NewReader(jsonData))

	if tracker.Resource.MarkerID == nil {
		resp, err := mcpRequest("POST", "/api/markers", body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return ErrMCPRequestFailed
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var respData map[string]any
		err = json.Unmarshal(body, &respData)
		if err != nil {
			return err
		}

		respID, ok := respData["id"].(string)
		if !ok {
			return ErrMCPRequestFailed
		}

		markerID, err := uuid.FromString(respID)
		if err != nil {
			return err
		}

		err = repository.UpdateMarkerIDForResource(tracker.Resource.ID, markerID)
		if err != nil {
			return err
		}
	} else {
		resp, err := mcpRequest("PUT", "/api/markers", body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return ErrMCPRequestFailed
		}
	}

	return nil
}
