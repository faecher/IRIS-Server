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

// Helper function to insert test resources
func insertTestResource(t *testing.T, name, resourceType string, status uint16) uuid.UUID {
	resourceID := uuid.Must(uuid.NewV4())
	tableauResourceID := uuid.Must(uuid.NewV4())

	// Get current operation ID from mcp_config (or use a default for testing)
	var operationID uuid.UUID
	err := repository.DBConnPool.QueryRow(context.Background(),
		`SELECT COALESCE(operation_id, '00000000-0000-0000-0000-000000000000'::UUID) FROM mcp_config WHERE id = 1`).Scan(&operationID)
	if err != nil {
		// If no mcp_config exists, use a test operation ID
		operationID = uuid.Must(uuid.NewV4())
	}

	// Insert base resource
	SQL1 := `INSERT INTO resources (resource_id, name, type) VALUES ($1, $2, $3)`
	_, err = repository.DBConnPool.Exec(context.Background(), SQL1, resourceID, name, resourceType)
	require.NoError(t, err)

	// Insert tableau_resource
	SQL2 := `INSERT INTO tableau_resources (tableau_resource_id, resource_id, operation_id, status) VALUES ($1, $2, $3, $4)`
	_, err = repository.DBConnPool.Exec(context.Background(), SQL2, tableauResourceID, resourceID, operationID, status)
	require.NoError(t, err)

	return tableauResourceID
}

// Helper to setup MCP config with siteplan and operation
func setupMCPConfigWithSiteplan(t *testing.T) uuid.UUID {
	operationID := uuid.Must(uuid.NewV4())
	config := models.MCPConfig{
		URL:     "https://test.com",
		APIKey:  "test",
		Enabled: true,
	}
	err := repository.UpdateMCPConfig(config)
	require.NoError(t, err)

	// Set operation ID
	err = repository.UpdateMCPOperation(operationID)
	require.NoError(t, err)

	siteplanID := uuid.Must(uuid.NewV4())
	err = repository.UpdateMCPSiteplan(siteplanID)
	require.NoError(t, err)

	return siteplanID
}

// Helper to get base resource_id from tableau_resource_id
func getBaseResourceID(t *testing.T, tableauResourceID uuid.UUID) uuid.UUID {
	var resourceID uuid.UUID
	SQL := `SELECT resource_id FROM tableau_resources WHERE tableau_resource_id = $1`
	err := repository.DBConnPool.QueryRow(context.Background(), SQL, tableauResourceID).Scan(&resourceID)
	require.NoError(t, err)
	return resourceID
}

func TestGetAllResources(t *testing.T) {
	t.Run("empty database returns empty list", func(t *testing.T) {
		resources, err := repository.GetAllResources()
		require.NoError(t, err)
		assert.Empty(t, resources)
	})

	t.Run("returns single resource", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t) // Setup operation for filtering
		tableauResourceID := insertTestResource(t, "Test Resource", "vehicle", 1)
		defer cleanupResource(t, tableauResourceID)

		resources, err := repository.GetAllResources()
		require.NoError(t, err)
		assert.Len(t, resources, 1)
		assert.Equal(t, tableauResourceID, resources[0].ID)
		assert.Equal(t, "Test Resource", resources[0].Resource.Name)
		assert.Equal(t, "vehicle", resources[0].Resource.Type)
		assert.Equal(t, uint16(1), resources[0].Status)
	})

	t.Run("returns multiple resources", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t) // Setup operation for filtering
		ids := []uuid.UUID{
			insertTestResource(t, "Resource 1", "vehicle", 1),
			insertTestResource(t, "Resource 2", "person", 2),
			insertTestResource(t, "Resource 3", "equipment", 3),
		}
		defer func() {
			for _, id := range ids {
				cleanupResource(t, id)
			}
		}()

		resources, err := repository.GetAllResources()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(resources), 3)

		// Verify all inserted resources are present
		foundIDs := make(map[uuid.UUID]bool)
		for _, resource := range resources {
			foundIDs[resource.ID] = true
		}
		for _, id := range ids {
			assert.True(t, foundIDs[id], "Resource %s should be in results", id)
		}
	})

	t.Run("various resource types", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t)
		types := []string{"vehicle", "person", "equipment", "drone", "robot"}
		ids := make([]uuid.UUID, len(types))
		for i, resourceType := range types {
			ids[i] = insertTestResource(t, "Resource "+resourceType, resourceType, uint16(i+1))
		}
		defer func() {
			for _, id := range ids {
				cleanupResource(t, id)
			}
		}()

		resources, err := repository.GetAllResources()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(resources), len(types))
	})

	t.Run("various status values", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t)
		statuses := []uint16{0, 1, 100, 255, 32767}
		ids := make([]uuid.UUID, len(statuses))
		for i, status := range statuses {
			ids[i] = insertTestResource(t, "Resource", "vehicle", status)
		}
		defer func() {
			for _, id := range ids {
				cleanupResource(t, id)
			}
		}()

		resources, err := repository.GetAllResources()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(resources), len(statuses))
	})

	t.Run("long resource names", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t)
		longName := "A very long resource name that exceeds typical length limits and is repeated multiple times to make it very long"
		id := insertTestResource(t, longName, "vehicle", 1)
		defer cleanupResource(t, id)

		resources, err := repository.GetAllResources()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(resources), 1)
	})

	t.Run("special characters in names", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t)
		specialNames := []string{
			"Resource with spaces",
			"Resource-with-dashes",
			"Resource_with_underscores",
			"Resource (with) parentheses",
			"Ресурс с кириллицей",
			"资源中文",
		}
		ids := make([]uuid.UUID, len(specialNames))
		for i, name := range specialNames {
			ids[i] = insertTestResource(t, name, "vehicle", 1)
		}
		defer func() {
			for _, id := range ids {
				cleanupResource(t, id)
			}
		}()

		resources, err := repository.GetAllResources()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(resources), len(specialNames))
	})
}

