package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/maksim-mshp/ozon-internship-task/internal/config"
	"github.com/maksim-mshp/ozon-internship-task/internal/httpserver"
	"github.com/maksim-mshp/ozon-internship-task/internal/links"
	"github.com/maksim-mshp/ozon-internship-task/internal/shortcode"
	"github.com/maksim-mshp/ozon-internship-task/internal/storage/memory"
	"github.com/maksim-mshp/ozon-internship-task/internal/storage/postgres"
)

type App struct {
	server     *http.Server
	closeStore func()
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	generator, err := shortcode.NewGenerator(cfg.ShortCode.Alphabet, cfg.ShortCode.Length)
	if err != nil {
		return nil, err
	}

	store, closeStore, err := newStorage(ctx, cfg)
	if err != nil {
		return nil, err
	}

	service := links.NewService(store, generator, cfg.ShortCode.MaxRetries)
	handler := links.NewHandler(service, cfg.BaseURL)

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", httpserver.WithError(func(w http.ResponseWriter, _ *http.Request) error {
		httpserver.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return nil
	}))
	handler.Register(mux)
	httpserver.RegisterSwagger(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: httpserver.Logging(mux),
	}

	return &App{server: server, closeStore: closeStore}, nil
}

func newStorage(ctx context.Context, cfg *config.Config) (links.Storage, func(), error) {
	switch cfg.Storage {
	case "memory":
		return memory.New(), func() {}, nil
	case "postgres":
		if err := postgres.RunMigrations(cfg.Database); err != nil {
			return nil, nil, err
		}

		pool, err := postgres.Connect(ctx, cfg.Database)
		if err != nil {
			return nil, nil, err
		}

		return postgres.New(pool), pool.Close, nil
	default:
		return nil, nil, fmt.Errorf("unknown storage type: %q", cfg.Storage)
	}
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	shutdownErr := a.server.Shutdown(ctx)
	a.closeStore()
	return shutdownErr
}
