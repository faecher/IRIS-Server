package handlers

import (
	"IRIS-Server/internal/repository"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// TrackerHandler registers all tracker-related HTTP endpoints
func TrackerHandler(router *gin.Engine) {
	trackerGroup := router.Group("/tracker")

	trackerGroup.GET("/", listTrackers)
	trackerGroup.POST("assign/:tracker_id/:resource_id", assignResourceToTracker)
	trackerGroup.POST("rename/:tracker_id", renameTracker)
}

// listTrackers returns all trackers in the system
// @Summary Get all trackers
// @Description Returns a list of all trackers in the system with their current status, position, and assigned resources
// @Tags trackers
// @Produce json
// @Success 200 {array} models.BaseTracker "List of trackers"
// @Failure 500 {object} map[string]string "Failed to fetch trackers"
// @Router /tracker/ [get]
func listTrackers(c *gin.Context) {
	trackers, err := repository.GetAllTrackers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trackers"})
		return
	}

	c.JSON(http.StatusOK, trackers)
}

// assignResourceToTracker assigns or unassigns a resource to a tracker
// @Summary Assign or unassign resource to tracker
// @Description Assigns a resource to a tracker or removes the assignment. Pass empty string for resource_id to unassign.
// @Tags trackers
// @Param tracker_id path string true "Tracker UUID"
// @Param resource_id path string false "Resource UUID (empty to unassign)"
// @Success 200 "Assignment updated successfully"
// @Failure 400 {object} map[string]string "Invalid tracker or resource ID"
// @Failure 404 {object} map[string]string "Tracker or resource not found"
// @Failure 409 {object} map[string]string "Resource already assigned to another tracker"
// @Failure 500 {object} map[string]string "Failed to update assignment"
// @Router /tracker/assign/{tracker_id}/{resource_id} [post]
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			c.JSON(http.StatusConflict, gin.H{"error": "Resource is already assigned to another tracker"})
			return
		}
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tracker resource assignment"})
		return
	}

	c.Status(http.StatusOK)
}

// renameTracker updates the display name of a tracker
// @Summary Rename a tracker
// @Description Updates the display name of a tracker
// @Tags trackers
// @Accept json
// @Param tracker_id path string true "Tracker UUID"
// @Param request body object{newName=string} true "New tracker name"
// @Success 200 "Tracker renamed successfully"
// @Failure 400 {object} map[string]string "Invalid tracker ID or request body"
// @Failure 500 {object} map[string]string "Failed to rename tracker"
// @Router /tracker/rename/{tracker_id} [post]
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
	err = c.BindJSON(&req)
	if err != nil {
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
