package repository

import (
	"IRIS-Server/internal/models"
	"context"

	"github.com/gofrs/uuid/v5"
)

// CreateChirpstackTracker creates a new Chirpstack tracker record in the database
// it automatically takes the battery and position from the embedded BaseTracker.
// the provided tracker will have its ID field updated with the new UUID.
func CreateChirpstackTracker(tracker *models.ChirpstackTracker) error {
	newID := uuid.Must(uuid.NewV4())

	// Insert into trackers table first
	_, err := DBConnPool.Exec(context.Background(),
		`INSERT INTO trackers (tracker_id, name, battery, position_longitude, position_latitude)
		VALUES ($1, $2, $3, $4, $5)`,
		newID,
		tracker.Name,
		tracker.Battery,
		tracker.Position.Longitude,
		tracker.Position.Latitude,
	)
	if err != nil {
		return err
	}

	// Then insert into chirpstack_trackers table
	_, err = DBConnPool.Exec(context.Background(),
		`INSERT INTO chirpstack_trackers (tracker_id, dev_eui)
		VALUES ($1, $2)`,
		newID,
		tracker.DevEUI,
	)
	if err != nil {
		return err
	}

	tracker.ID = newID

	return nil
}

func CreateTetraTracker(tracker *models.TetraTracker) error {
	// TODO
	return nil
}
