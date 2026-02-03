// SPDX-License-Identifier: EUPL-1.2

package handlers

import (
	"IRIS-Server/internal/mcpcontrol"
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
)

// MCPHandler registers all MCP integration HTTP endpoints
func MCPHandler(router *gin.Engine) {
	mcpGroup := router.Group("/mcp")

	mcpGroup.GET("/operations", getMCPOperations)
	mcpGroup.POST("/operations/set/:id", setMCPOperation)

	mcpGroup.GET("/siteplans", getMCPSiteplans)
	mcpGroup.POST("/siteplans/set/:id", setMCPSiteplan)

	mcpGroup.POST("/start", startMCPIntegration)
	mcpGroup.GET("/config", getMCPConfig)
}

// getMCPOperations returns all active MCP operations
// @Summary Get MCP operations
// @Description Fetches all active operations from the MCP API
// @Tags mcp
// @Produce json
// @Success 200 {array} models.MCPOperation "List of active operations"
// @Failure 500 {object} map[string]string "Failed to fetch MCP operations"
// @Router /mcp/operations [get]
func getMCPOperations(c *gin.Context) {
	operations, err := mcpcontrol.GetMCPOperations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch MCP operations: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, operations)
}

// setMCPOperation marks an operation as selected
// @Summary Select MCP operation
// @Description Validates and selects an operation from MCP. This operation will be used for subsequent siteplan selection.
// @Tags mcp
// @Param id path string true "Operation UUID"
// @Success 200 "Operation selected successfully"
// @Failure 400 {object} map[string]string "Invalid operation ID or operation not found in MCP"
// @Failure 500 {object} map[string]string "Failed to save operation"
// @Router /mcp/operations/set/{id} [post]
//
//nolint:dupl
func setMCPOperation(c *gin.Context) {
	operationID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operation ID"})
		return
	}

	// Check if operation exists in MCP
	allOperations, err := mcpcontrol.GetMCPOperations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch MCP operations: " + err.Error()})
		return
	}

	if !slices.ContainsFunc(allOperations, func(operation models.MCPOperation) bool {
		return operation.ID == operationID
	}) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Operation ID not found in mcp"})
		return
	}

	// Save selected operation in database
	err = repository.UpdateMCPOperation(operationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save MCP operation: " + err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// getMCPSiteplans returns all MCP siteplans
// @Summary Get MCP siteplans
// @Description Fetches all siteplans for the currently selected operation from the MCP API
// @Tags mcp
// @Produce json
// @Success 200 {array} models.MCPSiteplan "List of siteplans"
// @Failure 500 {object} map[string]string "Failed to fetch siteplans or no operation selected"
// @Router /mcp/siteplans [get]
func getMCPSiteplans(c *gin.Context) {
	siteplans, err := mcpcontrol.GetMCPSiteplans()
	if err != nil {
		slog.Error("Failed to fetch MCP siteplans", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch MCP siteplans: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, siteplans)
}

// setMCPSiteplan selects a siteplan for the current operation
// @Summary Select MCP siteplan
// @Description Validates and selects a siteplan from MCP. This siteplan will be used for marker operations.
// @Tags mcp
// @Param id path string true "Siteplan UUID"
// @Success 200 "Siteplan selected successfully"
// @Failure 400 {object} map[string]string "Invalid siteplan ID or siteplan not found in MCP"
// @Failure 500 {object} map[string]string "Failed to save siteplan"
// @Router /mcp/siteplans/set/{id} [post]
//
//nolint:dupl
func setMCPSiteplan(c *gin.Context) {
	siteplanID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid siteplan ID"})
		return
	}
	// Check if siteplan exists in MCP
	allSiteplans, err := mcpcontrol.GetMCPSiteplans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch MCP siteplans: " + err.Error()})
		return
	}
	if !slices.ContainsFunc(allSiteplans, func(siteplan models.MCPSiteplan) bool {
		return siteplan.ID == siteplanID
	}) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Siteplan ID not found in mcp"})
		return
	}

	// Save selected siteplan in database
	err = repository.UpdateMCPSiteplan(siteplanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save MCP siteplan: " + err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// startMCPIntegration configures and enables MCP integration
// @Summary Configure MCP integration
// @Description Tests connection and saves MCP configuration (URL, API key, enabled state)
// @Tags mcp
// @Accept json
// @Param config body models.MCPConfig true "MCP configuration"
// @Success 200 "MCP integration configured successfully"
// @Failure 400 {object} map[string]string "Invalid request or MCP server not reachable"
// @Failure 500 {object} map[string]string "Failed to save configuration"
// @Router /mcp/start [post]
func startMCPIntegration(c *gin.Context) {
	var config models.MCPConfig

	err := c.ShouldBindJSON(&config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Sanitize URL:
	// ensure https:// prefix
	if !strings.HasPrefix(config.URL, "https://") && !strings.HasPrefix(config.URL, "http://") {
		config.URL = "https://" + config.URL
	}
	// remove trailing slash
	config.URL = strings.TrimSuffix(config.URL, "/")

	// Test MCP connection
	err = testMCPConnection(config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MCP server not reachable: " + err.Error()})
		return
	}

	// Store config in database
	err = repository.UpdateMCPConfig(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save MCP config: " + err.Error()})
		return
	}

	// update local MCPConfig variable only if DB update was successful
	mcpcontrol.MCPConfig = config

	err = mcpcontrol.UpdateMCPResourcesInDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update MCP resources: " + err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// getMCPConfig returns current MCP configuration
// @Summary Get MCP configuration
// @Description Returns the current MCP configuration (including sensitive data like full API key)
// @Tags mcp
// @Produce json
// @Success 200 {object} object{enabled=bool,api_key=string,url=string,operation_id=string,siteplan_id=string} "Current MCP configuration"
// @Failure 500 {object} map[string]string "Failed to fetch configuration"
// @Router /mcp/config [get]
func getMCPConfig(c *gin.Context) {
	config, err := repository.GetMCPConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch MCP config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled":      config.Enabled,
		"api_key":      config.APIKey,
		"url":          config.URL,
		"operation_id": config.OperationID,
		"siteplan_id":  config.SiteplanID,
	})
}

func testMCPConnection(newConfig models.MCPConfig) error {
	oldConfig := mcpcontrol.MCPConfig
	defer func() {
		mcpcontrol.MCPConfig = oldConfig
	}()

	mcpcontrol.MCPConfig = newConfig

	err := mcpcontrol.TestMCPConnection()
	if err != nil {
		return fmt.Errorf("failed to test MCP connection: %w", err)
	}

	return nil
}
