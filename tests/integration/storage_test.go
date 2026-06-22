package integration

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/maksim-mshp/ozon-internship-task/internal/config"
	"github.com/maksim-mshp/ozon-internship-task/internal/links"
	"github.com/maksim-mshp/ozon-internship-task/internal/storage/postgres"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	postgresImage   = "postgres:18-alpine"
	postgresUser    = "postgres"
	postgresPass    = "postgres"
	postgresDB      = "links_test"
	containerWindow = 2 * time.Minute
)

func newStorage(t *testing.T) *postgres.Storage {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), containerWindow)
	t.Cleanup(cancel)

	container, err := tcpostgres.Run(
		ctx,
		postgresImage,
		tcpostgres.BasicWaitStrategies(),
		tcpostgres.WithDatabase(postgresDB),
		tcpostgres.WithUsername(postgresUser),
		tcpostgres.WithPassword(postgresPass),
	)
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, container)

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	dbCfg := parseDatabaseConfig(t, connStr)
	require.NoError(t, postgres.RunMigrations(dbCfg))

	pool, err := postgres.Connect(ctx, dbCfg)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	return postgres.New(pool)
}

func parseDatabaseConfig(t *testing.T, connStr string) config.Database {
	t.Helper()

	parsed, err := url.Parse(connStr)
	require.NoError(t, err)

	password, ok := parsed.User.Password()
	require.True(t, ok)

	port, err := strconv.Atoi(parsed.Port())
	require.NoError(t, err)

	return config.Database{
		Host:     parsed.Hostname(),
		Port:     port,
		User:     parsed.User.Username(),
		Password: password,
		Database: strings.TrimPrefix(parsed.Path, "/"),
	}
}

func TestPostgresStorage_SaveAndGet(t *testing.T) {
	store := newStorage(t)
	ctx := context.Background()

	link := links.Link{Code: "abc1234567", OriginalURL: "https://example.com/page"}
	require.NoError(t, store.Save(ctx, link))

	byCode, err := store.GetByCode(ctx, link.Code)
	require.NoError(t, err)
	require.Equal(t, link, byCode)

	byOriginal, err := store.GetByOriginal(ctx, link.OriginalURL)
	require.NoError(t, err)
	require.Equal(t, link, byOriginal)
}

func TestPostgresStorage_NotFound(t *testing.T) {
	store := newStorage(t)
	ctx := context.Background()

	_, err := store.GetByCode(ctx, "missing")
	require.ErrorIs(t, err, links.ErrNotFound)

	_, err = store.GetByOriginal(ctx, "https://missing.example")
	require.ErrorIs(t, err, links.ErrNotFound)
}

func TestPostgresStorage_DuplicateCode(t *testing.T) {
	store := newStorage(t)
	ctx := context.Background()

	require.NoError(t, store.Save(ctx, links.Link{Code: "samecode00", OriginalURL: "https://a.example"}))

	err := store.Save(ctx, links.Link{Code: "samecode00", OriginalURL: "https://b.example"})
	require.ErrorIs(t, err, links.ErrCodeExists)
}

func TestPostgresStorage_DuplicateOriginal(t *testing.T) {
	store := newStorage(t)
	ctx := context.Background()

	require.NoError(t, store.Save(ctx, links.Link{Code: "code000001", OriginalURL: "https://same.example"}))

	err := store.Save(ctx, links.Link{Code: "code000002", OriginalURL: "https://same.example"})
	require.ErrorIs(t, err, links.ErrOriginalExists)
}
