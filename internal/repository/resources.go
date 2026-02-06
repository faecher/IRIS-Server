// SPDX-License-Identifier: EUPL-1.2

package repository

import (
	"IRIS-Server/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

// ErrNoResourceFound indicates that no resource was found for the given criteria
var ErrNoResourceFound = errors.New("no resource found")

// GetAllResources retrieves all tableau resources for the currently selected operation
func GetAllResources() ([]models.TableauResource, error) {
	SQL := `
	SELECT 
		tr.tableau_resource_id,
		tr.operation_id,
		tr.status,
		r.resource_id,
		r.name,
		r.type
	FROM tableau_resources tr
	JOIN resources r ON tr.resource_id = r.resource_id
	WHERE tr.operation_id = (SELECT operation_id FROM mcp_config WHERE id = 1)`

	rows, err := DBConnPool.Query(context.Background(), SQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query all resources: %w", err)
	}
	defer rows.Close()

	var resources []models.TableauResource
	for rows.Next() {
		var tableauResource models.TableauResource
		err := rows.Scan(
			&tableauResource.ID,
			&tableauResource.OperationID,
			&tableauResource.Status,
			&tableauResource.Resource.ID,
			&tableauResource.Resource.Name,
			&tableauResource.Resource.Type,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan resource row: %w", err)
		}
		resources = append(resources, tableauResource)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error iterating resource rows: %w", err)
	}

	return resources, nil
}

// GetResourceByID retrieves a single tableau resource by its base resource UUID
func GetResourceByID(resourceID uuid.UUID) (*models.TableauResource, error) {
	SQL := `
	SELECT 
		tr.tableau_resource_id,
		tr.operation_id,
		tr.status,
		r.resource_id,
		r.name,
		r.type
	FROM tableau_resources tr
	JOIN resources r ON tr.resource_id = r.resource_id
	WHERE r.resource_id = $1 
	  AND tr.operation_id = (SELECT operation_id FROM mcp_config WHERE id = 1)`

	var tableauResource models.TableauResource
	err := DBConnPool.QueryRow(context.Background(), SQL, resourceID).Scan(
		&tableauResource.ID,
		&tableauResource.OperationID,
		&tableauResource.Status,
		&tableauResource.Resource.ID,
		&tableauResource.Resource.Name,
		&tableauResource.Resource.Type,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoResourceFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query resource by ID: %w", err)
	}

	return &tableauResource, nil
}

// UpdateMarkerIDForResource associates an MCP marker ID with a resource for the current siteplan
func UpdateMarkerIDForResource(resourceID, markerID uuid.UUID) error {
	SQL := `INSERT INTO resource_marker (marker_id, resource_id, siteplan_id) 
	VALUES ($1, $2, (SELECT siteplan_id FROM mcp_config WHERE id = 1))
	ON CONFLICT (resource_id, siteplan_id) DO UPDATE SET marker_id = EXCLUDED.marker_id`

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

// UpsertResource creates or updates a tableau resource in the database
func UpsertResource(resource *models.TableauResource) error {
	// First, upsert the base resource
	resourceSQL := `
	INSERT INTO resources (resource_id, name, type) 
	VALUES ($1, $2, $3) 
	ON CONFLICT (resource_id) DO UPDATE 
	SET name = EXCLUDED.name, type = EXCLUDED.type`

	_, err := DBConnPool.Exec(context.Background(), resourceSQL,
		resource.Resource.ID,
		resource.Resource.Name,
		resource.Resource.Type,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert base resource: %w", err)
	}

	// Then, upsert the tableau_resource
	tableauSQL := `
	INSERT INTO tableau_resources (tableau_resource_id, resource_id, operation_id, status) 
	VALUES ($1, $2, $3, $4) 
	ON CONFLICT (resource_id, operation_id) DO UPDATE 
	SET status = EXCLUDED.status, tableau_resource_id = EXCLUDED.tableau_resource_id`

	_, err = DBConnPool.Exec(context.Background(), tableauSQL,
		resource.ID,
		resource.Resource.ID,
		resource.OperationID,
		resource.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert tableau resource: %w", err)
	}

	return nil
}
