package links

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestHandler(code string) *http.ServeMux {
	svc := NewService(newMapStorage(), fixedGenerator{code: code}, 5)
	mux := http.NewServeMux()
	NewHandler(svc, "http://short.test").Register(mux)
	return mux
}

func doJSON(t *testing.T, mux http.Handler, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, target, &buf)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func decodeErrorCode(t *testing.T, rec *httptest.ResponseRecorder) errorCode {
	t.Helper()
	var resp ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	return resp.Error.Code
}

func TestHandler_ShortenAndResolve(t *testing.T) {
	t.Parallel()
	mux := newTestHandler("ABCDEFGHIJ")

	rec := doJSON(t, mux, http.MethodPost, "/shorten", ShortenRequest{URL: "https://example.com/page"})
	require.Equal(t, http.StatusCreated, rec.Code)

	var created ShortenResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	require.Equal(t, "ABCDEFGHIJ", created.Code)
	require.Equal(t, "http://short.test/ABCDEFGHIJ", created.ShortURL)

	rec = doJSON(t, mux, http.MethodGet, "/"+created.Code, nil)
	require.Equal(t, http.StatusOK, rec.Code)

	var resolved ResolveResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resolved))
	require.Equal(t, "https://example.com/page", resolved.URL)
}

func TestHandler_ShortenDeduplicates(t *testing.T) {
	t.Parallel()
	mux := newTestHandler("ABCDEFGHIJ")

	first := doJSON(t, mux, http.MethodPost, "/shorten", ShortenRequest{URL: "https://example.com"})
	require.Equal(t, http.StatusCreated, first.Code)
	second := doJSON(t, mux, http.MethodPost, "/shorten", ShortenRequest{URL: "https://example.com"})
	require.Equal(t, http.StatusCreated, second.Code)

	require.JSONEq(t, first.Body.String(), second.Body.String())
}

func TestHandler_ShortenInvalidJSON(t *testing.T) {
	t.Parallel()
	mux := newTestHandler("ABCDEFGHIJ")

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader("{not json"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, errorBadRequest, decodeErrorCode(t, rec))
}

func TestHandler_ShortenInvalidURL(t *testing.T) {
	t.Parallel()
	mux := newTestHandler("ABCDEFGHIJ")

	rec := doJSON(t, mux, http.MethodPost, "/shorten", ShortenRequest{URL: "not a url"})
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, errorInvalidURL, decodeErrorCode(t, rec))
}

func TestHandler_ResolveNotFound(t *testing.T) {
	t.Parallel()
	mux := newTestHandler("ABCDEFGHIJ")

	rec := doJSON(t, mux, http.MethodGet, "/missingcode", nil)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Equal(t, errorNotFound, decodeErrorCode(t, rec))
}
