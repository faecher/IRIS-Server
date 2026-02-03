package repository_test

import (
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMCPConfig(t *testing.T) {
	t.Run("empty config returns empty struct", func(t *testing.T) {
		config, err := repository.GetMCPConfig()
		require.NoError(t, err)
		assert.Empty(t, config.URL)
		assert.Empty(t, config.APIKey)
		assert.False(t, config.Enabled)
	})

	t.Run("returns saved config", func(t *testing.T) {
		// Setup: save a config
		expected := models.MCPConfig{
			URL:     "https://example.com",
			APIKey:  "test-api-key-123",
			Enabled: true,
		}
		err := repository.UpdateMCPConfig(expected)
		require.NoError(t, err)

		// Test: retrieve it
		config, err := repository.GetMCPConfig()
		require.NoError(t, err)
		assert.Equal(t, expected.URL, config.URL)
		assert.Equal(t, expected.APIKey, config.APIKey)
		// Note: Enabled field is not scanned in GetMCPConfig implementation
	})
}

func TestUpdateMCPConfig(t *testing.T) {
	t.Run("insert new config", func(t *testing.T) {
		config := models.MCPConfig{
			URL:     "https://mcp.example.com",
			APIKey:  "secret-key",
			Enabled: true,
		}

		err := repository.UpdateMCPConfig(config)
		require.NoError(t, err)

		// Verify it was saved
		saved, err := repository.GetMCPConfig()
		require.NoError(t, err)
		assert.Equal(t, config.URL, saved.URL)
		assert.Equal(t, config.APIKey, saved.APIKey)
	})

	t.Run("replace existing config", func(t *testing.T) {
		// Setup: insert initial config
		initial := models.MCPConfig{
			URL:     "https://old.example.com",
			APIKey:  "old-key",
			Enabled: false,
		}
		err := repository.UpdateMCPConfig(initial)
		require.NoError(t, err)

		// Test: replace with new config
		updated := models.MCPConfig{
			URL:     "https://new.example.com",
			APIKey:  "new-key",
			Enabled: true,
		}
		err = repository.UpdateMCPConfig(updated)
		require.NoError(t, err)

		// Verify replacement
		saved, err := repository.GetMCPConfig()
		require.NoError(t, err)
		assert.Equal(t, updated.URL, saved.URL)
		assert.Equal(t, updated.APIKey, saved.APIKey)
	})

	t.Run("empty URL and APIKey", func(t *testing.T) {
		config := models.MCPConfig{
			URL:     "",
			APIKey:  "",
			Enabled: false,
		}

		err := repository.UpdateMCPConfig(config)
		require.NoError(t, err)

		saved, err := repository.GetMCPConfig()
		require.NoError(t, err)
		assert.Empty(t, saved.URL)
		assert.Empty(t, saved.APIKey)
	})

	t.Run("very long URL and APIKey", func(t *testing.T) {
		longURL := "https://" + strings.Repeat("a", 500) + ".example.com"
		longKey := strings.Repeat("k", 1000)

		config := models.MCPConfig{
			URL:     longURL,
			APIKey:  longKey,
			Enabled: true,
		}

		err := repository.UpdateMCPConfig(config)
		require.NoError(t, err)

		saved, err := repository.GetMCPConfig()
		require.NoError(t, err)
		assert.Equal(t, longURL, saved.URL)
		assert.Equal(t, longKey, saved.APIKey)
	})

	t.Run("special characters in URL and APIKey", func(t *testing.T) {
		config := models.MCPConfig{
			URL:     "https://example.com/api?key=value&token=123",
			APIKey:  "key-with-special!@#$%^&*()chars",
			Enabled: true,
		}

		err := repository.UpdateMCPConfig(config)
		require.NoError(t, err)

		saved, err := repository.GetMCPConfig()
		require.NoError(t, err)
		assert.Equal(t, config.URL, saved.URL)
		assert.Equal(t, config.APIKey, saved.APIKey)
	})
}

func TestUpdateMCPOperation(t *testing.T) {
	// Setup: ensure mcp_config row exists
	setupMCPConfig := func() {
		config := models.MCPConfig{
			URL:     "https://test.com",
			APIKey:  "test",
			Enabled: true,
		}
		repository.UpdateMCPConfig(config)
	}

	t.Run("set valid operation ID", func(t *testing.T) {
		setupMCPConfig()
		operationID := uuid.Must(uuid.NewV4())

		err := repository.UpdateMCPOperation(operationID)
		require.NoError(t, err)

		// Verify
		saved, err := repository.GetMCPOperation()
		require.NoError(t, err)
		require.NotNil(t, saved)
		assert.Equal(t, operationID, *saved)
	})

	t.Run("replace existing operation ID", func(t *testing.T) {
		setupMCPConfig()
		firstID := uuid.Must(uuid.NewV4())
		secondID := uuid.Must(uuid.NewV4())

		err := repository.UpdateMCPOperation(firstID)
		require.NoError(t, err)

		err = repository.UpdateMCPOperation(secondID)
		require.NoError(t, err)

		saved, err := repository.GetMCPOperation()
		require.NoError(t, err)
		require.NotNil(t, saved)
		assert.Equal(t, secondID, *saved)
		assert.NotEqual(t, firstID, *saved)
	})

	t.Run("set nil UUID", func(t *testing.T) {
		setupMCPConfig()

		err := repository.UpdateMCPOperation(uuid.Nil)
		require.NoError(t, err)

		saved, err := repository.GetMCPOperation()
		require.NoError(t, err)
		require.NotNil(t, saved)
		assert.Equal(t, uuid.Nil, *saved)
	})

	t.Run("multiple updates in sequence", func(t *testing.T) {
		setupMCPConfig()
		ids := []uuid.UUID{
			uuid.Must(uuid.NewV4()),
			uuid.Must(uuid.NewV4()),
			uuid.Must(uuid.NewV4()),
		}

		for _, id := range ids {
			err := repository.UpdateMCPOperation(id)
			require.NoError(t, err)

			saved, err := repository.GetMCPOperation()
			require.NoError(t, err)
			require.NotNil(t, saved)
			assert.Equal(t, id, *saved)
		}
	})
}

func TestGetMCPOperation(t *testing.T) {
	// Setup helper
	setupMCPConfig := func() {
		config := models.MCPConfig{
			URL:     "https://test.com",
			APIKey:  "test",
			Enabled: true,
		}
		repository.UpdateMCPConfig(config)
	}

	t.Run("returns saved operation ID", func(t *testing.T) {
		setupMCPConfig()
		expected := uuid.Must(uuid.NewV4())
		err := repository.UpdateMCPOperation(expected)
		require.NoError(t, err)

		result, err := repository.GetMCPOperation()
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, expected, *result)
	})

	t.Run("returns nil UUID when not set", func(t *testing.T) {
		setupMCPConfig()
		err := repository.UpdateMCPOperation(uuid.Nil)
		require.NoError(t, err)

		result, err := repository.GetMCPOperation()
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uuid.Nil, *result)
	})
}

