package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/maksim-mshp/ozon-internship-task/internal/config"
	"github.com/maksim-mshp/ozon-internship-task/internal/httpserver"
)

type App struct {
	server *http.Server
}

func New(_ context.Context, cfg *config.Config) (*App, error) {
	mux := http.NewServeMux()
	mux.Handle("GET /healthz", httpserver.WithError(func(w http.ResponseWriter, _ *http.Request) error {
		httpserver.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return nil
	}))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: httpserver.Logging(mux),
	}

	return &App{server: server}, nil
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
