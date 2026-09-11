package repository

import (
	"IRIS-Server/internal/models"
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

// GetAllRuns retrieves all MCP runs for the currently selected operation
func GetAllRuns() ([]models.MCPRun, error) {
	SQL := `
	SELECT
		run_id,
		operation_id,
		house_object,
		place, address_field,
		latitude, longitude,
		has_patient, active
	FROM mcp_runs
	WHERE operation_id = (SELECT operation_id FROM mcp_config WHERE id = 1) AND active = true`

	rows, err := DBConnPool.Query(context.Background(), SQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query all resources: %w", err)
	}
	defer rows.Close()

	var runs []models.MCPRun
	for rows.Next() {
		var run models.MCPRun
		err := rows.Scan(
			&run.ID,
			&run.Operation.ID,
			&run.House,
			&run.City,
			&run.Street,
			&run.Latitude,
			&run.Longitude,
			&run.HasPatient,
			&run.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan run row: %w", err)
		}
		runs = append(runs, run)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error iterating run rows: %w", err)
	}

	return runs, nil
}

// GetRunByID retrieves a specific MCP run by its ID
func GetRunByID(runID uuid.UUID) (*models.MCPRun, error) {
	SQL := `
	SELECT
		run_id,
		operation_id,
		house_object,
		place, address_field,
		latitude, longitude,
		has_patient, active
	FROM mcp_runs
	WHERE run_id = $1`

	row := DBConnPool.QueryRow(context.Background(), SQL, runID)

	var run models.MCPRun
	err := row.Scan(
		&run.ID,
		&run.Operation.ID,
		&run.House,
		&run.City,
		&run.Street,
		&run.Latitude,
		&run.Longitude,
		&run.HasPatient,
		&run.Active,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan run row: %w", err)
	}

	return &run, nil
}

// UpdateRunPosition updates the position (latitude and longitude) of a specific MCP run
func UpdateRunPosition(runID string, latitude, longitude float64) error {
	SQL := `
	UPDATE mcp_runs
	SET latitude = $1, longitude = $2
	WHERE run_id = $3`

	_, err := DBConnPool.Exec(context.Background(), SQL, latitude, longitude, runID)
	if err != nil {
		return fmt.Errorf("failed to update run position: %w", err)
	}

	return nil
}

// UpsertRun creates or updates an MCP run in the database
func UpsertRun(run *models.MCPRun) error {
	runSQL := `
	INSERT INTO mcp_runs (
	 	run_id,
		operation_id,
		house_object,
		place, address_field,
		latitude, longitude,
		has_patient, active
	) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
	ON CONFLICT (run_id) DO UPDATE 
	SET operation_id = EXCLUDED.operation_id,
	    place = EXCLUDED.place,
	    address_field = EXCLUDED.address_field,
	    latitude = EXCLUDED.latitude,
	    longitude = EXCLUDED.longitude,
	    has_patient = EXCLUDED.has_patient,
	    active = EXCLUDED.active`

	_, err := DBConnPool.Exec(context.Background(), runSQL,
		run.ID,
		run.Operation.ID,
		run.House,
		run.City,
		run.Street,
		run.Latitude,
		run.Longitude,
		run.HasPatient,
		run.Active,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert MCP run: %w", err)
	}

	return nil
}
