package pgcql

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPgx(t *testing.T) {
	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx, "postgres",
		postgres.WithDatabase("crosslink"),
		postgres.WithUsername("crosslink"),
		postgres.WithPassword("crosslink"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	assert.NoError(t, err, "failed to start db container")
	var connStr string
	connStr, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
	assert.NoError(t, err, "failed to get db connection string")
	assert.NotEmpty(t, connStr, "connection string should not be empty")

	conn, err := pgx.Connect(ctx, connStr)
	assert.NoError(t, err, "failed to connect to db")
	defer func() {
		err := conn.Close(ctx)
		assert.NoError(t, err, "failed to close db connection")
	}()

	err = pgContainer.Terminate(ctx)
	assert.NoError(t, err, "failed to stop db container")
}
