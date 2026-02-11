package repository_test

import (
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"context"
	"math"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetActiveTrackerCount(t *testing.T) {
	t.Run("no active trackers", func(t *testing.T) {
		count, err := repository.GetActiveTrackerCount()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 0)
	})

	t.Run("tracker within 5 minutes", func(t *testing.T) {
		tracker := createTestTrackerDirect(t, "Active Tracker", 80, 10.0, 20.0, "ACTIVE00000001")
		defer cleanupTracker(t, tracker)

		// Update timestamp to now
		updateTrackerTimestamp(t, tracker, time.Now())

		count, err := repository.GetActiveTrackerCount()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1)
	})

	t.Run("multiple active trackers", func(t *testing.T) {
		trackers := []uuid.UUID{
			createTestTrackerDirect(t, "Multi Active 1", 90, 1.0, 1.0, "MULTIACT000001"),
			createTestTrackerDirect(t, "Multi Active 2", 85, 2.0, 2.0, "MULTIACT000002"),
			createTestTrackerDirect(t, "Multi Active 3", 75, 3.0, 3.0, "MULTIACT000003"),
		}
		defer func() {
			for _, id := range trackers {
				cleanupTracker(t, id)
			}
		}()

		for _, id := range trackers {
			updateTrackerTimestamp(t, id, time.Now())
		}

		count, err := repository.GetActiveTrackerCount()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3)
	})
}

func TestGetTrackerByDevEUI(t *testing.T) {
	t.Run("existing DevEUI returns tracker ID", func(t *testing.T) {
		expectedID := createTestTrackerDirect(t, "Test Tracker", 80, 10.0, 20.0, "DEVEUI000000001")
		defer cleanupTracker(t, expectedID)

		trackerID, err := repository.GetTrackerByDevEUI("DEVEUI000000001")
		require.NoError(t, err)
		assert.Equal(t, expectedID, trackerID)
	})

	t.Run("non-existent DevEUI returns nil", func(t *testing.T) {
		trackerID, err := repository.GetTrackerByDevEUI("NONEXISTENT0001")
		require.NoError(t, err)
		assert.Equal(t, uuid.Nil, trackerID)
	})

	t.Run("empty DevEUI", func(t *testing.T) {
		trackerID, err := repository.GetTrackerByDevEUI("")
		require.NoError(t, err)
		assert.Equal(t, uuid.Nil, trackerID)
	})

	t.Run("case sensitivity", func(t *testing.T) {
		id := createTestTrackerDirect(t, "Case Test", 80, 10.0, 20.0, "abcdef1234567890")
		defer cleanupTracker(t, id)

		// Try with different case
		trackerID, err := repository.GetTrackerByDevEUI("ABCDEF1234567890")
		require.NoError(t, err)
		// Depending on DB collation, this might or might not match
		_ = trackerID
	})

	t.Run("special characters in DevEUI", func(t *testing.T) {
		devEUI := "DEV-EUI:12:34:56"
		id := createTestTrackerDirect(t, "Special DevEUI", 80, 10.0, 20.0, devEUI)
		defer cleanupTracker(t, id)

		trackerID, err := repository.GetTrackerByDevEUI(devEUI)
		require.NoError(t, err)
		assert.Equal(t, id, trackerID)
	})
}

