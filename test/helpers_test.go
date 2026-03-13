package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"omnifacts/internal/api"
	"omnifacts/internal/config"
	"omnifacts/internal/db"
	"omnifacts/internal/models"
	"omnifacts/internal/repotype"
	_ "omnifacts/internal/repotype/plugins"
	"omnifacts/internal/service"
	"omnifacts/internal/storage"
	"omnifacts/pkg/database"
)

func setupTestRouter(t *testing.T) http.Handler {
	t.Helper()
	cfg := config.Load()

	database.Connect(cfg)
	database.Migrate()

	storageBackend := storage.NewORASStorage(cfg.RegistryURL, cfg.RegistryPlainHTTP)

	registry := repotype.NewRegistry()
	if err := repotype.LoadPlugins(registry, cfg.EnabledPlugins); err != nil {
		t.Fatalf("Impossible de charger les plugins : %v", err)
	}

	artefactDB := db.NewArtefactDB(database.DB)
	userDB := db.NewUserDB(database.DB)
	repoDB := db.NewRepositoryDB(database.DB)
	permissionDB := db.NewPermissionDB(database.DB)
	apiKeyDB := db.NewAPIKeyDB(database.DB)

	artefactService := service.NewArtefactService(artefactDB, repoDB, storageBackend, cfg.RegistryNamespace, registry)
	authService := service.NewAuthService(userDB, apiKeyDB)
	repoService := service.NewRepositoryService(repoDB, permissionDB, registry)

	return api.SetupRoutes(
		api.Services{
			ArtefactService:   artefactService,
			AuthService:       authService,
			RepositoryService: repoService,
		},
		api.DBs{
			PermissionDB: permissionDB,
			RepositoryDB: repoDB,
			APIKeyDB:     apiKeyDB,
		},
		registry,
	)
}

func setupTestRouterWithPlugins(t *testing.T, enabledPlugins []string) http.Handler {
	t.Helper()
	cfg := config.Load()

	database.Connect(cfg)
	database.Migrate()

	storageBackend := storage.NewORASStorage(cfg.RegistryURL, cfg.RegistryPlainHTTP)

	registry := repotype.NewRegistry()
	if err := repotype.LoadPlugins(registry, enabledPlugins); err != nil {
		t.Fatalf("Impossible de charger les plugins : %v", err)
	}

	artefactDB := db.NewArtefactDB(database.DB)
	userDB := db.NewUserDB(database.DB)
	repoDB := db.NewRepositoryDB(database.DB)
	permissionDB := db.NewPermissionDB(database.DB)
	apiKeyDB := db.NewAPIKeyDB(database.DB)

	artefactService := service.NewArtefactService(artefactDB, repoDB, storageBackend, cfg.RegistryNamespace, registry)
	authService := service.NewAuthService(userDB, apiKeyDB)
	repoService := service.NewRepositoryService(repoDB, permissionDB, registry)

	return api.SetupRoutes(
		api.Services{
			ArtefactService:   artefactService,
			AuthService:       authService,
			RepositoryService: repoService,
		},
		api.DBs{
			PermissionDB: permissionDB,
			RepositoryDB: repoDB,
			APIKeyDB:     apiKeyDB,
		},
		registry,
	)
}

func registerAndLogin(t *testing.T, router http.Handler, username, email, password string) string {
	t.Helper()

	regBody, _ := json.Marshal(map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	})
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated && rec.Code != http.StatusConflict {
		t.Fatalf("register: attendu 201 ou 409, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}

	loginBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	req = httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login: attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Data.Token == "" {
		t.Fatal("login: token vide")
	}
	return resp.Data.Token
}

func makeAdmin(t *testing.T, username string) {
	t.Helper()
	result := database.DB.Model(&models.User{}).Where("username = ?", username).Update("role", "admin")
	if result.Error != nil {
		t.Fatalf("makeAdmin: %v", result.Error)
	}
}

func registerAdmin(t *testing.T, router http.Handler, username, email, password string) string {
	t.Helper()
	registerAndLogin(t, router, username, email, password)
	makeAdmin(t, username)
	return registerAndLogin(t, router, username, email, password)
}

func authReq(method, path string, body io.Reader, token string) *http.Request {
	req := httptest.NewRequest(method, path, body)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}
