// Package handlers provides HTTP handlers for various system functionalities
package handlers

import (
	"IRIS-Server/internal/mcpcontrol"
	"IRIS-Server/internal/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime time.Time

// SystemHandler registers system monitoring and status HTTP endpoints
func SystemHandler(router *gin.Engine) {
	systemGroup := router.Group("/system")

	systemGroup.GET("/status", getSystemStatus)
	systemGroup.GET("/version", getSystemVersion)

	startTime = time.Now()
}

// getSystemStatus returns system health status
// @Summary Get system status
// @Description Returns system health information including database status, MCP connectivity, uptime, and active tracker count
// @Tags system
// @Produce json
// @Success 200 {object} object{status=string,uptime=string,database=string,mcp=string,active-trackers=int} "System status information"
// @Failure 500 {object} map[string]string "Failed to count active trackers"
// @Router /system/status [get]
func getSystemStatus(c *gin.Context) {
	dbStatus := repository.CheckDBConnection()
	mcpStatus := mcpcontrol.TestMCPConnection()
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
// @Summary Get application version
// @Description Returns the current version of the IRIS-Server application
// @Tags system
// @Produce json
// @Success 200 {object} object{version=string} "Application version"
// @Router /system/version [get]
func getSystemVersion(c *gin.Context) {
	// currently there is no versioning system yet. Return "beta" for now.

	c.JSON(http.StatusOK, gin.H{"version": "beta"})
}

func systemStatus(err error) string {
	if err != nil {
		return err.Error()
	}

	return "ok"
}