func TestGetResourceByID(t *testing.T) {
	t.Run("valid resource ID", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t)
		tableauResourceID := insertTestResource(t, "Test Vehicle", "vehicle", 1)
		defer cleanupResource(t, tableauResourceID)

		resource, err := repository.GetResourceByID(tableauResourceID)
		require.NoError(t, err)
		assert.NotNil(t, resource)
		assert.Equal(t, tableauResourceID, resource.ID)
		assert.Equal(t, "Test Vehicle", resource.Resource.Name)
		assert.Equal(t, "vehicle", resource.Resource.Type)
		assert.Equal(t, uint16(1), resource.Status)
	})

	t.Run("non-existent resource ID", func(t *testing.T) {
		nonExistentID := uuid.Must(uuid.NewV4())
		resource, err := repository.GetResourceByID(nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, resource)
	})

	t.Run("nil UUID", func(t *testing.T) {
		resource, err := repository.GetResourceByID(uuid.Nil)
		assert.Error(t, err)
		assert.Nil(t, resource)
	})

	t.Run("multiple resources with different IDs", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t)
		ids := []uuid.UUID{
			insertTestResource(t, "Resource 1", "vehicle", 1),
			insertTestResource(t, "Resource 2", "person", 2),
			insertTestResource(t, "Resource 3", "equipment", 3),
		}
		defer func() {
			for _, id := range ids {
				cleanupResource(t, id)
			}
		}()

		// Verify each can be retrieved individually
		for i, id := range ids {
			resource, err := repository.GetResourceByID(id)
			require.NoError(t, err)
			assert.Equal(t, id, resource.ID)
			assert.Contains(t, resource.Resource.Name, "Resource")
			assert.Equal(t, uint16(i+1), resource.Status)
		}
	})

	t.Run("resource with zero status", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t)
		id := insertTestResource(t, "Zero Status Resource", "vehicle", 0)
		defer cleanupResource(t, id)

		resource, err := repository.GetResourceByID(id)
		require.NoError(t, err)
		assert.Equal(t, uint16(0), resource.Status)
	})

	t.Run("resource with max status", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t)
		id := insertTestResource(t, "Max Status Resource", "vehicle", 32767)
		defer cleanupResource(t, id)

		resource, err := repository.GetResourceByID(id)
		require.NoError(t, err)
		assert.Equal(t, uint16(32767), resource.Status)
	})
}

