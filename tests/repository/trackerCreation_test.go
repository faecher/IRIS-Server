package repository_test

import (
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"context"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateChirpstackTracker(t *testing.T) {
	t.Run("create valid chirpstack tracker", func(t *testing.T) {
		tracker := &models.ChirpstackTracker{
			BaseTracker: models.BaseTracker{
				Name:    "Test Tracker",
				Battery: 85,
				Position: models.Position{
					Latitude:  52.5200,
					Longitude: 13.4050,
				},
			},
			DevEUI: "0102030405060708",
		}

		err := repository.CreateChirpstackTracker(tracker)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, tracker.ID, "ID should be set after creation")

		defer cleanupTracker(t, tracker.ID)

		// Verify tracker was created
		saved, err := repository.GetTrackerByID(tracker.ID)
		require.NoError(t, err)
		assert.Equal(t, tracker.Name, saved.Name)
		assert.Equal(t, tracker.Battery, saved.Battery)
		assert.Equal(t, tracker.Position.Latitude, saved.Position.Latitude)
		assert.Equal(t, tracker.Position.Longitude, saved.Position.Longitude)
	})

	t.Run("create tracker with negative battery", func(t *testing.T) {
		tracker := &models.ChirpstackTracker{
			BaseTracker: models.BaseTracker{
				Name:    "Negative Battery Tracker",
				Battery: -1,
				Position: models.Position{
					Latitude:  0.0,
					Longitude: 0.0,
				},
			},
			DevEUI: "0102030405060709",
		}

		err := repository.CreateChirpstackTracker(tracker)
		require.NoError(t, err)
		defer cleanupTracker(t, tracker.ID)

		saved, err := repository.GetTrackerByID(tracker.ID)
		require.NoError(t, err)
		assert.Equal(t, int16(-1), saved.Battery)
	})

	t.Run("create tracker with zero coordinates", func(t *testing.T) {
		tracker := &models.ChirpstackTracker{
			BaseTracker: models.BaseTracker{
				Name:    "Zero Position Tracker",
				Battery: 100,
				Position: models.Position{
					Latitude:  0.0,
					Longitude: 0.0,
				},
			},
			DevEUI: "010203040506070A",
		}

		err := repository.CreateChirpstackTracker(tracker)
		require.NoError(t, err)
		defer cleanupTracker(t, tracker.ID)

		saved, err := repository.GetTrackerByID(tracker.ID)
		require.NoError(t, err)
		assert.Equal(t, 0.0, saved.Position.Latitude)
		assert.Equal(t, 0.0, saved.Position.Longitude)
	})

	t.Run("create tracker with extreme coordinates", func(t *testing.T) {
		tracker := &models.ChirpstackTracker{
			BaseTracker: models.BaseTracker{
				Name:    "Extreme Coords Tracker",
				Battery: 50,
				Position: models.Position{
					Latitude:  -90.0,
					Longitude: 180.0,
				},
			},
			DevEUI: "010203040506070B",
		}

		err := repository.CreateChirpstackTracker(tracker)
		require.NoError(t, err)
		defer cleanupTracker(t, tracker.ID)

		saved, err := repository.GetTrackerByID(tracker.ID)
		require.NoError(t, err)
		assert.Equal(t, -90.0, saved.Position.Latitude)
		assert.Equal(t, 180.0, saved.Position.Longitude)
	})

	t.Run("create tracker with empty name", func(t *testing.T) {
		tracker := &models.ChirpstackTracker{
			BaseTracker: models.BaseTracker{
				Name:    "",
				Battery: 75,
				Position: models.Position{
					Latitude:  10.0,
					Longitude: 20.0,
				},
			},
			DevEUI: "010203040506070C",
		}

		err := repository.CreateChirpstackTracker(tracker)
		require.NoError(t, err)
		defer cleanupTracker(t, tracker.ID)

		saved, err := repository.GetTrackerByID(tracker.ID)
		require.NoError(t, err)
		assert.Equal(t, "", saved.Name)
	})

	t.Run("create tracker with long name", func(t *testing.T) {
		longName := "Very long tracker name with lots of characters to test database limits and repeated multiple times to exceed typical length"

		tracker := &models.ChirpstackTracker{
			BaseTracker: models.BaseTracker{
				Name:    longName,
				Battery: 60,
				Position: models.Position{
					Latitude:  45.0,
					Longitude: 90.0,
				},
			},
			DevEUI: "010203040506070D",
		}

		err := repository.CreateChirpstackTracker(tracker)
		require.NoError(t, err)
		defer cleanupTracker(t, tracker.ID)
	})

	t.Run("create tracker with special characters in name", func(t *testing.T) {
		tracker := &models.ChirpstackTracker{
			BaseTracker: models.BaseTracker{
				Name:    "Tracker-123 (Test) #1 <Special>",
				Battery: 90,
				Position: models.Position{
					Latitude:  51.5074,
					Longitude: -0.1278,
				},
			},
			DevEUI: "010203040506070E",
		}

		err := repository.CreateChirpstackTracker(tracker)
		require.NoError(t, err)
		defer cleanupTracker(t, tracker.ID)

		saved, err := repository.GetTrackerByID(tracker.ID)
		require.NoError(t, err)
		assert.Contains(t, saved.Name, "Special")
	})

	t.Run("create tracker with max battery", func(t *testing.T) {
		tracker := &models.ChirpstackTracker{
			BaseTracker: models.BaseTracker{
				Name:    "Max Battery Tracker",
				Battery: 32767, // max int16
				Position: models.Position{
					Latitude:  1.0,
					Longitude: 1.0,
				},
			},
			DevEUI: "010203040506070F",
		}

		err := repository.CreateChirpstackTracker(tracker)
		require.NoError(t, err)
		defer cleanupTracker(t, tracker.ID)

		saved, err := repository.GetTrackerByID(tracker.ID)
		require.NoError(t, err)
		assert.Equal(t, int16(32767), saved.Battery)
	})

	t.Run("create tracker with duplicate DevEUI fails", func(t *testing.T) {
		devEUI := "DUPLICATE0000001"

		tracker1 := &models.ChirpstackTracker{
			BaseTracker: models.BaseTracker{
				Name:    "First Tracker",
				Battery: 50,
				Position: models.Position{
					Latitude:  1.0,
					Longitude: 1.0,
				},
			},
			DevEUI: devEUI,
		}

		err := repository.CreateChirpstackTracker(tracker1)
		require.NoError(t, err)
		defer cleanupTracker(t, tracker1.ID)

		tracker2 := &models.ChirpstackTracker{
			BaseTracker: models.BaseTracker{
				Name:    "Second Tracker",
				Battery: 60,
				Position: models.Position{
					Latitude:  2.0,
					Longitude: 2.0,
				},
			},
			DevEUI: devEUI,
		}

		err = repository.CreateChirpstackTracker(tracker2)
		assert.Error(t, err, "Duplicate DevEUI should fail")
	})

	t.Run("create tracker with empty DevEUI", func(t *testing.T) {
		tracker := &models.ChirpstackTracker{
			BaseTracker: models.BaseTracker{
				Name:    "Empty DevEUI Tracker",
				Battery: 80,
				Position: models.Position{
					Latitude:  3.0,
					Longitude: 3.0,
				},
			},
			DevEUI: "",
		}

		err := repository.CreateChirpstackTracker(tracker)
		require.NoError(t, err)
		defer cleanupTracker(t, tracker.ID)
	})

	t.Run("create multiple trackers", func(t *testing.T) {
		trackers := []*models.ChirpstackTracker{
			{
				BaseTracker: models.BaseTracker{
					Name:     "Tracker 1",
					Battery:  10,
					Position: models.Position{Latitude: 10.0, Longitude: 10.0},
				},
				DevEUI: "MULTI0000000001",
			},
			{
				BaseTracker: models.BaseTracker{
					Name:     "Tracker 2",
					Battery:  20,
					Position: models.Position{Latitude: 20.0, Longitude: 20.0},
				},
				DevEUI: "MULTI0000000002",
			},
			{
				BaseTracker: models.BaseTracker{
					Name:     "Tracker 3",
					Battery:  30,
					Position: models.Position{Latitude: 30.0, Longitude: 30.0},
				},
				DevEUI: "MULTI0000000003",
			},
		}

		for _, tracker := range trackers {
			err := repository.CreateChirpstackTracker(tracker)
			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, tracker.ID)
			defer cleanupTracker(t, tracker.ID)
		}

		// Verify all have unique IDs
		ids := make(map[uuid.UUID]bool)
		for _, tracker := range trackers {
			assert.False(t, ids[tracker.ID], "IDs should be unique")
			ids[tracker.ID] = true
		}
	})

	t.Run("create tracker with unicode in name", func(t *testing.T) {
		tracker := &models.ChirpstackTracker{
			BaseTracker: models.BaseTracker{
				Name:    "Трекер 中文 🚗",
				Battery: 95,
				Position: models.Position{
					Latitude:  35.6762,
					Longitude: 139.6503,
				},
			},
			DevEUI: "0102030405060710",
		}

		err := repository.CreateChirpstackTracker(tracker)
		require.NoError(t, err)
		defer cleanupTracker(t, tracker.ID)

		saved, err := repository.GetTrackerByID(tracker.ID)
		require.NoError(t, err)
		assert.Contains(t, saved.Name, "中文")
	})

	t.Run("ID should not be preserved from input", func(t *testing.T) {
		originalID := uuid.Must(uuid.NewV4())
		tracker := &models.ChirpstackTracker{
			BaseTracker: models.BaseTracker{
				ID:      originalID,
				Name:    "Test ID Override",
				Battery: 75,
				Position: models.Position{
					Latitude:  5.0,
					Longitude: 5.0,
				},
			},
			DevEUI: "0102030405060711",
		}

		err := repository.CreateChirpstackTracker(tracker)
		require.NoError(t, err)
		defer cleanupTracker(t, tracker.ID)

		// The ID should be newly generated, not the original
		assert.NotEqual(t, originalID, tracker.ID)
		assert.NotEqual(t, uuid.Nil, tracker.ID)
	})
}

// Helper to cleanup test trackers
func cleanupTracker(t *testing.T, trackerID uuid.UUID) {
	SQL1 := `DELETE FROM chirpstack_trackers WHERE tracker_id = $1`
	SQL2 := `DELETE FROM tetra_trackers WHERE tracker_id = $1`
	SQL3 := `DELETE FROM trackers WHERE tracker_id = $1`

	repository.DBConnPool.Exec(context.Background(), SQL1, trackerID)
	repository.DBConnPool.Exec(context.Background(), SQL2, trackerID)
	_, err := repository.DBConnPool.Exec(context.Background(), SQL3, trackerID)
	if err != nil {
		t.Logf("Failed to cleanup tracker %s: %v", trackerID, err)
	}
}
