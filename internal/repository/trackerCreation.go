package repository

import "IRIS-Server/internal/models"

// TODO: mention that this file should be edited for another tracker addition

// CreateChirpstackTracker creates a new Chirpstack tracker record in the database
// it automatically takes the battery and position from the embedded BaseTracker.
// the provided tracker will have its ID field updated with the new UUID.
func CreateChirpstackTracker(tracker *models.ChirpstackTracker) error {
	SQL := `
		INSERT INTO trackers (tracker_id, name, battery, position_longitude, position_latitude)
		VALUES ($1, $2, $3, $4, $5);
		INSERT INTO chirpstack_trackers (tracker_id, dev_eui)
		VALUES ($1, $6);
	`

	newID := uuid.Must(uuid.NewV4())

	_, err := DBConnPool.Exec(context.Background(), SQL,
		newID,
		tracker.Name,
		tracker.Battery,
		tracker.Position.Longitude,
		tracker.Position.Latitude,
		tracker.DevEUI,
	)
	if err != nil {
		return err
	}

	tracker.ID = newID

	return nil
}

func CreateTetraTracker(tracker models.TetraTracker) error {
	// TODO
	return nil
}
