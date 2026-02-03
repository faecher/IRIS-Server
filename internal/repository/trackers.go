package repository

import (
	"IRIS-Server/internal/models"
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

var (
	// ErrNoDataToSave indicates that there is no data to save in the provided struct
	ErrNoDataToSave = errors.New("no data to save in struct")
)

// GetActiveTrackerCount returns the number of trackers updated within the last 5 minutes
func GetActiveTrackerCount() (int, error) {
	SQL := `SELECT COUNT(*) FROM trackers WHERE updated_at >= NOW() - INTERVAL '5 minutes'`

	var count int
	err := DBConnPool.QueryRow(context.Background(), SQL).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to query active tracker count: %w", err)
	}

	return count, nil
}

// GetTrackerByDevEUI returns the tracker ID for a given Chirpstack DevEUI. Returns uuid.Nil if not found.
func GetTrackerByDevEUI(devEUI string) (uuid.UUID, error) {
	SQL := `
		SELECT tracker_id
		FROM chirpstack_trackers
		WHERE dev_eui = $1
	`

	var trackerID uuid.UUID
	err := DBConnPool.QueryRow(context.Background(), SQL, devEUI).Scan(&trackerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, nil // Tracker not found
	} else if err != nil {
		return uuid.Nil, fmt.Errorf("failed to query tracker by devEUI: %w", err)
	}

	return trackerID, nil
}

// GetTrackerByID retrieves a single tracker by UUID
func GetTrackerByID(trackerID uuid.UUID) (*models.BaseTracker, error) {
	SQL := `
		SELECT trackers.tracker_id, name, battery, position_longitude, position_latitude, updated_at, tr.resource_id
		FROM trackers
		LEFT JOIN trackers_resource tr ON trackers.tracker_id = tr.tracker_id
		WHERE trackers.tracker_id = $1
	`

	// get base tracker info
	var tracker models.BaseTracker
	var resourceID *uuid.UUID
	err := DBConnPool.QueryRow(context.Background(), SQL, trackerID).Scan(
		&tracker.ID,
		&tracker.Name,
		&tracker.Battery,
		&tracker.Position.Longitude,
		&tracker.Position.Latitude,
		&tracker.LastUpdate,
		&resourceID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query tracker by ID: %w", err)
	}

	if resourceID != nil {
		err = fillTrackerResource(&tracker, *resourceID)
		if err != nil {
			return nil, fmt.Errorf("failed to fill tracker resource: %w", err)
		}
	}

	return &tracker, nil
}

// GetAllTrackers retrieves all trackers with their assigned resources
func GetAllTrackers() ([]models.Tracker, error) {
	SQL := `
		SELECT 
			t.tracker_id, t.name, t.battery, t.position_longitude, t.position_latitude, t.updated_at,
			CASE 
				WHEN ct.dev_eui IS NOT NULL THEN 'chirpstack'
				WHEN tt.issi IS NOT NULL THEN 'tetra'
				ELSE 'unknown'
			END as tracker_type,
			ct.dev_eui,
			tt.issi,
			tr.resource_id
		FROM trackers t
		LEFT JOIN chirpstack_trackers ct ON t.tracker_id = ct.tracker_id
		LEFT JOIN tetra_trackers tt ON t.tracker_id = tt.tracker_id
		LEFT JOIN trackers_resource tr ON t.tracker_id = tr.tracker_id
		LEFT JOIN resources r ON tr.resource_id = r.resource_id
	`

	rows, err := DBConnPool.Query(context.Background(), SQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query all trackers: %w", err)
	}
	defer rows.Close()

	trackers := make([]models.Tracker, 0)
	for rows.Next() {
		var base models.BaseTracker
		var trackerType string
		var devEUI, issi *string
		var resourceID *uuid.UUID

		err = rows.Scan(&base.ID, &base.Name, &base.Battery,
			&base.Position.Longitude, &base.Position.Latitude,
			&base.LastUpdate, &trackerType, &devEUI, &issi, &resourceID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tracker row: %w", err)
		}

		if resourceID != nil {
			err = fillTrackerResource(&base, *resourceID)
			if err != nil {
				return nil, fmt.Errorf("failed to fill tracker resource: %w", err)
			}
		}

		switch trackerType {
		case "chirpstack":
			trackers = append(trackers, &models.ChirpstackTracker{
				BaseTracker: base,
				DevEUI:      *devEUI,
			})
		case "tetra":
			trackers = append(trackers, &models.TetraTracker{
				BaseTracker: base,
				ISSI:        *issi,
			})
		default:
			trackers = append(trackers, &base)
		}
	}

	return trackers, nil
}

// UpdateTrackerResource assigns a resource to a tracker, or updates an existing assignment
func UpdateTrackerResource(trackerID, resourceID uuid.UUID) error {
	SQL := `
		INSERT INTO trackers_resource (tracker_id, resource_id)
		VALUES ($1, $2)
		ON CONFLICT (tracker_id) DO UPDATE SET resource_id = EXCLUDED.resource_id
	`

	_, err := DBConnPool.Exec(context.Background(), SQL, trackerID, resourceID)
	if err != nil {
		return fmt.Errorf("failed to update tracker resource: %w", err)
	}

	return nil
}

// RemoveTrackerAssignment removes a resource assignment from a tracker
func RemoveTrackerAssignment(trackerID uuid.UUID) error {
	SQL := `DELETE FROM trackers_resource WHERE tracker_id = $1`

	_, err := DBConnPool.Exec(context.Background(), SQL, trackerID)
	if err != nil {
		return fmt.Errorf("failed to remove tracker assignment: %w", err)
	}

	return nil
}

// RenameTracker updates the name of a tracker
func RenameTracker(trackerID uuid.UUID, newName string) error {
	SQL := `UPDATE trackers SET name = $1 WHERE tracker_id = $2`

	_, err := DBConnPool.Exec(context.Background(), SQL, newName, trackerID)
	if err != nil {
		return fmt.Errorf("failed to rename tracker: %w", err)
	}

	return nil
}

// UpdateTracker updates battery and position values for a tracker.
// Skips battery if < 0, skips position if latitude or longitude is infinity.
// Caller is responsible to make sure the given tracker ID exists. Otherwise you will get an error
func UpdateTracker(tracker models.BaseTracker) error {
	SQLLatLong := `UPDATE trackers 
	SET position_latitude = $1, position_longitude = $2 
	WHERE tracker_id = $3`

	SQLBatt := `UPDATE trackers 
	SET battery = $1 
	WHERE tracker_id = $2`

	batteryUpdate := tracker.Battery >= 0
	locationUpdate := !math.IsInf(tracker.Position.Latitude, 0) && !math.IsInf(tracker.Position.Longitude, 0)

	if !batteryUpdate && !locationUpdate {
		return ErrNoDataToSave
	}

	if batteryUpdate {
		_, err := DBConnPool.Exec(context.Background(), SQLBatt,
			tracker.Battery, tracker.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update tracker battery: %w", err)
		}
	}
	if locationUpdate {
		_, err := DBConnPool.Exec(context.Background(), SQLLatLong,
			tracker.Position.Latitude, tracker.Position.Longitude, tracker.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update tracker position: %w", err)
		}
	}

	return nil
}

func fillTrackerResource(tracker *models.BaseTracker, resourceID uuid.UUID) error {
	if resourceID == uuid.Nil {
		return nil
	}

	resource, err := GetResourceByID(resourceID)
	if err != nil && !errors.Is(err, ErrNoResourceFound) {
		return fmt.Errorf("failed to get resource by ID: %w", err)
	}

	tracker.Resource = resource

	return nil
}
