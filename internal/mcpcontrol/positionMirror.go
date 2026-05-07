// SPDX-License-Identifier: EUPL-1.2

package mcpcontrol

import (
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gofrs/uuid/v5"
)

// ErrMCPRequestFailed indicates that an HTTP request to the MCP server failed
var ErrMCPRequestFailed = errors.New("MCP request failed")

// UpdateMarkerInMCP sends the updated tracker information to the MCP system.
// This is done by identifying the resource the tracker belongs to and updating its position marker in MCP.
// If no resource is found, no action is taken.
func UpdateMarkerInMCP(trackerID uuid.UUID) error {
	tracker, err := repository.GetTrackerByID(trackerID)
	if err != nil {
		return fmt.Errorf("failed to get tracker: %w", err)
	}
	if tracker.TableauResource == nil {
		// No resource assigned, nothing to update
		return nil
	}

	marker, err := repository.GetResourceMarker(tracker.TableauResource.ID)
	if err != nil {
		return fmt.Errorf("failed to get resource marker: %w", err)
	}

	requestBody := map[string]any{
		"name": tracker.TableauResource.Resource.Name,
		"position": map[string]any{
			"lat": tracker.Position.Latitude,
			"lng": tracker.Position.Longitude,
		},
		"id":         marker.MarkerID,
		"icon":       "BASIC_PIN",
		"entityType": "TEMPLATE",
		"siteplanId": marker.SiteplanID,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	body := io.NopCloser(bytes.NewReader(jsonData))

	if marker.MarkerID == uuid.Nil { // Create new marker
		err := createNewMarkerInMCP(body, tracker)
		if err != nil {
			return fmt.Errorf("failed to create new marker in MCP: %w", err)
		}
	} else { // Update existing marker
		resp, err := mcpRequest(http.MethodPut, "/api/markers", body)
		if err != nil {
			return fmt.Errorf("failed to update marker: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return ErrMCPRequestFailed
		}
	}

	return nil
}

func createNewMarkerInMCP(body io.ReadCloser, tracker *models.BaseTracker) error {
	resp, err := mcpRequest(http.MethodPost, "/api/markers", body)
	if err != nil {
		return fmt.Errorf("failed to create marker: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("MCP marker creation failed", "status", resp.Status, "respBody", string(respBody))
		return ErrMCPRequestFailed
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read marker creation response: %w", err)
	}

	var respData map[string]any
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal marker creation response: %w", err)
	}

	respID, ok := respData["id"].(string)
	if !ok {
		return ErrMCPRequestFailed
	}

	markerID, err := uuid.FromString(respID)
	if err != nil {
		return fmt.Errorf("failed to parse marker ID: %w", err)
	}

	err = repository.UpdateMarkerIDForResource(tracker.TableauResource.Resource.ID, markerID)
	if err != nil {
		return fmt.Errorf("failed to update marker ID for resource: %w", err)
	}

	return nil
}
