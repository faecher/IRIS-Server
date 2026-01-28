package repository

import (
	"IRIS-Server/internal/models"
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func GetAllResources() ([]models.Resource, error) {
	SQL := `SELECT resource_id, name, type, status
	FROM resources`

	rows, err := DBConnPool.Query(context.Background(), SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resources, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Resource])
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return &resource, nil
}

func UpdateMarkerIDForResource(resourceID, markerID uuid.UUID) error {
	SQL := `INSERT INTO resource_marker (marker_id, resource_id, siteplan_id) VALUES ($1, $2, (SELECT siteplan_id FROM mcp_config WHERE id = 1))`

	_, err := DBConnPool.Exec(context.Background(), SQL, markerID, resourceID)
	return err
}

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
			return marker, err
		}
	}

	return marker, err
}