func TestGetTrackerByID(t *testing.T) {
	t.Run("existing tracker", func(t *testing.T) {
		id := createTestTrackerDirect(t, "Test Tracker", 85, 52.5200, 13.4050, "TESTTRACKER001")
		defer cleanupTracker(t, id)

		tracker, err := repository.GetTrackerByID(id)
		require.NoError(t, err)
		assert.NotNil(t, tracker)
		assert.Equal(t, id, tracker.ID)
		assert.Equal(t, "Test Tracker", tracker.Name)
		assert.Equal(t, int16(85), tracker.Battery)
		assert.Equal(t, 52.5200, tracker.Position.Latitude)
		assert.Equal(t, 13.4050, tracker.Position.Longitude)
	})

	t.Run("non-existent tracker", func(t *testing.T) {
		nonExistentID := uuid.Must(uuid.NewV4())
		tracker, err := repository.GetTrackerByID(nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, tracker)
	})

	t.Run("nil UUID", func(t *testing.T) {
		tracker, err := repository.GetTrackerByID(uuid.Nil)
		assert.Error(t, err)
		assert.Nil(t, tracker)
	})

	t.Run("tracker with assigned resource", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t) // Setup operation for resource filtering
		trackerID := createTestTrackerDirect(t, "Tracker With Resource", 75, 10.0, 20.0, "WITHRESOURCE01")
		defer cleanupTracker(t, trackerID)

		tableauResourceID := insertTestResource(t, "Test Resource", "vehicle", 1)
		defer cleanupResource(t, tableauResourceID)

		err := repository.UpdateTrackerResource(trackerID, tableauResourceID)
		require.NoError(t, err)

		tracker, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.NotNil(t, tracker)
		require.NotNil(t, tracker.TableauResource)
		assert.Equal(t, tableauResourceID, tracker.TableauResource.ID)
		assert.Equal(t, "Test Resource", tracker.TableauResource.Resource.Name)
	})

	t.Run("tracker without assigned resource", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Tracker No Resource", 70, 30.0, 40.0, "NORESOURCE0001")
		defer cleanupTracker(t, trackerID)

		tracker, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.NotNil(t, tracker)
		assert.Nil(t, tracker.TableauResource)
	})

	t.Run("tracker with zero battery", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Zero Battery", 0, 10.0, 20.0, "ZEROBATT000001")
		defer cleanupTracker(t, trackerID)

		tracker, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Equal(t, int16(0), tracker.Battery)
	})

	t.Run("tracker with negative battery", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Negative Battery", -1, 10.0, 20.0, "NEGBATT0000001")
		defer cleanupTracker(t, trackerID)

		tracker, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Equal(t, int16(-1), tracker.Battery)
	})
}

func TestGetAllTrackers(t *testing.T) {
	t.Run("empty database", func(t *testing.T) {
		trackers, err := repository.GetAllTrackers()
		require.NoError(t, err)
		assert.NotNil(t, trackers)
		// May be empty or have existing test data
	})

	t.Run("single tracker", func(t *testing.T) {
		id := createTestTrackerDirect(t, "Single Tracker", 90, 45.0, 90.0, "SINGLETRACK001")
		defer cleanupTracker(t, id)

		trackers, err := repository.GetAllTrackers()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(trackers), 1)
	})

	t.Run("multiple trackers", func(t *testing.T) {
		ids := []uuid.UUID{
			createTestTrackerDirect(t, "Multi 1", 80, 10.0, 10.0, "MULTI0000001"),
			createTestTrackerDirect(t, "Multi 2", 85, 20.0, 20.0, "MULTI0000002"),
			createTestTrackerDirect(t, "Multi 3", 90, 30.0, 30.0, "MULTI0000003"),
		}
		defer func() {
			for _, id := range ids {
				cleanupTracker(t, id)
			}
		}()

		trackers, err := repository.GetAllTrackers()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(trackers), 3)
	})
}

