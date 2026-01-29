package repository

import (
	"IRIS-Server/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

// GetAllResources retrieves all resources from the database
func GetAllResources() ([]models.Resource, error) {
	SQL := `SELECT resource_id, name, type, status
	FROM resources`

	rows, err := DBConnPool.Query(context.Background(), SQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query all resources: %w", err)
	}
	defer rows.Close()

	resources, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Resource])
	if err != nil {
		return nil, fmt.Errorf("failed to collect resource rows: %w", err)
	}

	return resources, nil
}

// GetResourceByID retrieves a single resource by UUID
func GetResourceByID(resourceID uuid.UUID) (*models.Resource, error) {
	SQL := `SELECT resource_id, name, type, status 
	FROM resources
	WHERE resource_id = $1`

	var resource models.Resource
	err := DBConnPool.QueryRow(context.Background(), SQL, resourceID).Scan(
		&resource.ID,
		&resource.Name,
		&resource.Type,
		&resource.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query resource by ID: %w", err)
	}

	return &resource, nil
}

// UpdateMarkerIDForResource associates an MCP marker ID with a resource for the current siteplan
func UpdateMarkerIDForResource(resourceID, markerID uuid.UUID) error {
	SQL := `INSERT INTO resource_marker (marker_id, resource_id, siteplan_id) 
	VALUES ($1, $2, (SELECT siteplan_id FROM mcp_config WHERE id = 1))`

	_, err := DBConnPool.Exec(context.Background(), SQL, markerID, resourceID)
	if err != nil {
		return fmt.Errorf("failed to update marker ID for resource: %w", err)
	}

	return nil
}

// GetResourceMarker retrieves the MCP marker information for a resource on the current siteplan
func GetResourceMarker(resourceID uuid.UUID) (models.ResourceMarker, error) {
	SQL := `SELECT resource_id, marker_id, siteplan_id 
	FROM resource_marker
	WHERE resource_id = $1 AND siteplan_id = (SELECT siteplan_id FROM mcp_config WHERE id = 1)`

	var marker models.ResourceMarker
	err := DBConnPool.QueryRow(context.Background(), SQL, resourceID).Scan(
		&marker.ResourceID,
		&marker.MarkerID,
		&marker.SiteplanID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = DBConnPool.QueryRow(context.Background(), `SELECT siteplan_id FROM mcp_config WHERE id = 1`).Scan(&marker.SiteplanID)
		if err != nil {
			return marker, fmt.Errorf("failed to get siteplan ID: %w", err)
		}
	} else if err != nil {
		return marker, fmt.Errorf("failed to query resource marker: %w", err)
	}

	return marker, nil
}
