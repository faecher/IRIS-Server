package handlers

import (
	"IRIS-Server/internal/mcp_control"
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// !!!!! TODO: do we need the rest of this functionality? we just want to move the pins in mcp?

func MCPHandler(router *gin.Engine) {
	mcpGroup := router.Group("/mcp")

	mcpGroup.GET("/operations", getMCPOperations)
	mcpGroup.POST("/operations/enable", enableMCPOperation)
	mcpGroup.POST("/operations/disable", disableMCPOperation)
	mcpGroup.POST("/start", startMCPIntegration)
	mcpGroup.GET("/config", getMCPConfig)
}

// getMCPOperations returns all active MCP operations
// GET /mcp/operations
func getMCPOperations(c *gin.Context) {
	// TODO: Query operations table filtered by active=true
	// TODO: Return JSON array of Operation models with 200 status

	c.Status(http.StatusNotImplemented)
}

// enableMCPOperation marks an operation as selected
// POST /mcp/operations/enable
func enableMCPOperation(c *gin.Context) {
	// TODO: Parse MCPOperationConfig from request body (contains uid)
	// TODO: Find operation by uid in database
	// TODO: If operation not found, return 404 error
	// TODO: Set operation.selected = true
	// TODO: Commit changes to database
	// TODO: Return {"status": 200}

	c.Status(http.StatusNotImplemented)
}

// disableMCPOperation marks an operation as not selected
// POST /mcp/operations/disable
func disableMCPOperation(c *gin.Context) {
	// TODO: Parse MCPOperationConfig from request body (contains uid)
	// TODO: Find operation by uid in database
	// TODO: If operation not found, return 404 error
	// TODO: Set operation.selected = false
	// TODO: Commit changes to database
	// TODO: Return {"status": 200}

	c.Status(http.StatusNotImplemented)
}

// startMCPIntegration configures and enables MCP integration
// POST /mcp/start
func startMCPIntegration(c *gin.Context) {
	var config models.MCPConfig

	if err := c.ShouldBindJSON(&config); err != nil {
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
	err := testMCPConnection(config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MCP server not reachable: " + err.Error()})
		return
	}

	// Store config in database
	err = repository.UpdateMCPConfig(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save MCP config: " + err.Error()})
		return
	} else {
		// update local MCPConfig variable only if DB update was successful
		mcp_control.MCPConfig = config
	}

	c.Status(http.StatusOK)
}

// getMCPConfig returns current MCP configuration
// GET /mcp/config
func getMCPConfig(c *gin.Context) {
	config, err := repository.GetMCPConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch MCP config"})
		return
	}

	// TODO: Query for active AND selected operation
	// TODO: Build response object with: ..., operation_selected, operation (uid)

	c.JSON(http.StatusOK, gin.H{
		"enabled": config.Enabled,
		"api_key": config.APIKey,
		"url":     config.URL,
	})
}

func testMCPConnection(newConfig models.MCPConfig) error {
	oldConfig := mcp_control.MCPConfig
	defer func() {
		mcp_control.MCPConfig = oldConfig
	}()

	mcp_control.MCPConfig = newConfig

	err := mcp_control.TestMCPConnection()
	if err != nil {
		return err
	}

	return nil
}
