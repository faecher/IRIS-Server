package main

import (
	"IRIS-Server/internal/config"
	"IRIS-Server/internal/handlers"
	"IRIS-Server/internal/repository"
	"log"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	slog.Info(`
            ██╗██████╗ ██╗███████╗
            ██║██╔══██╗██║██╔════╝
            ██║██████╔╝██║███████╗
            ██║██╔══██╗██║╚════██║
            ██║██║  ██║██║███████║
Welcome to	╚═╝╚═╝  ╚═╝╚═╝╚══════╝`)
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

	router := gin.Default()
	registerHandlers(router)

	// If you need goroutines for background tasks, such as tracker polling, start them here.
	// TODO: sync resources in mcp to registered resources in db

	slog.Info("Starting web server on " + cfg.Server.Address + "...")

	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router,
	}

	err = server.ListenAndServe()
	if err != nil {
		slog.Error("Server failed:", "error", err)
	}
}

func registerHandlers(router *gin.Engine) {
	handlers.TrackerHandler(router)
	handlers.GatewayHandler(router)
	handlers.ResourcesHandler(router)
	handlers.MCPHandler(router)
	handlers.SystemHandler(router)
}