func TestUpdateMarkerIDForResource(t *testing.T) {
	t.Run("insert marker for resource", func(t *testing.T) {
		tableauResourceID := insertTestResource(t, "Test Resource", "vehicle", 1)
		defer cleanupResource(t, tableauResourceID)
		baseResourceID := getBaseResourceID(t, tableauResourceID)

		siteplanID := setupMCPConfigWithSiteplan(t)
		markerID := uuid.Must(uuid.NewV4())

		err := repository.UpdateMarkerIDForResource(baseResourceID, markerID)
		require.NoError(t, err)

		// Verify insertion
		marker, err := repository.GetResourceMarker(tableauResourceID)
		require.NoError(t, err)
		assert.Equal(t, markerID, marker.MarkerID)
		assert.Equal(t, baseResourceID, marker.ResourceID)
		assert.Equal(t, siteplanID, marker.SiteplanID)
	})

	t.Run("insert marker for non-existent resource fails", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t)
		nonExistentResourceID := uuid.Must(uuid.NewV4())
		markerID := uuid.Must(uuid.NewV4())

		err := repository.UpdateMarkerIDForResource(nonExistentResourceID, markerID)
		assert.Error(t, err) // Should fail due to foreign key constraint
	})

	t.Run("insert with nil marker ID", func(t *testing.T) {
		tableauResourceID := insertTestResource(t, "Test Resource", "vehicle", 1)
		defer cleanupResource(t, tableauResourceID)
		baseResourceID := getBaseResourceID(t, tableauResourceID)

		setupMCPConfigWithSiteplan(t)

		err := repository.UpdateMarkerIDForResource(baseResourceID, uuid.Nil)
		require.NoError(t, err)
	})

	t.Run("insert with nil resource ID fails", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t)
		markerID := uuid.Must(uuid.NewV4())

		err := repository.UpdateMarkerIDForResource(uuid.Nil, markerID)
		assert.Error(t, err)
	})

	t.Run("multiple markers for same resource different siteplans", func(t *testing.T) {
		tableauResourceID := insertTestResource(t, "Test Resource", "vehicle", 1)
		defer cleanupResource(t, tableauResourceID)
		baseResourceID := getBaseResourceID(t, tableauResourceID)

		// First siteplan
		siteplan1 := setupMCPConfigWithSiteplan(t)
		marker1 := uuid.Must(uuid.NewV4())
		err := repository.UpdateMarkerIDForResource(baseResourceID, marker1)
		require.NoError(t, err)

		// Change to second siteplan
		siteplan2 := uuid.Must(uuid.NewV4())
		err = repository.UpdateMCPSiteplan(siteplan2)
		require.NoError(t, err)
		marker2 := uuid.Must(uuid.NewV4())
		err = repository.UpdateMarkerIDForResource(baseResourceID, marker2)
		require.NoError(t, err)

		// Verify both exist
		err = repository.UpdateMCPSiteplan(siteplan1)
		require.NoError(t, err)
		result1, err := repository.GetResourceMarker(tableauResourceID)
		require.NoError(t, err)
		assert.Equal(t, marker1, result1.MarkerID)

		err = repository.UpdateMCPSiteplan(siteplan2)
		require.NoError(t, err)
		result2, err := repository.GetResourceMarker(tableauResourceID)
		require.NoError(t, err)
		assert.Equal(t, marker2, result2.MarkerID)
	})

	t.Run("upsert updates existing marker", func(t *testing.T) {
		tableauResourceID := insertTestResource(t, "Test Resource", "vehicle", 1)
		defer cleanupResource(t, tableauResourceID)
		baseResourceID := getBaseResourceID(t, tableauResourceID)

		setupMCPConfigWithSiteplan(t)
		markerID1 := uuid.Must(uuid.NewV4())
		markerID2 := uuid.Must(uuid.NewV4())

		// Insert first marker
		err := repository.UpdateMarkerIDForResource(baseResourceID, markerID1)
		require.NoError(t, err)

		// Update with new marker ID (upsert should succeed)
		err = repository.UpdateMarkerIDForResource(baseResourceID, markerID2)
		require.NoError(t, err)

		// Verify the marker was updated
		result, err := repository.GetResourceMarker(tableauResourceID)
		require.NoError(t, err)
		assert.Equal(t, markerID2, result.MarkerID)
	})
}

