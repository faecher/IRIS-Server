// Package repository provides the data access layer for the IRIS application,
// handling all database operations, connections, and SQL query execution.
package repository

import (
	"IRIS-Server/internal/config"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"

	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed db-v1-setup.sql
var sqlSetup string

const currentDatabaseVersion = 1

// DBConnPool is the global database connection pool used by all repository functions
var DBConnPool *pgxpool.Pool

var (
	// ErrDatabaseNotInitialized indicates that the database connection pool is not initialized
	ErrDatabaseNotInitialized = errors.New("database not initialized")

	// ErrDatabaseVersionMismatch indicates that the database version does not match the expected version
	ErrDatabaseVersionMismatch = errors.New("database version mismatch")
)

// CheckDBConnection verifies that the database connection pool is initialized and responsive
func CheckDBConnection() error {
	if DBConnPool == nil {
		return ErrDatabaseNotInitialized
	}

	err := DBConnPool.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// ConnectAndInitDatabase connects to the database and initializes it if necessary.
// Returns the connection pool and an error if any.
func ConnectAndInitDatabase(config config.SQLConfig) error {
	slog.Info("Connecting to database...")

	conn, err := pgxpool.New(context.Background(), config.ConnectionString())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create database connection pool: %v\n", err)
		return fmt.Errorf("failed to create database connection pool: %w", err)
	}
	slog.Info("Connected to database.")
	conn.Config().AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxuuid.Register(conn.TypeMap())
		return nil
	}

	_, err = CheckDBVersionAndInit(conn, false)
	if err != nil {
		return fmt.Errorf("failed to check database version and initialize: %w", err)
	}

	DBConnPool = conn
	return nil
}

// CheckDBVersionAndInit checks the DB version and runs the init script if the db doesnt have a version.
// Returns an error and bool = true if the db has been freshly initialized.
// Exported for testing purposes.
func CheckDBVersionAndInit(conn *pgxpool.Pool, skipInit bool) (bool, error) {
	slog.Info("Checking database version...")
	var version int

	err := conn.QueryRow(context.Background(), "SELECT COALESCE(MAX(version), 0) FROM schema_versions").Scan(&version)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		slog.Error("Database error", "message", pgErr.Message)
	}

	switch {
	case version == 0:
		if skipInit {
			return false, ErrDatabaseNotInitialized
		}
		slog.Info("Database is not initialized. Initializing now...")
		err := initDatabase(conn)
		if err != nil {
			return false, fmt.Errorf("failed to initialize database: %w", err)
		}
		_, err = CheckDBVersionAndInit(conn, true)
		return true, err

	case version == currentDatabaseVersion:
		slog.Info("Database v1 detected. Yippie!")

	case version > currentDatabaseVersion:
		return false, fmt.Errorf("%w: expected v%d, got v%d", ErrDatabaseVersionMismatch, currentDatabaseVersion, version)
	}

	slog.Info("Connected to database", "version", version)
	return false, nil
}

func initDatabase(conn *pgxpool.Pool) error {
	slog.Info("Initializing database...")

	_, err := conn.Exec(context.Background(), sqlSetup)
	if err != nil {
		return fmt.Errorf("failed to execute initialization script: %w", err)
	}

	slog.Info("Database initialization script run successfully.")

	return nil
}
