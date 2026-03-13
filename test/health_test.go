package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter(t)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest("GET", "/api/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("attendu 200, reçu %d", rec.Code)
	}
}

func TestHealthCheckResponse(t *testing.T) {
	router := setupTestRouter(t)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest("GET", "/api/health", nil))

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("attendu Content-Type application/json, reçu %s", ct)
	}
}