func TestGetResourceMarker(t *testing.T) {
	t.Run("get existing marker", func(t *testing.T) {
		tableauResourceID := insertTestResource(t, "Test Resource", "vehicle", 1)
		defer cleanupResource(t, tableauResourceID)
		baseResourceID := getBaseResourceID(t, tableauResourceID)

		siteplanID := setupMCPConfigWithSiteplan(t)
		markerID := uuid.Must(uuid.NewV4())

		err := repository.UpdateMarkerIDForResource(baseResourceID, markerID)
		require.NoError(t, err)

		marker, err := repository.GetResourceMarker(tableauResourceID)
		require.NoError(t, err)
		assert.Equal(t, markerID, marker.MarkerID)
		assert.Equal(t, baseResourceID, marker.ResourceID)
		assert.Equal(t, siteplanID, marker.SiteplanID)
	})

	t.Run("marker not found returns siteplan ID only", func(t *testing.T) {
		tableauResourceID := insertTestResource(t, "Test Resource", "vehicle", 1)
		defer cleanupResource(t, tableauResourceID)
		baseResourceID := getBaseResourceID(t, tableauResourceID)

		siteplanID := setupMCPConfigWithSiteplan(t)

		marker, err := repository.GetResourceMarker(tableauResourceID)
		require.NoError(t, err)
		assert.Equal(t, uuid.Nil, marker.MarkerID)
		assert.Equal(t, baseResourceID, marker.ResourceID)
		assert.Equal(t, siteplanID, marker.SiteplanID)
	})

	t.Run("non-existent resource", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t)
		nonExistentID := uuid.Must(uuid.NewV4())

		marker, err := repository.GetResourceMarker(nonExistentID)
		require.NoError(t, err)
		assert.Equal(t, uuid.Nil, marker.MarkerID)
	})

	t.Run("nil resource ID", func(t *testing.T) {
		setupMCPConfigWithSiteplan(t)

		marker, err := repository.GetResourceMarker(uuid.Nil)
		require.NoError(t, err)
		assert.Equal(t, uuid.Nil, marker.MarkerID)
	})

	t.Run("correct marker for current siteplan only", func(t *testing.T) {
		tableauResourceID := insertTestResource(t, "Test Resource", "vehicle", 1)
		defer cleanupResource(t, tableauResourceID)
		baseResourceID := getBaseResourceID(t, tableauResourceID)

		// Setup two siteplans with different markers
		siteplan1 := setupMCPConfigWithSiteplan(t)
		marker1 := uuid.Must(uuid.NewV4())
		err := repository.UpdateMarkerIDForResource(baseResourceID, marker1)
		require.NoError(t, err)

		siteplan2 := uuid.Must(uuid.NewV4())
		err = repository.UpdateMCPSiteplan(siteplan2)
		require.NoError(t, err)
		marker2 := uuid.Must(uuid.NewV4())
		err = repository.UpdateMarkerIDForResource(baseResourceID, marker2)
		require.NoError(t, err)

		// Switch back to siteplan1 and verify we get marker1
		err = repository.UpdateMCPSiteplan(siteplan1)
		require.NoError(t, err)
		result, err := repository.GetResourceMarker(tableauResourceID)
		require.NoError(t, err)
		assert.Equal(t, marker1, result.MarkerID)
		assert.Equal(t, siteplan1, result.SiteplanID)

		// Switch to siteplan2 and verify we get marker2
		err = repository.UpdateMCPSiteplan(siteplan2)
		require.NoError(t, err)
		result, err = repository.GetResourceMarker(tableauResourceID)
		require.NoError(t, err)
		assert.Equal(t, marker2, result.MarkerID)
		assert.Equal(t, siteplan2, result.SiteplanID)
	})

	t.Run("no mcp config set", func(t *testing.T) {
		// Clean up any existing MCP config from previous tests
		cleanupMCPConfig(t)

		tableauResourceID := insertTestResource(t, "Test Resource", "vehicle", 1)
		defer cleanupResource(t, tableauResourceID)

		// Don't setup MCP config
		_, err := repository.GetResourceMarker(tableauResourceID)
		// Should error because mcp_config doesn't exist or has NULL siteplan_id
		assert.Error(t, err)
	})
}

// Helper to cleanup test resources
func cleanupResource(t *testing.T, tableauResourceID uuid.UUID) {
	// First get the resource_id from tableau_resources
	var resourceID uuid.UUID
	SQL1 := `SELECT resource_id FROM tableau_resources WHERE tableau_resource_id = $1`
	err := repository.DBConnPool.QueryRow(context.Background(), SQL1, tableauResourceID).Scan(&resourceID)
	if err != nil {
		t.Logf("Failed to get resource_id for tableau_resource %s: %v", tableauResourceID, err)
		return
	}

	// Delete tableau_resource first (due to foreign key)
	SQL2 := `DELETE FROM tableau_resources WHERE tableau_resource_id = $1`
	_, err = repository.DBConnPool.Exec(context.Background(), SQL2, tableauResourceID)
	if err != nil {
		t.Logf("Failed to cleanup tableau_resource %s: %v", tableauResourceID, err)
	}

	// Then delete the base resource
	SQL3 := `DELETE FROM resources WHERE resource_id = $1`
	_, err = repository.DBConnPool.Exec(context.Background(), SQL3, resourceID)
	if err != nil {
		t.Logf("Failed to cleanup resource %s: %v", resourceID, err)
	}
}

// Helper to cleanup MCP config
func cleanupMCPConfig(t *testing.T) {
	SQL := `DELETE FROM mcp_config`
	_, err := repository.DBConnPool.Exec(context.Background(), SQL)
	if err != nil {
		t.Logf("Failed to cleanup MCP config: %v", err)
	}
}
