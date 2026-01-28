package handlers

import (
	"IRIS-Server/internal/mcp_control"
	"IRIS-Server/internal/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime time.Time

func SystemHandler(router *gin.Engine) {
	systemGroup := router.Group("/system")

	systemGroup.GET("/status", getSystemStatus)
	systemGroup.GET("/version", getSystemVersion)

	startTime = time.Now()
}

// getSystemStatus returns system health status
// GET /system/status
func getSystemStatus(c *gin.Context) {
	dbStatus := repository.CheckDBConnection()
	mcpStatus := mcp_control.TestMCPConnection()
	activeTrackers, err := repository.GetActiveTrackerCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count active trackers"})
		return
	}

	statusString := "ok"
	if dbStatus != nil || mcpStatus != nil {
		statusString = "error"
	}

	status := gin.H{
		"status":          statusString,
		"uptime":          time.Since(startTime).String(),
		"database":        systemStatus(dbStatus),
		"mcp":             systemStatus(mcpStatus),
		"active-trackers": activeTrackers,
	}

	c.JSON(http.StatusOK, status)
}

// getSystemVersion returns the application version
// GET /system/version
func getSystemVersion(c *gin.Context) {
	// TODO: Read version from build-time variable or config
	// TODO: Return version information
	// currently there is no versioning system yet. Return "beta" for now.

	c.JSON(http.StatusOK, gin.H{"version": "beta"})
}

func systemStatus(err error) string {
	if err != nil {
		return err.Error()
	}

	return "ok"
}
