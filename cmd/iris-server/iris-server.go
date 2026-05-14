// SPDX-License-Identifier: EUPL-1.2

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

	// --- Database ---
	err = repository.ConnectAndInitDatabase(cfg.SQL)
	if err != nil {
		slog.Error("Database connection and initialization failed:", "error", err)
		return
	}

	// --- MCP ---
	// Load mcp config from db
	mcpcontrol.InitMCPClient(cfg.MCP)
	loadAndInitMCP()

	// MCP Resource Sync Goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go syncResources(ctx, cfg.Update.ResourceUpdate)

	// If you need goroutines for background tasks, such as tracker polling,
	// start them here just like the MCP resource sync above. You can reuse the provided context.

	// --- Webserver ---
	// Start Webserver
	router := gin.Default()
	registerHandlers(router)

	slog.Info("Starting web server on " + cfg.Server.Address + ":" + cfg.Server.Port + "...")

	server := &http.Server{
		Addr:           cfg.Server.Address + ":" + cfg.Server.Port,
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

	if mcpConfig.URL == "" || mcpConfig.APIKey == "" {
		slog.Info("MCP configuration incomplete (missing URL or API key). MCP client not initialized.")
		return
	}

	slog.Info("MCP configuration found in database, initializing MCP client...")
	mcpcontrol.MCPConfig = mcpConfig

	if !mcpConfig.Enabled {
		slog.Info("MCP integration is disabled in the configuration. Skipping connection test.")
		return
	}

	err = mcpcontrol.TestMCPConnection(mcpcontrol.MCPConfig)
	if err != nil {
		slog.Error("MCP connection test failed. Disabling MCP client.", "error", err)

		mcpcontrol.MCPConfig.Enabled = false
		err := repository.UpdateMCPConfig(mcpcontrol.MCPConfig)
		if err != nil {
			slog.Error("Failed to update MCP configuration in database:", "error", err)
		}
	} else {
		slog.Info("MCP connection successful.")
	}
}

// syncResources periodically updates MCP resources in the database.
// interval is specified in seconds.
//
//nolint:contextcheck
func syncResources(ctx context.Context, interval uint16) {
	if interval <= 0 {
		slog.Info("MCP resource sync disabled (interval <= 0).")
		return
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !mcpcontrol.MCPConfig.Enabled {
				slog.Info("MCP integration is disabled. Skipping resource sync.")
				continue
			}

			err := mcpcontrol.UpdateMCPResourcesInDB()
			if err != nil {
				slog.Error("Failed to sync MCP resources:", "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
