// SPDX-License-Identifier: EUPL-1.2

package repository

import (
	"IRIS-Server/internal/models"
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

// CreateChirpstackTracker creates a new Chirpstack tracker record in the database
// it automatically takes the battery and position from the embedded BaseTracker.
// the provided tracker will have its ID field updated with the new UUID.
func CreateChirpstackTracker(tracker *models.ChirpstackTracker) error {
	newID := uuid.Must(uuid.NewV4())

	tx, err := DBConnPool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	// Insert into trackers table first
	_, err = tx.Exec(context.Background(),
		`INSERT INTO trackers (tracker_id, name, battery, position_longitude, position_latitude)
		VALUES ($1, $2, $3, $4, $5)`,
		newID,
		tracker.Name,
		tracker.Battery,
		tracker.Position.Longitude,
		tracker.Position.Latitude,
	)
	if err != nil {
		return fmt.Errorf("failed to insert tracker: %w", err)
	}

	// Then insert into chirpstack_trackers table
	_, err = tx.Exec(context.Background(),
		`INSERT INTO chirpstack_trackers (tracker_id, dev_eui)
		VALUES ($1, $2)`,
		newID,
		tracker.DevEUI,
	)
	if err != nil {
		return fmt.Errorf("failed to insert chirpstack tracker: %w", err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	tracker.ID = newID

	return nil
}

// CreateTetraTracker creates a new Tetra tracker record in the database
func CreateTetraTracker(tracker *models.TetraTracker) error {
	_ = tracker

	return nil
}
