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

var DBConnPool *pgxpool.Pool

var (
	ErrDatabaseNotInitialized  = errors.New("database not initialized")
	ErrDatabaseVersionMismatch = errors.New("database version mismatch")
)

func CheckDBConnection() error {
	if DBConnPool == nil {
		return ErrDatabaseNotInitialized
	}

	return DBConnPool.Ping(context.Background())
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
	conn.Config().AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxuuid.Register(conn.TypeMap())
		return nil
	}

	err, _ = checkDBVersionAndInit(conn, false)
	if err != nil {
		return fmt.Errorf("failed to check database version and initialize: %w", err)
	}

	DBConnPool = conn
	return nil
}

// checkDBVersionAndInit checks the DB version and runs the init script if the db doesnt have a version.
// Returns an error and bool = true if the db has been freshly initialized.
func checkDBVersionAndInit(conn *pgxpool.Pool, skipInit bool) (error, bool) {
	slog.Info("Checking database version...")
	var version int

	err := conn.QueryRow(context.Background(), "SELECT COALESCE(MAX(version), 0) FROM schema_versions").Scan(&version)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			slog.Error("Database error", "message", pgErr.Message)
		}
	}

	switch {
	case version == 0:
		if skipInit {
			return ErrDatabaseNotInitialized, false
		}
		slog.Info("Database is not initialized. Initializing now...")
		err := initDatabase(conn)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err), false
		}
		err, _ = checkDBVersionAndInit(conn, true)
		return err, true

	case version == currentDatabaseVersion:
		slog.Info("Database v1 detected. Yippie!")

	case version > currentDatabaseVersion:
		return fmt.Errorf("%w: expected v%d, got v%d", ErrDatabaseVersionMismatch, currentDatabaseVersion, version), false
	}

	slog.Info("Connected to database", "version", version)
	return nil, false
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
