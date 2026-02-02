// Package main implements the entry point for the IRIS Server application.
package main

import (
	"IRIS-Server/internal/config"
	"IRIS-Server/internal/handlers"
	"IRIS-Server/internal/mcpcontrol"
	"IRIS-Server/internal/repository"
	"context"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title IRIS-Server
// @version 1.0
// @description Tracker and MCP integration
// @host 0.0.0.0:8080
// @BasePath /
func main() {
	slog.Info(`
            ██╗██████╗ ██╗███████╗
            ██║██╔══██╗██║██╔════╝
            ██║██████╔╝██║███████╗
            ██║██╔══██╗██║╚════██║
            ██║██║  ██║██║███████║
Welcome to  ╚═╝╚═╝  ╚═╝╚═╝╚══════╝`)
	slog.Info("Starting IRIS Server...")

	slog.Info("Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	err = repository.ConnectAndInitDatabase(cfg.SQL)
	if err != nil {
		slog.Error("Database connection and initialization failed:", "error", err)
		return
	}

	// Load mcp config from db
	loadAndInitMCP()

	// Start Webserver
	router := gin.Default()
	registerHandlers(router)

	slog.Info("Starting web server on " + cfg.Server.Address + "...")

	server := &http.Server{
		Addr:           cfg.Server.Address,
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(cfg.Server.IdleTimeout) * time.Minute,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	err = server.ListenAndServe()
	if err != nil {
		slog.Error("Server failed:", "error", err)
	}

	// MCP Resource Sync Goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go syncResources(ctx)

	// If you need goroutines for background tasks, such as tracker polling,
	// start them here just like the MCP resource sync above. You can reuse the provided context.
	// TODO: handle tracker update when no resource is assigned -> currently errors
}

func registerHandlers(router *gin.Engine) {
	handlers.MCPHandler(router)
	handlers.SystemHandler(router)
	handlers.TrackerHandler(router)
	handlers.GatewayHandler(router)
	handlers.ResourcesHandler(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// loadAndInitMCP loads MCP configuration from the database and initializes the MCP client.
// If the configuration is incomplete or the connection test fails, it disables the MCP client.
func loadAndInitMCP() {
	mcpConfig, err := repository.GetMCPConfig()
	if err != nil {
		slog.Error("Failed to load MCP configuration from database:", "error", err)
		return
	}

	if mcpConfig.URL != "" && mcpConfig.APIKey != "" {
		slog.Info("MCP configuration found in database, initializing MCP client...")
		mcpcontrol.MCPConfig = mcpConfig

		err = mcpcontrol.TestMCPConnection()
		if err != nil {
			slog.Error("MCP connection test failed. Disabling MCP client.", "error", err)

			mcpcontrol.MCPConfig.Enabled = false
			repository.UpdateMCPConfig(mcpcontrol.MCPConfig)
		} else {
			slog.Info("MCP connection successful.")
		}
	}
}

//nolint:contextcheck
func syncResources(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // TODO: make interval configurable
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := mcpcontrol.UpdateMCPResourcesInDB()
			if err != nil {
				slog.Error("Failed to sync MCP resources:", "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
