package repository

import (
	"IRIS-Server/internal/models"
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func GetAllResources() ([]models.Resource, error) {
	SQL := `SELECT resource_id, name, type, status, marker_id FROM resources`

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
	SQL := `SELECT resource_id, name, type, status, marker_id FROM resources WHERE resource_id = $1`

	var resource models.Resource
	err := DBConnPool.QueryRow(context.Background(), SQL, resourceID).Scan(
		&resource.ID,
		&resource.Name,
		&resource.Type,
		&resource.Status,
		&resource.MarkerID,
	)

	if err != nil {
		return nil, err
	}

	return &resource, nil
}

func UpdateMarkerIDForResource(resourceID uuid.UUID, markerID uuid.UUID) error {
	SQL := `UPDATE resources SET marker_id = $1 WHERE resource_id = $2`

	_, err := DBConnPool.Exec(context.Background(), SQL, markerID, resourceID)
	return err
}
