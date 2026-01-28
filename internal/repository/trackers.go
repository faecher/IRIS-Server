package repository

import (
	"IRIS-Server/internal/models"
	"context"
	"errors"
	"math"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

var (
	ErrNoDataToSave = errors.New("no data to save in struct")
)

func GetActiveTrackerCount() (int, error) {
	SQL := `SELECT COUNT(*) FROM trackers WHERE updated_at >= NOW() - INTERVAL '5 minutes'`

	var count int
	err := DBConnPool.QueryRow(context.Background(), SQL).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetTrackerByDevEUI(devEUI string) (uuid.UUID, error) {
	SQL := `
		SELECT tracker_id
		FROM chirpstack_trackers
		WHERE dev_eui = $1
	`

	var trackerID uuid.UUID
	err := DBConnPool.QueryRow(context.Background(), SQL, devEUI).Scan(&trackerID)
	if err == pgx.ErrNoRows {
		return uuid.Nil, nil // Tracker not found
	} else if err != nil {
		return uuid.Nil, err
	}

	return trackerID, nil
}

// GetTrackerByID retrieves a single tracker by UUID
func GetTrackerByID(trackerID uuid.UUID) (*models.BaseTracker, error) {
	SQL := `
		SELECT tracker_id, name, battery, position_longitude, position_latitude, updated_at, resource_id
		FROM trackers
		LEFT JOIN tracker_resources tr ON trackers.tracker_id = tr.tracker_id
		WHERE tracker_id = $1
	`

	// get base tracker info
	var tracker models.BaseTracker
	var resourceID uuid.UUID
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
		return nil, err
	}

	err = fillTrackerResource(&tracker, resourceID)
	if err != nil {
		return nil, err
	}

	return &tracker, nil
}

// GetAllTrackers retrieves all trackers from the database
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
		FROM trackers t
		LEFT JOIN chirpstack_trackers ct ON t.tracker_id = ct.tracker_id
		LEFT JOIN tetra_trackers tt ON t.tracker_id = tt.tracker_id
		LEFT JOIN tracker_resources tr ON t.tracker_id = tr.tracker_id
		LEFT JOIN resources r ON tr.resource_id = r.resource_id
	`

	rows, err := DBConnPool.Query(context.Background(), SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trackers []models.Tracker
	for rows.Next() {
		var base models.BaseTracker
		var trackerType string
		var devEUI, issi *string

		rows.Scan(&base.ID, &base.Name, &base.Battery,
			&base.Position.Longitude, &base.Position.Latitude,
			&base.LastUpdate, &trackerType, &devEUI, &issi)

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
		}

		err = fillTrackerResource(&base, base.AssignedResource)
		if err != nil {
			return nil, err
		}
	}

	return trackers, nil
}

func UpdateTrackerResource(trackerID, resourceID uuid.UUID) error {
	SQL := `
		INSERT INTO tracker_resources (tracker_id, resource_id)
		VALUES ($1, $2)
		ON CONFLICT (tracker_id) DO UPDATE SET resource_id = EXCLUDED.resource_id
	`

	_, err := DBConnPool.Exec(context.Background(), SQL, trackerID, resourceID)
	return err
}

func RemoveTrackerAssignment(trackerID uuid.UUID) error {
	SQL := `DELETE FROM tracker_resources WHERE tracker_id = $1`

	_, err := DBConnPool.Exec(context.Background(), SQL, trackerID)
	return err
}

func RenameTracker(trackerID uuid.UUID, newName string) error {
	SQL := `UPDATE trackers SET name = $1 WHERE tracker_id = $2`

	_, err := DBConnPool.Exec(context.Background(), SQL, newName, trackerID)
	return err
}

// UpdateTracker updates the DB entry for battery, lat and long values.
// if the battery is < 0 or lat/long values are -inf, they are skipped
// Caller is responsible to make sure the given tracker ID exists. Otherwise you will get an error
func UpdateTracker(tracker models.BaseTracker) error {
	SQL_LAT_LONG := `UPDATE trackers 
	SET position_latitude = $1, position_longitude = $2 
	WHERE tracker_id = $3`

	SQL_BATT := `UPDATE trackers 
	SET battery = $1 
	WHERE tracker_id = $2`

	batteryUpdate := tracker.Battery >= 0
	locationUpdate := !math.IsInf(tracker.Position.Latitude, -1) && !math.IsInf(tracker.Position.Longitude, -1)

	if !batteryUpdate && !locationUpdate {
		return ErrNoDataToSave
	}

	if batteryUpdate {
		_, err := DBConnPool.Exec(context.Background(), SQL_BATT,
			tracker.Battery, tracker.ID,
		)
		if err != nil {
			return err
		}
	}
	if locationUpdate {
		_, err := DBConnPool.Exec(context.Background(), SQL_LAT_LONG,
			tracker.Position.Latitude, tracker.Position.Longitude, tracker.ID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func fillTrackerResource(tracker *models.BaseTracker, resourceID uuid.UUID) error {
	if resourceID == uuid.Nil {
		return nil
	}

	resource, err := GetResourceByID(resourceID)
	if err != nil {
		return err
	}

	tracker.Resource = *resource

	return nil
}