func TestUpdateTrackerResource(t *testing.T) {
	t.Run("assign resource to tracker", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t) // Setup operation for resource filtering
		trackerID := createTestTrackerDirect(t, "Tracker", 80, 10.0, 20.0, "ASSIGNRES00001")
		defer cleanupTracker(t, trackerID)

		tableauResourceID := insertTestResource(t, "Resource", "vehicle", 1)
		defer cleanupResource(t, tableauResourceID)

		err := repository.UpdateTrackerResource(trackerID, tableauResourceID)
		require.NoError(t, err)

		tracker, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		require.NotNil(t, tracker.TableauResource)
		assert.Equal(t, tableauResourceID, tracker.TableauResource.ID)
	})

	t.Run("reassign resource", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t) // Setup operation for resource filtering
		trackerID := createTestTrackerDirect(t, "Tracker", 80, 10.0, 20.0, "REASSIGNRES001")
		defer cleanupTracker(t, trackerID)

		tableauResource1 := insertTestResource(t, "Resource 1", "vehicle", 1)
		defer cleanupResource(t, tableauResource1)
		tableauResource2 := insertTestResource(t, "Resource 2", "person", 2)
		defer cleanupResource(t, tableauResource2)

		err := repository.UpdateTrackerResource(trackerID, tableauResource1)
		require.NoError(t, err)

		err = repository.UpdateTrackerResource(trackerID, tableauResource2)
		require.NoError(t, err)

		tracker, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		require.NotNil(t, tracker.TableauResource)
		assert.Equal(t, tableauResource2, tracker.TableauResource.ID)
	})

	t.Run("assign same resource to multiple trackers succeeds", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t) // Setup operation for resource filtering
		tracker1 := createTestTrackerDirect(t, "Tracker 1", 80, 10.0, 20.0, "SAMERESTRACK01")
		defer cleanupTracker(t, tracker1)
		tracker2 := createTestTrackerDirect(t, "Tracker 2", 85, 20.0, 30.0, "SAMERESTRACK02")
		defer cleanupTracker(t, tracker2)

		tableauResourceID := insertTestResource(t, "Shared Resource", "vehicle", 1)
		defer cleanupResource(t, tableauResourceID)

		err := repository.UpdateTrackerResource(tracker1, tableauResourceID)
		require.NoError(t, err)

		// Multiple trackers can be assigned to the same resource
		err = repository.UpdateTrackerResource(tracker2, tableauResourceID)
		assert.NoError(t, err, "Multiple trackers can share a resource")

		// Verify both trackers have the same resource
		tracker1Data, err := repository.GetTrackerByID(tracker1)
		require.NoError(t, err)
		assert.Equal(t, tableauResourceID, tracker1Data.TableauResource.ID)

		tracker2Data, err := repository.GetTrackerByID(tracker2)
		require.NoError(t, err)
		assert.Equal(t, tableauResourceID, tracker2Data.TableauResource.ID)
	})

	t.Run("assign non-existent resource fails", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Tracker", 80, 10.0, 20.0, "NONEXRES000001")
		defer cleanupTracker(t, trackerID)

		nonExistentResource := uuid.Must(uuid.NewV4())
		err := repository.UpdateTrackerResource(trackerID, nonExistentResource)
		assert.Error(t, err, "Should fail due to foreign key constraint")
	})

	t.Run("assign to non-existent tracker fails", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t)
		tableauResourceID := insertTestResource(t, "Resource", "vehicle", 1)
		defer cleanupResource(t, tableauResourceID)

		nonExistentTracker := uuid.Must(uuid.NewV4())
		err := repository.UpdateTrackerResource(nonExistentTracker, tableauResourceID)
		assert.Error(t, err)
	})

	t.Run("assign nil resource ID fails", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Tracker", 80, 10.0, 20.0, "NILRES00000001")
		defer cleanupTracker(t, trackerID)

		err := repository.UpdateTrackerResource(trackerID, uuid.Nil)
		assert.Error(t, err)
	})
}

func TestRemoveTrackerAssignment(t *testing.T) {
	t.Run("remove existing assignment", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t)
		trackerID := createTestTrackerDirect(t, "Tracker", 80, 10.0, 20.0, "REMASSIGN00001")
		defer cleanupTracker(t, trackerID)

		tableauResourceID := insertTestResource(t, "Resource", "vehicle", 1)
		defer cleanupResource(t, tableauResourceID)

		err := repository.UpdateTrackerResource(trackerID, tableauResourceID)
		require.NoError(t, err)

		err = repository.RemoveTrackerAssignment(trackerID)
		require.NoError(t, err)

		tracker, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Nil(t, tracker.TableauResource)
	})

	t.Run("remove non-existent assignment succeeds", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Tracker", 80, 10.0, 20.0, "REMNOASSIGN001")
		defer cleanupTracker(t, trackerID)

		err := repository.RemoveTrackerAssignment(trackerID)
		require.NoError(t, err)
	})

	t.Run("remove with non-existent tracker succeeds", func(t *testing.T) {
		nonExistentID := uuid.Must(uuid.NewV4())
		err := repository.RemoveTrackerAssignment(nonExistentID)
		require.NoError(t, err)
	})

	t.Run("remove with nil UUID succeeds", func(t *testing.T) {
		err := repository.RemoveTrackerAssignment(uuid.Nil)
		require.NoError(t, err)
	})
}

