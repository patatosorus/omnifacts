package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthRegisterAndLogin(t *testing.T) {
	router := setupTestRouter(t)
	token := registerAndLogin(t, router, "auth-test-user", "auth@test.com", "password123")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("GET", "/api/repos", nil, token))
	if rec.Code != http.StatusOK {
		t.Fatalf("accès authentifié: attendu 200, reçu %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest("GET", "/api/repos", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("accès non-authentifié: attendu 401, reçu %d", rec.Code)
	}
}

func TestAuthRegisterValidation(t *testing.T) {
	router := setupTestRouter(t)

	t.Run("champs manquants", func(t *testing.T) {
		b, _ := json.Marshal(map[string]string{"username": "x"})
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("attendu 400, reçu %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("mot de passe court", func(t *testing.T) {
		b, _ := json.Marshal(map[string]string{"username": "shortpw", "email": "s@t.com", "password": "123"})
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("attendu 400, reçu %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("doublon", func(t *testing.T) {
		registerAndLogin(t, router, "dup-user", "dup@test.com", "password123")
		b, _ := json.Marshal(map[string]string{"username": "dup-user", "email": "dup2@test.com", "password": "password123"})
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusConflict {
			t.Fatalf("doublon: attendu 409, reçu %d", rec.Code)
		}
	})
}

func TestAuthAPIKeyGeneration(t *testing.T) {
	router := setupTestRouter(t)
	token := registerAndLogin(t, router, "apikey-user", "apikey@test.com", "password123")

	t.Run("create", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"name": "test-key"})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, authReq("POST", "/api/auth/apikeys", bytes.NewReader(body), token))

		if rec.Code != http.StatusCreated {
			t.Fatalf("création clé API: attendu 201, reçu %d. Body: %s", rec.Code, rec.Body.String())
		}

		var keyResp struct {
			Data struct {
				Key string `json:"key"`
			} `json:"data"`
		}
		json.NewDecoder(rec.Body).Decode(&keyResp)
		if keyResp.Data.Key == "" {
			t.Fatal("clé API vide")
		}

		req := httptest.NewRequest("GET", "/api/repos", nil)
		req.Header.Set("X-API-Key", keyResp.Data.Key)
		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("accès via API key: attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("list", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, authReq("GET", "/api/auth/apikeys", nil, token))
		if rec.Code != http.StatusOK {
			t.Fatalf("liste clés API: attendu 200, reçu %d", rec.Code)
		}
	})
}

func TestAuthUnauthenticatedAccess(t *testing.T) {
	router := setupTestRouter(t)

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/repos"},
		{"POST", "/api/repos/generic"},
		{"GET", "/api/artefacts"},
		{"POST", "/api/auth/apikeys"},
		{"GET", "/api/auth/apikeys"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(ep.method, ep.path, nil))
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("attendu 401, reçu %d", rec.Code)
			}
		})
	}
}
