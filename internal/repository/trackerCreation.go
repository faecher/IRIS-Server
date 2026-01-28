package repository

import "IRIS-Server/internal/models"

// TODO: mention that this file should be edited for another tracker addition

// CreateChirpstackTracker creates a new Chirpstack tracker record in the database
// it automatically takes the battery and position from the embedded BaseTracker.
// the provided tracker will have its ID field updated with the new UUID.
func CreateChirpstackTracker(tracker *models.ChirpstackTracker) error {
	// TODO
	return nil
}

func CreateTetraTracker(tracker models.TetraTracker) error {
	// TODO
	return nil
}
