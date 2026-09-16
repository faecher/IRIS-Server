package repository

import (
	"IRIS-Server/internal/models"
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

// GetAllRuns retrieves all MCP runs for the currently selected operation
func GetAllRuns() ([]models.Run, error) {
	SQL := `
	SELECT
		run_id,
		operation_id,
		house_object,
		place, address_field,
		COALESCE(position_latitude, 0),
		COALESCE(position_longitude, 0),
		unset_position,
		has_patient, active,
		COALESCE(nr, 0),
		COALESCE(notes, ''),
		COALESCE(deleted, false)
	FROM runs
	WHERE operation_id = (SELECT operation_id FROM mcp_config WHERE id = 1) AND active = true AND NOT deleted`

	rows, err := DBConnPool.Query(context.Background(), SQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query all resources: %w", err)
	}
	defer rows.Close()

	var runs []models.Run
	for rows.Next() {
		var run models.Run
		err := rows.Scan(
			&run.ID,
			&run.OperationID,
			&run.House,
			&run.City,
			&run.Street,
			&run.Latitude,
			&run.Longitude,
			&run.UnsetPosition,
			&run.HasPatient,
			&run.Active,
			&run.Nr,
			&run.Text,
			&run.Deleted,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan run row: %w", err)
		}
		run.Operation.ID = run.OperationID
		runs = append(runs, run)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error iterating run rows: %w", err)
	}

	return runs, nil
}

// GetRunByID retrieves a specific MCP run by its ID
func GetRunByID(runID uuid.UUID) (*models.Run, error) {
	SQL := `
	SELECT
		run_id,
		operation_id,
		house_object,
		place, address_field,
		COALESCE(position_latitude, 0),
		COALESCE(position_longitude, 0),
		unset_position,
		has_patient, active,
		COALESCE(nr, 0),
		COALESCE(notes, ''),
		COALESCE(deleted, false)
	FROM runs
	WHERE run_id = $1`

	row := DBConnPool.QueryRow(context.Background(), SQL, runID)

	var run models.Run
	err := row.Scan(
		&run.ID,
		&run.OperationID,
		&run.House,
		&run.City,
		&run.Street,
		&run.Latitude,
		&run.Longitude,
		&run.UnsetPosition,
		&run.HasPatient,
		&run.Active,
		&run.Nr,
		&run.Text,
		&run.Deleted,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan run row: %w", err)
	}
	run.Operation.ID = run.OperationID

	return &run, nil
}

// UpdateRunPosition updates the position (latitude and longitude) of a specific MCP run
func UpdateRunPosition(runID string, latitude, longitude float64) error {
	SQL := `
	UPDATE runs
	SET position_latitude = $1, position_longitude = $2, unset_position = FALSE
	WHERE run_id = $3`

	_, err := DBConnPool.Exec(context.Background(), SQL, latitude, longitude, runID)
	if err != nil {
		return fmt.Errorf("failed to update run position: %w", err)
	}

	return nil
}

// UpsertRun creates or updates an MCP run in the database
func UpsertRun(run *models.Run) error {
	runSQL := `
	INSERT INTO runs (
	 	run_id,
		operation_id,
		house_object,
		place, address_field,
		position_latitude, position_longitude,
		unset_position,
			has_patient, active,
		nr, notes, deleted
	) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	ON CONFLICT (run_id) DO UPDATE 
	SET operation_id = EXCLUDED.operation_id,
		house_object = EXCLUDED.house_object,
	    place = EXCLUDED.place,
	    address_field = EXCLUDED.address_field,
	    position_latitude = EXCLUDED.position_latitude,
	    position_longitude = EXCLUDED.position_longitude,
	    unset_position = EXCLUDED.unset_position,
	    has_patient = EXCLUDED.has_patient,
	    active = EXCLUDED.active,
	    nr = EXCLUDED.nr,
	    notes = EXCLUDED.notes,
	    deleted = EXCLUDED.deleted`

	_, err := DBConnPool.Exec(context.Background(), runSQL,
		run.ID,
		run.Operation.ID,
		run.House,
		run.City,
		run.Street,
		run.Latitude,
		run.Longitude,
		run.UnsetPosition,
		run.HasPatient,
		run.Active,
		run.Nr,
		run.Text,
		run.Deleted,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert MCP run: %w", err)
	}

	return nil
}

// DeleteAllRuns deletes all runs from the database by running truncate table runs
func DeleteAllRuns() error {
	const sql = `TRUNCATE TABLE runs`
	
	_, err := DBConnPool.Exec(context.Background(), sql)
	if err != nil {
		return fmt.Errorf("failed to update run position: %w", err)
	}

	return nil
}
