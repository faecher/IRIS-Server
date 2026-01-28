package handlers

import (
	"IRIS-Server/internal/repository"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TrackerHandler(router *gin.Engine) {
	trackerGroup := router.Group("/tracker")

	trackerGroup.GET("/", listTrackers)
	trackerGroup.POST("assign/:tracker_id/:resource_id", assignResourceToTracker)
	trackerGroup.POST("rename/:tracker_id", renameTracker)

}

// listTrackers returns all trackers in the system
// GET /tracker/
func listTrackers(c *gin.Context) {
	trackers, err := repository.GetAllTrackers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trackers"})
		return
	}

	c.JSON(http.StatusOK, trackers)
}

// assignResourceToTracker assigns or unassigns a resource to a tracker
// POST /tracker/assign/{tracker_id}/{resource_id}
func assignResourceToTracker(c *gin.Context) {
	// parse tracker id
	trackerID, err := uuid.FromString(c.Param("tracker_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tracker id"})
		return
	}
	// test if tracker exists
	_, err = repository.GetTrackerByID(trackerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tracker not found"})
		return
	}

	// parse resource id
	resourceID, err := uuid.FromString(c.Param("resource_id"))
	if err != nil && c.Param("resource_id") != "" {
		// no resource id provided, unassign resource
		resourceID = uuid.Nil
	} else if err != nil {
		// resource id provided but invalid
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource id"})
		return
	}

	// assign or unassign resource
	if resourceID == uuid.Nil {
		err = repository.RemoveTrackerAssignment(trackerID)
	} else {
		_, err = repository.GetResourceByID(resourceID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}

		err = repository.UpdateTrackerResource(trackerID, resourceID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" { // unique_violation
					c.JSON(http.StatusConflict, gin.H{"error": "Resource is already assigned to another tracker"})
				}
			}
		}
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tracker resource assignment"})
		return
	}

	c.Status(http.StatusOK)
}

func renameTracker(c *gin.Context) {
	// parse tracker id
	trackerID, err := uuid.FromString(c.Param("tracker_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tracker id"})
		return
	}

	var req struct {
		NewName string `json:"newName"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = repository.RenameTracker(trackerID, req.NewName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rename tracker"})
		return
	}

	c.Status(http.StatusOK)
}
