package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"omnifacts/pkg/database"
)

func TestRepositoryLifecycle(t *testing.T) {
	router := setupTestRouter(t)
	token := registerAdmin(t, router, "repo-admin", "repo-admin@test.com", "password123")

	repoName := fmt.Sprintf("repo-lifecycle-%s", t.Name())

	t.Run("create", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"name":         repoName,
			"description":  "Dépôt de test",
			"storage_mode": "local",
		})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, authReq("POST", "/api/repos/generic", bytes.NewReader(body), token))
		if rec.Code != http.StatusCreated {
			t.Fatalf("création dépôt: attendu 201, reçu %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("list", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, authReq("GET", "/api/repos", nil, token))
		if rec.Code != http.StatusOK {
			t.Fatalf("liste dépôts: attendu 200, reçu %d", rec.Code)
		}
	})

	t.Run("get", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, authReq("GET", "/api/repos/generic/"+repoName, nil, token))
		if rec.Code != http.StatusOK {
			t.Fatalf("détail dépôt: attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("delete", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, authReq("DELETE", "/api/repos/generic/"+repoName, nil, token))
		if rec.Code != http.StatusOK {
			t.Fatalf("suppression dépôt: attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestRepositoryCreationRequiresAdmin(t *testing.T) {
	router := setupTestRouter(t)
	token := registerAndLogin(t, router, "nonadmin-repo", "nonadmin-repo@test.com", "password123")

	body, _ := json.Marshal(map[string]string{
		"name": "should-fail-repo",
	})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("POST", "/api/repos/generic", bytes.NewReader(body), token))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("création par non-admin: attendu 403, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}
}

func TestRepositoryPermissions(t *testing.T) {
	router := setupTestRouter(t)

	adminToken := registerAdmin(t, router, "perm-admin", "perm-admin@test.com", "password123")
	userToken := registerAndLogin(t, router, "perm-user", "perm-user@test.com", "password123")

	var user struct {
		ID string `json:"id"`
	}
	database.DB.Raw("SELECT id FROM users WHERE username = ?", "perm-user").Scan(&user)

	repoBody, _ := json.Marshal(map[string]string{"name": "perm-test-repo"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("POST", "/api/repos/generic", bytes.NewReader(repoBody), adminToken))
	if rec.Code != http.StatusCreated && rec.Code != http.StatusBadRequest {
		t.Fatalf("création dépôt: attendu 201, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}

	t.Run("denied without permission", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, authReq("GET", "/api/repos/generic/perm-test-repo/artefacts", nil, userToken))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("sans permission: attendu 403, reçu %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("grant read", func(t *testing.T) {
		permBody, _ := json.Marshal(map[string]interface{}{
			"user_id": user.ID,
			"level":   "read",
		})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, authReq("PUT", "/api/repos/generic/perm-test-repo/permissions", bytes.NewReader(permBody), adminToken))
		if rec.Code != http.StatusOK {
			t.Fatalf("set permission: attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("allowed with permission", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, authReq("GET", "/api/repos/generic/perm-test-repo/artefacts", nil, userToken))
		if rec.Code != http.StatusOK {
			t.Fatalf("avec permission read: attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestRepositoryWrongType(t *testing.T) {
	router := setupTestRouter(t)
	token := registerAdmin(t, router, "wrongtype-admin", "wrongtype@test.com", "password123")

	body, _ := json.Marshal(map[string]string{"name": "wrong-type-repo"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("POST", "/api/repos/generic", bytes.NewReader(body), token))
	if rec.Code != http.StatusCreated && rec.Code != http.StatusBadRequest {
		t.Fatalf("attendu 201 ou 400, reçu %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("GET", "/api/repos/docker/wrong-type-repo", nil, token))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("mauvais type dans URL: attendu 404, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}
}
