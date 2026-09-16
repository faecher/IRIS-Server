// SPDX-License-Identifier: EUPL-1.2

package mcpcontrol

import (
	"IRIS-Server/internal/config"
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
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
func UpdateMCPRunsInDB(cfg config.GeocodingConfig) error {
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
			latitude, longitude, err := getCoordinates(run, cfg)
			if err != nil {
				slog.Error("Failed to get coordinates for run", "runID", run.ID.String(), "error", err, "address", run.Street+" "+run.House)
				run.UnsetPosition = true
				run.Latitude = 0
				run.Longitude = 0
			} else {
				run.UnsetPosition = false
				run.Latitude = latitude
				run.Longitude = longitude
			}
		} else {
			run.Latitude = oldRun.Latitude
			run.Longitude = oldRun.Longitude
			run.UnsetPosition = oldRun.UnsetPosition
		}

		err = repository.UpsertRun(&run)
		if err != nil {
			return fmt.Errorf("failed to upsert runs %s: %w", run.ID.String(), err)
		}
	}

	return nil
}

func getCoordinates(run models.Run, cfg config.GeocodingConfig) (float64, float64, error) {
	// access nominatim.org api, see https://nominatim.org/release-docs/develop/api/Search/ for more information
	query := buildNominatimURL(run, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), positionRequestTimeout)
	defer cancel()

	slog.Info("Fetching coordinates for run", "runID", run.ID.String(), "query", query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, query, nil) // #nosec G107
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "IRIS/1.0 (github.com/FAECHER/IRIS-Server)")

	resp, err := http.DefaultClient.Do(req) // #nosec G107
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get coordinates: %w", err)
	}
	defer resp.Body.Close()

	responseStatus := resp.StatusCode
	if responseStatus != http.StatusOK {
		return 0, 0, fmt.Errorf("nominatim API returned non-200 status: %d", responseStatus)
	}

	// Parse the JSON response
	var results []struct {
		Lat float64 `json:"lat,string"`
		Lon float64 `json:"lon,string"`
	}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode coordinates: %w", err)
	}

	if len(results) == 0 {
		return 0, 0, ErrLocationNotFound
	}

	slog.Info("Coordinates fetched for run", "runID", run.ID.String(), "latitude", results[0].Lat, "longitude", results[0].Lon)

	return results[0].Lat, results[0].Lon, nil
}

func buildNominatimURL(run models.Run, cfg config.GeocodingConfig) string {
	params := url.Values{}
	params.Set("street", run.Street+" "+run.House)
	params.Set("city", run.City)
	params.Set("format", "json")
	params.Set("limit", "1")

	geocodingUrl := url.URL{
		Scheme:   cfg.Scheme,
		Host:     cfg.Host,
		Path:     "/search",
		RawQuery: params.Encode(),
	}
	return geocodingUrl.String()
}
