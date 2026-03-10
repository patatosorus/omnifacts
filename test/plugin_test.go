package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPluginListTypes(t *testing.T) {
	router := setupTestRouter(t)
	token := registerAndLogin(t, router, "types-user", "types@test.com", "password123")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("GET", "/api/repos/types", nil, token))

	if rec.Code != http.StatusOK {
		t.Fatalf("attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data []string `json:"data"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	hasGeneric := false
	for _, tp := range resp.Data {
		if tp == "generic" {
			hasGeneric = true
		}
	}
	if !hasGeneric {
		t.Fatalf("generic doit toujours être présent dans les types, reçu: %v", resp.Data)
	}
}

func TestPluginCreateRepoWithType(t *testing.T) {
	router := setupTestRouter(t)
	token := registerAdmin(t, router, "plugtype-admin", "plugtype@test.com", "password123")

	types := []string{"docker", "helm", "npm"}
	for _, tp := range types {
		t.Run(tp, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{"name": "plugin-repo-" + tp})
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, authReq("POST", "/api/repos/"+tp, bytes.NewReader(body), token))
			if rec.Code != http.StatusCreated && rec.Code != http.StatusBadRequest {
				t.Fatalf("création dépôt %s: attendu 201 ou 400, reçu %d. Body: %s", tp, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestPluginRejectUnknownType(t *testing.T) {
	router := setupTestRouter(t)
	token := registerAdmin(t, router, "unknown-admin", "unknown@test.com", "password123")

	body, _ := json.Marshal(map[string]string{"name": "unknown-repo"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("POST", "/api/repos/foobar", bytes.NewReader(body), token))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("type inconnu: attendu 400, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}
}

func TestPluginDisabledTypeReadOnly(t *testing.T) {
	routerFull := setupTestRouter(t)
	adminToken := registerAdmin(t, routerFull, "degrade-admin", "degrade@test.com", "password123")

	repoBody, _ := json.Marshal(map[string]string{"name": "degrade-repo"})
	rec := httptest.NewRecorder()
	routerFull.ServeHTTP(rec, authReq("POST", "/api/repos/docker", bytes.NewReader(repoBody), adminToken))
	if rec.Code != http.StatusCreated && rec.Code != http.StatusBadRequest {
		t.Fatalf("création dépôt docker: attendu 201 ou 400, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}

	routerReduced := setupTestRouterWithPlugins(t, []string{})
	token := registerAndLogin(t, routerReduced, "degrade-admin", "degrade@test.com", "password123")

	body, _ := json.Marshal(map[string]string{"name": "should-fail"})
	rec = httptest.NewRecorder()
	routerReduced.ServeHTTP(rec, authReq("POST", "/api/repos/docker", bytes.NewReader(body), token))

	if rec.Code == http.StatusCreated {
		t.Fatalf("type désactivé devrait refuser la création, reçu 201")
	}
}