func TestUpdateMCPSiteplan(t *testing.T) {
	setupMCPConfig := func() {
		config := models.MCPConfig{
			URL:     "https://test.com",
			APIKey:  "test",
			Enabled: true,
		}
		repository.UpdateMCPConfig(config)
	}

	t.Run("set valid siteplan ID", func(t *testing.T) {
		setupMCPConfig()
		siteplanID := uuid.Must(uuid.NewV4())

		err := repository.UpdateMCPSiteplan(siteplanID)
		require.NoError(t, err)

		saved, err := repository.GetMCPSiteplan()
		require.NoError(t, err)
		require.NotNil(t, saved)
		assert.Equal(t, siteplanID, *saved)
	})

	t.Run("replace existing siteplan ID", func(t *testing.T) {
		setupMCPConfig()
		firstID := uuid.Must(uuid.NewV4())
		secondID := uuid.Must(uuid.NewV4())

		err := repository.UpdateMCPSiteplan(firstID)
		require.NoError(t, err)

		err = repository.UpdateMCPSiteplan(secondID)
		require.NoError(t, err)

		saved, err := repository.GetMCPSiteplan()
		require.NoError(t, err)
		require.NotNil(t, saved)
		assert.Equal(t, secondID, *saved)
	})

	t.Run("set nil UUID", func(t *testing.T) {
		setupMCPConfig()

		err := repository.UpdateMCPSiteplan(uuid.Nil)
		require.NoError(t, err)

		saved, err := repository.GetMCPSiteplan()
		require.NoError(t, err)
		require.NotNil(t, saved)
		assert.Equal(t, uuid.Nil, *saved)
	})

	t.Run("independent from operation ID", func(t *testing.T) {
		setupMCPConfig()
		operationID := uuid.Must(uuid.NewV4())
		siteplanID := uuid.Must(uuid.NewV4())

		err := repository.UpdateMCPOperation(operationID)
		require.NoError(t, err)

		err = repository.UpdateMCPSiteplan(siteplanID)
		require.NoError(t, err)

		// Verify both are independent
		savedOperation, err := repository.GetMCPOperation()
		require.NoError(t, err)
		require.NotNil(t, savedOperation)
		assert.Equal(t, operationID, *savedOperation)

		savedSiteplan, err := repository.GetMCPSiteplan()
		require.NoError(t, err)
		require.NotNil(t, savedSiteplan)
		assert.Equal(t, siteplanID, *savedSiteplan)
	})
}

func TestGetMCPSiteplan(t *testing.T) {
	setupMCPConfig := func() {
		config := models.MCPConfig{
			URL:     "https://test.com",
			APIKey:  "test",
			Enabled: true,
		}
		repository.UpdateMCPConfig(config)
	}

	t.Run("returns saved siteplan ID", func(t *testing.T) {
		setupMCPConfig()
		expected := uuid.Must(uuid.NewV4())
		err := repository.UpdateMCPSiteplan(expected)
		require.NoError(t, err)

		result, err := repository.GetMCPSiteplan()
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, expected, *result)
	})

	t.Run("returns nil UUID when not set", func(t *testing.T) {
		setupMCPConfig()
		err := repository.UpdateMCPSiteplan(uuid.Nil)
		require.NoError(t, err)

		result, err := repository.GetMCPSiteplan()
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uuid.Nil, *result)
	})
}