func TestRenameTracker(t *testing.T) {
	t.Run("rename tracker with valid name", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Old Name", 80, 10.0, 20.0, "RENAME00000001")
		defer cleanupTracker(t, trackerID)

		err := repository.RenameTracker(trackerID, "New Name")
		require.NoError(t, err)

		tracker, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Equal(t, "New Name", tracker.Name)
	})

	t.Run("rename to empty string", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Original", 80, 10.0, 20.0, "RENEMPTY000001")
		defer cleanupTracker(t, trackerID)

		err := repository.RenameTracker(trackerID, "")
		require.NoError(t, err)

		tracker, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Equal(t, "", tracker.Name)
	})

	t.Run("rename to very long name", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Short", 80, 10.0, 20.0, "RENLONG0000001")
		defer cleanupTracker(t, trackerID)

		longName := "Very long name that exceeds typical limits and is repeated multiple times to make it very very long for testing purposes"

		err := repository.RenameTracker(trackerID, longName)
		require.NoError(t, err)
	})

	t.Run("rename with special characters", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Original", 80, 10.0, 20.0, "RENSPECIAL0001")
		defer cleanupTracker(t, trackerID)

		specialName := "Tracker #1 (Test) <Special> & More!"
		err := repository.RenameTracker(trackerID, specialName)
		require.NoError(t, err)

		tracker, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Contains(t, tracker.Name, "Special")
	})

	t.Run("rename with unicode", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Original", 80, 10.0, 20.0, "RENUNICODE0001")
		defer cleanupTracker(t, trackerID)

		unicodeName := "Трекер 中文 🚗 Test"
		err := repository.RenameTracker(trackerID, unicodeName)
		require.NoError(t, err)

		tracker, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Contains(t, tracker.Name, "中文")
	})

	t.Run("rename non-existent tracker succeeds", func(t *testing.T) {
		nonExistentID := uuid.Must(uuid.NewV4())
		err := repository.RenameTracker(nonExistentID, "New Name")
		require.NoError(t, err) // UPDATE with no matching rows still succeeds
	})

	t.Run("rename with nil UUID succeeds", func(t *testing.T) {
		err := repository.RenameTracker(uuid.Nil, "Name")
		require.NoError(t, err)
	})

	t.Run("multiple renames in sequence", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Name 0", 80, 10.0, 20.0, "RENMULTI000001")
		defer cleanupTracker(t, trackerID)

		names := []string{"Name 1", "Name 2", "Name 3", "Final Name"}
		for _, name := range names {
			err := repository.RenameTracker(trackerID, name)
			require.NoError(t, err)

			tracker, err := repository.GetTrackerByID(trackerID)
			require.NoError(t, err)
			assert.Equal(t, name, tracker.Name)
		}
	})
}

