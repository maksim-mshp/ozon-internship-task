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

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

type resolveResponse struct {
	URL string `json:"url"`
}

type errorCode string

const (
	errorBadRequest errorCode = "BAD_REQUEST"
	errorInvalidURL errorCode = "INVALID_URL"
	errorNotFound   errorCode = "NOT_FOUND"
	errorInternal   errorCode = "INTERNAL_ERROR"
)

type errorResponse struct {
	Error struct {
		Code    errorCode `json:"code"`
		Message string    `json:"message"`
	} `json:"error"`
}

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) error {
	var req shortenRequest
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

	httpserver.RespondJSON(w, http.StatusCreated, shortenResponse{
		Code:     link.Code,
		ShortURL: h.baseURL + "/" + link.Code,
	})
	return nil
}

func (h *Handler) resolve(w http.ResponseWriter, r *http.Request) error {
	link, err := h.service.Resolve(r.Context(), r.PathValue("code"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, errorNotFound, "short link not found")
			return nil
		}
		return err
	}

	httpserver.RespondJSON(w, http.StatusOK, resolveResponse{URL: link.OriginalURL})
	return nil
}

func writeError(w http.ResponseWriter, status int, code errorCode, message string) {
	var resp errorResponse
	resp.Error.Code = code
	resp.Error.Message = message
	httpserver.RespondJSON(w, status, resp)
}
