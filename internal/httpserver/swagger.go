package httpserver

import (
	"net/http"

	projectembed "github.com/maksim-mshp/ozon-internship-task"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @Title			Link Shortener API
// @Description	Сервис укорачивания ссылок
// @Servers.Url	/
func RegisterSwagger(mux *http.ServeMux) {
	mux.HandleFunc("/swagger/openapi.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, projectembed.SwaggerFS, "api/openapi.yml")
	})
	mux.HandleFunc("/swagger/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, projectembed.SwaggerFS, "api/openapi.json")
	})
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/openapi.yml"),
	))
}