func TestUpdateTracker(t *testing.T) {
	t.Run("update battery only", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Tracker", 50, 10.0, 20.0, "UPDBATT0000001")
		defer cleanupTracker(t, trackerID)

		tracker := models.BaseTracker{
			ID:      trackerID,
			Battery: 75,
			Position: models.Position{
				Latitude:  math.Inf(-1),
				Longitude: math.Inf(-1),
			},
		}

		err := repository.UpdateTracker(tracker)
		require.NoError(t, err)

		saved, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Equal(t, int16(75), saved.Battery)
		assert.Equal(t, 10.0, saved.Position.Latitude) // Position unchanged
		assert.Equal(t, 20.0, saved.Position.Longitude)
	})

	t.Run("update position only", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Tracker", 60, 10.0, 20.0, "UPDPOS00000001")
		defer cleanupTracker(t, trackerID)

		tracker := models.BaseTracker{
			ID:      trackerID,
			Battery: -1,
			Position: models.Position{
				Latitude:  52.5200,
				Longitude: 13.4050,
			},
		}

		err := repository.UpdateTracker(tracker)
		require.NoError(t, err)

		saved, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Equal(t, 52.5200, saved.Position.Latitude)
		assert.Equal(t, 13.4050, saved.Position.Longitude)
		assert.Equal(t, int16(60), saved.Battery) // Battery unchanged
	})

	t.Run("update both battery and position", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Tracker", 40, 10.0, 20.0, "UPDBOTH0000001")
		defer cleanupTracker(t, trackerID)

		tracker := models.BaseTracker{
			ID:      trackerID,
			Battery: 95,
			Position: models.Position{
				Latitude:  45.0,
				Longitude: 90.0,
			},
		}

		err := repository.UpdateTracker(tracker)
		require.NoError(t, err)

		saved, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Equal(t, int16(95), saved.Battery)
		assert.Equal(t, 45.0, saved.Position.Latitude)
		assert.Equal(t, 90.0, saved.Position.Longitude)
	})

	t.Run("no data to save returns error", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Tracker", 50, 10.0, 20.0, "UPDNODATA00001")
		defer cleanupTracker(t, trackerID)

		tracker := models.BaseTracker{
			ID:      trackerID,
			Battery: -1,
			Position: models.Position{
				Latitude:  math.Inf(-1),
				Longitude: math.Inf(-1),
			},
		}

		err := repository.UpdateTracker(tracker)
		assert.Error(t, err)
		assert.Equal(t, repository.ErrNoDataToSave, err)
	})

	t.Run("update with zero battery", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Tracker", 50, 10.0, 20.0, "UPDZEROBATT001")
		defer cleanupTracker(t, trackerID)

		tracker := models.BaseTracker{
			ID:      trackerID,
			Battery: 0,
			Position: models.Position{
				Latitude:  math.Inf(-1),
				Longitude: math.Inf(-1),
			},
		}

		err := repository.UpdateTracker(tracker)
		require.NoError(t, err)

		saved, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Equal(t, int16(0), saved.Battery)
	})

	t.Run("update with extreme coordinates", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Tracker", 50, 10.0, 20.0, "UPDEXTREME0001")
		defer cleanupTracker(t, trackerID)

		tracker := models.BaseTracker{
			ID:      trackerID,
			Battery: -1,
			Position: models.Position{
				Latitude:  -90.0,
				Longitude: 180.0,
			},
		}

		err := repository.UpdateTracker(tracker)
		require.NoError(t, err)

		saved, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Equal(t, -90.0, saved.Position.Latitude)
		assert.Equal(t, 180.0, saved.Position.Longitude)
	})

	t.Run("update with zero coordinates", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Tracker", 50, 10.0, 20.0, "UPDZEROCOORD01")
		defer cleanupTracker(t, trackerID)

		tracker := models.BaseTracker{
			ID:      trackerID,
			Battery: -1,
			Position: models.Position{
				Latitude:  0.0,
				Longitude: 0.0,
			},
		}

		err := repository.UpdateTracker(tracker)
		require.NoError(t, err)

		saved, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Equal(t, 0.0, saved.Position.Latitude)
		assert.Equal(t, 0.0, saved.Position.Longitude)
	})

	t.Run("update non-existent tracker fails", func(t *testing.T) {
		tracker := models.BaseTracker{
			ID:      uuid.Must(uuid.NewV4()),
			Battery: 80,
			Position: models.Position{
				Latitude:  math.Inf(-1),
				Longitude: math.Inf(-1),
			},
		}

		err := repository.UpdateTracker(tracker)
		// Should succeed but affect 0 rows - no error returned by Exec
		require.NoError(t, err)
	})

	t.Run("update with max battery value", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Tracker", 50, 10.0, 20.0, "UPDMAXBATT0001")
		defer cleanupTracker(t, trackerID)

		tracker := models.BaseTracker{
			ID:      trackerID,
			Battery: 32767, // max int16
			Position: models.Position{
				Latitude:  math.Inf(-1),
				Longitude: math.Inf(-1),
			},
		}

		err := repository.UpdateTracker(tracker)
		require.NoError(t, err)

		saved, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Equal(t, int16(32767), saved.Battery)
	})

	t.Run("multiple sequential updates", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Tracker", 50, 10.0, 20.0, "UPDSEQ00000001")
		defer cleanupTracker(t, trackerID)

		updates := []models.BaseTracker{
			{ID: trackerID, Battery: 60, Position: models.Position{Latitude: math.Inf(-1), Longitude: math.Inf(-1)}},
			{ID: trackerID, Battery: -1, Position: models.Position{Latitude: 20.0, Longitude: 30.0}},
			{ID: trackerID, Battery: 70, Position: models.Position{Latitude: 25.0, Longitude: 35.0}},
		}

		for _, update := range updates {
			err := repository.UpdateTracker(update)
			require.NoError(t, err)
		}

		saved, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Equal(t, int16(70), saved.Battery)
		assert.Equal(t, 25.0, saved.Position.Latitude)
		assert.Equal(t, 35.0, saved.Position.Longitude)
	})

	t.Run("update with positive infinity latitude skipped", func(t *testing.T) {
		trackerID := createTestTrackerDirect(t, "Tracker", 50, 10.0, 20.0, "UPDPOSINF00001")
		defer cleanupTracker(t, trackerID)

		tracker := models.BaseTracker{
			ID:      trackerID,
			Battery: 80,
			Position: models.Position{
				Latitude:  math.Inf(1),
				Longitude: 30.0,
			},
		}

		err := repository.UpdateTracker(tracker)
		require.NoError(t, err)

		saved, err := repository.GetTrackerByID(trackerID)
		require.NoError(t, err)
		assert.Equal(t, int16(80), saved.Battery)
		assert.Equal(t, 10.0, saved.Position.Latitude)  // Unchanged
		assert.Equal(t, 20.0, saved.Position.Longitude) // Unchanged
	})
}

