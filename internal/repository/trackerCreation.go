// SPDX-License-Identifier: EUPL-1.2

package repository

import (
	"IRIS-Server/internal/models"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateChirpstackTracker creates a new Chirpstack tracker record in the database
// it automatically takes the battery and position from the embedded BaseTracker.
// the provided tracker will have its ID field updated with the new UUID.
func CreateChirpstackTracker(tracker *models.ChirpstackTracker) error {
	// Insert into trackers table first
	transaction, newID, err := startTransactionAndInsertGenericTracker(tracker.BaseTracker)
	if err != nil {
		return fmt.Errorf("error inserting generic tracker portion for Tracca Tracker: %w", err)
	}

	// Then insert into chirpstack_trackers table
	_, err = transaction.Exec(context.Background(),
		`INSERT INTO chirpstack_trackers (tracker_id, dev_eui)
		VALUES ($1, $2)`,
		newID,
		tracker.DevEUI,
	)
	if err != nil {
		return fmt.Errorf("failed to insert chirpstack tracker: %w", err)
	}

	err = transaction.Commit(context.Background())
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

// CreateTraccarTracker creates a new Traccar tracker record in the database
func CreateTraccarTracker(tracker *models.TraccarTracker) error {
	// Insert into trackers table first
	transaction, newID, err := startTransactionAndInsertGenericTracker(tracker.BaseTracker)
	if err != nil {
		return fmt.Errorf("error inserting generic tracker portion for Tracca Tracker: %w", err)
	}

	// Then insert into traccar_trackers table
	_, err = transaction.Exec(context.Background(),
		`INSERT INTO traccar_trackers (tracker_id, traccar_id)
		VALUES ($1, $2)`,
		newID,
		tracker.TraccarID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert traccar tracker: %w", err)
	}

	err = transaction.Commit(context.Background())
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	tracker.ID = newID

	return nil
}

func startTransactionAndInsertGenericTracker(tracker models.BaseTracker) (*pgxpool.Tx, uuid.UUID, error) {
	newID := uuid.Must(uuid.NewV4())

	transaction, err := DBConnPool.Begin(context.Background())
	if err != nil {
		return nil, newID, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		err := transaction.Rollback(context.Background())
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("failed to rollback transaction", "error", err)
		}
	}()

	_, err = transaction.Exec(context.Background(),
		`INSERT INTO trackers (tracker_id, name, battery, position_longitude, position_latitude, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		newID,
		tracker.Name,
		tracker.Battery,
		tracker.Position.Longitude,
		tracker.Position.Latitude,
		tracker.LastUpdate,
	)
	if err != nil {
		return nil, newID, fmt.Errorf("failed to insert tracker: %w", err)
	}

	// The documentation of pgxpool.Pool::Begin states it will return this type, so check should not be required
	// We cast to avoid returning an interface
	typedTranasaction, _ := transaction.(*pgxpool.Tx)
	return typedTranasaction, newID, nil
}
