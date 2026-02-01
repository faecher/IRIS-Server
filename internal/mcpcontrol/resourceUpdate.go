package mcpcontrol

import (
	"IRIS-Server/internal/repository"
	"fmt"
)

func UpdateMCPResourcesInDB() error {
	resources, err := getMCPResources()
	if err != nil {
		return fmt.Errorf("failed to get MCP resources: %w", err)
	}

	for _, res := range resources {
		err := repository.UpsertResource(&res)
		if err != nil {
			return fmt.Errorf("failed to upsert resource %s: %w", res.ID.String(), err)
		}
	}

	return nil
}
