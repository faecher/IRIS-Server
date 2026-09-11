// SPDX-License-Identifier: EUPL-1.2

package mcpcontrol

import (
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrLocationNotFound indicates that the location for a given run could not be found
var ErrLocationNotFound = errors.New("location not found")

var positionRequestTimeout = 5 * time.Second

// UpdateMCPResourcesInDB fetches resources from the MCP system and updates/inserts them into the local database
func UpdateMCPResourcesInDB() error {
	resources, err := getMCPResources()
	if err != nil {
		return fmt.Errorf("failed to get MCP resources: %w", err)
	}

	for _, res := range resources {
		err := repository.UpsertResource(&res)
		if err != nil {
			return fmt.Errorf("failed to upsert resource %s: %w", res.ID.String(), err)
		}
	}

	return nil
}

// UpdateMCPRunsInDB fetches runs from the MCP system and updates/inserts them into the local database
func UpdateMCPRunsInDB() error {
	runs, err := getMCPRuns()
	if err != nil {
		return fmt.Errorf("failed to get MCP runs: %w", err)
	}

	for _, run := range runs {
		oldRun, err := repository.GetRunByID(run.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get run by ID %s: %w", run.ID.String(), err)
		}

		if oldRun == nil || oldRun.Street != run.Street || oldRun.House != run.House || oldRun.City != run.City {
			// get coordinates for the given address
			latitude, longitude, err := getCoordinates(run)
			if err != nil {
				slog.Error("Failed to get coordinates for run", "runID", run.ID.String(), "error", err, "address", run.Street+" "+run.House)
			}

			run.Latitude = latitude
			run.Longitude = longitude
		} else {
			run.Latitude = oldRun.Latitude
			run.Longitude = oldRun.Longitude
		}

		err = repository.UpsertRun(&run)
		if err != nil {
			return fmt.Errorf("failed to upsert runs %s: %w", run.ID.String(), err)
		}
	}

	return nil
}

func getCoordinates(run models.MCPRun) (float64, float64, error) {
	// access nominatim.org api, see https://nominatim.org/release-docs/develop/api/Search/ for more information
	query := fmt.Sprintf("https://nominatim.openstreetmap.org/search?street=%s&city=%s&format=json&limit=1",
		run.Street+" "+run.House, run.City)

	ctx, cancel := context.WithTimeout(context.Background(), positionRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, query, nil) // #nosec G107
	if err != nil {
		return math.NaN(), math.NaN(), fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req) // #nosec G107
	if err != nil {
		return math.NaN(), math.NaN(), fmt.Errorf("failed to get coordinates: %w", err)
	}
	defer resp.Body.Close()

	// Parse the JSON response
	var results []struct {
		Lat float64 `json:"lat,string"`
		Lon float64 `json:"lon,string"`
	}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return math.NaN(), math.NaN(), fmt.Errorf("failed to decode coordinates: %w", err)
	}

	if len(results) == 0 {
		return math.NaN(), math.NaN(), ErrLocationNotFound
	}

	return results[0].Lat, results[0].Lon, nil
}
