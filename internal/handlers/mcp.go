package handlers

import (
	"IRIS-Server/internal/mcp_control"
	"IRIS-Server/internal/models"
	"IRIS-Server/internal/repository"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
)

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
// GET /mcp/operations
func getMCPOperations(c *gin.Context) {
	operations, err := mcp_control.GetMCPOperations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch MCP operations: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, operations)
}

// enableMCPOperation marks an operation as selected
// POST /mcp/operations/set/:id
func setMCPOperation(c *gin.Context) {
	operationID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operation ID"})
		return
	}

	// Check if operation exists in MCP
	allOperations, err := mcp_control.GetMCPOperations()
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
// GET /mcp/siteplans
func getMCPSiteplans(c *gin.Context) {
	siteplans, err := mcp_control.GetMCPSiteplans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch MCP siteplans: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, siteplans)
}

func setMCPSiteplan(c *gin.Context) {
	siteplanID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid siteplan ID"})
		return
	}
	// Check if siteplan exists in MCP
	allSiteplans, err := mcp_control.GetMCPSiteplans()
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
