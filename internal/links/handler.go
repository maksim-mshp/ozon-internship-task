package links

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/maksim-mshp/ozon-internship-task/internal/httpserver"
)

type Handler struct {
	service *Service
	baseURL string
}

func NewHandler(service *Service, baseURL string) *Handler {
	return &Handler{
		service: service,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("POST /shorten", httpserver.WithError(h.shorten))
	mux.Handle("GET /{code}", httpserver.WithError(h.resolve))
}

const maxRequestBodyBytes = 1 << 20

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

type ResolveResponse struct {
	URL string `json:"url"`
}

type errorCode string

const (
	errorBadRequest errorCode = "BAD_REQUEST"
	errorInvalidURL errorCode = "INVALID_URL"
	errorNotFound   errorCode = "NOT_FOUND"
	errorInternal   errorCode = "INTERNAL_ERROR"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    errorCode `json:"code"`
	Message string    `json:"message"`
}

// @Summary Создать короткую ссылку
// @Tags Links
// @Param request body ShortenRequest true "ShortenRequest"
// @Success 201 {object} ShortenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /shorten [POST]
func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, errorBadRequest, "invalid json body")
		return nil
	}

	link, err := h.service.Shorten(r.Context(), req.URL)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidURL):
			writeError(w, http.StatusBadRequest, errorInvalidURL, "url is invalid")
			return nil
		case errors.Is(err, ErrCodeGeneration):
			writeError(w, http.StatusInternalServerError, errorInternal, "could not generate unique code")
			return nil
		default:
			return err
		}
	}

	httpserver.RespondJSON(w, http.StatusCreated, ShortenResponse{
		Code:     link.Code,
		ShortURL: h.baseURL + "/" + link.Code,
	})
	return nil
}

// @Summary Получить оригинальный URL по короткому коду
// @Tags Links
// @Param code path string true "Короткий код"
// @Success 200 {object} ResolveResponse
// @Failure 404 {object} ErrorResponse
// @Router /{code} [GET]
func (h *Handler) resolve(w http.ResponseWriter, r *http.Request) error {
	link, err := h.service.Resolve(r.Context(), r.PathValue("code"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, errorNotFound, "short link not found")
			return nil
		}
		return err
	}

	httpserver.RespondJSON(w, http.StatusOK, ResolveResponse{URL: link.OriginalURL})
	return nil
}

func writeError(w http.ResponseWriter, status int, code errorCode, message string) {
	httpserver.RespondJSON(w, status, ErrorResponse{
		Error: ErrorBody{Code: code, Message: message},
	})
}