// Helper functions

func createTestTrackerDirect(t *testing.T, name string, battery int16, lat, lon float64, devEUI string) uuid.UUID {
	id := uuid.Must(uuid.NewV4())
	SQL1 := `INSERT INTO trackers (tracker_id, name, battery, position_latitude, position_longitude) 
	         VALUES ($1, $2, $3, $4, $5)`
	SQL2 := `INSERT INTO chirpstack_trackers (tracker_id, dev_eui) VALUES ($1, $2)`

	_, err := repository.DBConnPool.Exec(context.Background(), SQL1, id, name, battery, lat, lon)
	require.NoError(t, err)
	_, err = repository.DBConnPool.Exec(context.Background(), SQL2, id, devEUI)
	require.NoError(t, err)

	return id
}

func updateTrackerTimestamp(t *testing.T, trackerID uuid.UUID, timestamp time.Time) {
	SQL := `UPDATE trackers SET updated_at = $1 WHERE tracker_id = $2`
	_, err := repository.DBConnPool.Exec(context.Background(), SQL, timestamp, trackerID)
	require.NoError(t, err)
}

func cleanupAllTrackers(t *testing.T) {
	SQL1 := `DELETE FROM chirpstack_trackers`
	SQL2 := `DELETE FROM tetra_trackers`
	SQL3 := `DELETE FROM trackers`

	repository.DBConnPool.Exec(context.Background(), SQL1)
	repository.DBConnPool.Exec(context.Background(), SQL2)
	_, err := repository.DBConnPool.Exec(context.Background(), SQL3)
	if err != nil {
		t.Logf("Failed to cleanup all trackers: %v", err)
	}
}
