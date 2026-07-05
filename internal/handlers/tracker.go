// SPDX-License-Identifier: EUPL-1.2

package handlers

import (
	"IRIS-Server/internal/mcpcontrol"
	"IRIS-Server/internal/repository"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
)

// TrackerHandler registers all tracker-related HTTP endpoints
func TrackerHandler(router *gin.Engine) {
	trackerGroup := router.Group("/tracker")

	trackerGroup.GET("/", listTrackers)
	trackerGroup.POST("assign/:tracker_id/:tableau_resource_id", assignResourceToTracker)
	trackerGroup.POST("assign/:tracker_id", unassignResourceFromTracker)
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
		slog.Error("Failed to fetch trackers", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trackers"})
		return
	}

	c.JSON(http.StatusOK, trackers)
}

// assignResourceToTracker assigns a resource to a tracker
// @Summary Assign a resource to tracker
// @Description Assigns a resource to a tracker.
// @Tags trackers
// @Param tracker_id path string true "Tracker UUID"
// @Param tableau_resource_id path string true "Tableau Resource UUID"
// @Success 200 "Assignment updated successfully"
// @Failure 400 {object} map[string]string "Invalid tracker or tableau resource ID"
// @Failure 404 {object} map[string]string "Tracker or tableau resource not found"
// @Failure 500 {object} map[string]string "Failed to update assignment"
// @Router /tracker/assign/{tracker_id}/{tableau_resource_id} [post]
func assignResourceToTracker(c *gin.Context) {
	trackerID, err := parseAndVerifyTrackerID(c)
	if err != nil {
		return
	}

	// parse tableau resource id
	tableauResourceID, err := uuid.FromString(c.Param("tableau_resource_id"))
	if err != nil {
		// tableau resource id provided but invalid
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tableau resource id"})
		return
	}

	// assign resource
	_, err = repository.GetResourceByID(context.Background(), tableauResourceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tableau resource not found"})
		return
	}

	err = repository.UpdateTrackerResource(trackerID, tableauResourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tracker resource assignment"})
		return
	}

	err = mcpcontrol.UpdateMarkerInMCP(c.Request.Context(), trackerID)
	if err != nil {
		slog.Error("Failed to update marker in MCP for tracker", "trackerID", trackerID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update marker in MCP: %v", err)})
		return
	}

	c.Status(http.StatusOK)
}

// unassignResourceFromTracker removes the resource assignment from a tracker
// @Summary Unassign resource from tracker
// @Description Removes the resource assignment from a tracker
// @Tags trackers
// @Param tracker_id path string true "Tracker UUID"
// @Success 200 "Resource unassigned successfully"
// @Failure 400 {object} map[string]string "Invalid tracker ID"
// @Failure 404 {object} map[string]string "Tracker not found"
// @Failure 500 {object} map[string]string "Failed to remove tracker resource assignment"
// @Router /tracker/assign/{tracker_id} [post]
func unassignResourceFromTracker(c *gin.Context) {
	trackerID, err := parseAndVerifyTrackerID(c)
	if err != nil {
		return
	}

	if mcpcontrol.MCPConfig.DeleteMarkersOnUnassign {
		tracker, err := repository.GetTrackerByID(context.Background(), trackerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get resource marker"})
			return
		}

		count, err := repository.GetTrackerCountForResource(tracker.TableauResource.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get resource marker"})
			return
		}

		if count <= 1 {
			err = mcpcontrol.DeleteMarkerForTracker(trackerID)
			if err != nil {
				slog.Error("Failed to delete marker in MCP for tracker", "trackerID", trackerID, "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete marker in MCP: %v", err)})
				return
			}
		}
	}

	err = repository.RemoveTrackerAssignment(trackerID)
	if err != nil {
		slog.Error("Failed to remove tracker assignment", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove tracker resource assignment"})
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

func parseAndVerifyTrackerID(c *gin.Context) (uuid.UUID, error) {
	// parse tracker id
	trackerID, err := uuid.FromString(c.Param("tracker_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tracker id"})
		return uuid.Nil, fmt.Errorf("invalid tracker ID: %w", err)
	}
	// test if tracker exists
	_, err = repository.GetTrackerByID(context.Background(), trackerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tracker not found"})
		return uuid.Nil, fmt.Errorf("tracker not found: %w", err)
	}

	return trackerID, nil
}
