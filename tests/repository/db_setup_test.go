package repository_test

import (
	"IRIS-Server/internal/repository"
	"context"
	"os"
	"testing"
	"time"

	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMain(m *testing.M) {
	testPool, container := DatabaseSetup()
	if testPool == nil {
		panic("Failed to set up database")
	}
	repository.DBConnPool = testPool

	code := m.Run()

	if testPool != nil {
		testPool.Close()
	}

	container.Terminate(context.Background())

	os.Exit(code)
}

func DatabaseSetup() (*pgxpool.Pool, *postgres.PostgresContainer) {
	container, err := postgres.Run(context.Background(), "postgres:17",
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
	)
	if err != nil {
		panic(err)
	}

	connStr, err := container.ConnectionString(context.Background(), "sslmode=disable")
	if err != nil {
		panic(err)
	}

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		panic(err)
	}
	pool.Config().AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxuuid.Register(conn.TypeMap())
		return nil
	}

	// Init with the current database schema
	_, err = repository.CheckDBVersionAndInit(pool, false)
	if err != nil {
		panic(err)
	}

	return pool, container
}
