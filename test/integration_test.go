package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"omnifacts/internal/api"
	"omnifacts/internal/config"
	"omnifacts/internal/db"
	"omnifacts/internal/models"
	"omnifacts/internal/service"
	"omnifacts/internal/storage"
	"omnifacts/pkg/database"
)

// setupTestRouter initialise le routeur complet pour les tests d'intégration
func setupTestRouter(t *testing.T) http.Handler {
	t.Helper()
	cfg := config.Load()

	database.Connect(cfg)
	database.Migrate()

	storageBackend := storage.NewORASStorage(cfg.RegistryURL, cfg.RegistryPlainHTTP)
	artefactDB := db.NewArtefactDB(database.DB)
	userDB := db.NewUserDB(database.DB)
	repoDB := db.NewRepositoryDB(database.DB)
	permissionDB := db.NewPermissionDB(database.DB)
	apiKeyDB := db.NewAPIKeyDB(database.DB)

	artefactService := service.NewArtefactService(artefactDB, repoDB, storageBackend, cfg.RegistryNamespace)
	authService := service.NewAuthService(userDB, apiKeyDB)
	repoService := service.NewRepositoryService(repoDB, permissionDB)

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
	)
}

// registerUser crée un utilisateur et retourne le token JWT
func registerAndLogin(t *testing.T, router http.Handler, username, email, password string) string {
	t.Helper()

	// Inscription (on ignore 409 si l'utilisateur existe déjà)
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

	// Connexion
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

// makeAdmin promeut un utilisateur en admin directement en base
func makeAdmin(t *testing.T, username string) {
	t.Helper()
	result := database.DB.Model(&models.User{}).Where("username = ?", username).Update("role", "admin")
	if result.Error != nil {
		t.Fatalf("makeAdmin: %v", result.Error)
	}
}

// authReq crée une requête HTTP avec le header Authorization
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

// --- TESTS ---

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter(t)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest("GET", "/api/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("attendu 200, reçu %d", rec.Code)
	}
}

func TestAuthRegisterAndLogin(t *testing.T) {
	router := setupTestRouter(t)
	token := registerAndLogin(t, router, "auth-test-user", "auth@test.com", "password123")

	// Accès protégé avec token → 200
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("GET", "/api/repos", nil, token))
	if rec.Code != http.StatusOK {
		t.Fatalf("accès authentifié: attendu 200, reçu %d", rec.Code)
	}

	// Accès sans token → 401
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest("GET", "/api/repos", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("accès non-authentifié: attendu 401, reçu %d", rec.Code)
	}
}

func TestAuthRegisterValidation(t *testing.T) {
	router := setupTestRouter(t)

	tests := []struct {
		name string
		body map[string]string
		code int
	}{
		{"champs manquants", map[string]string{"username": "x"}, http.StatusBadRequest},
		{"mot de passe court", map[string]string{"username": "shortpw", "email": "s@t.com", "password": "123"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tt.code {
				t.Errorf("attendu %d, reçu %d. Body: %s", tt.code, rec.Code, rec.Body.String())
			}
		})
	}

	// Doublon
	registerAndLogin(t, router, "dup-user", "dup@test.com", "password123")
	b, _ := json.Marshal(map[string]string{"username": "dup-user", "email": "dup2@test.com", "password": "password123"})
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("doublon: attendu 409, reçu %d", rec.Code)
	}
}

func TestAPIKeyGeneration(t *testing.T) {
	router := setupTestRouter(t)
	token := registerAndLogin(t, router, "apikey-user", "apikey@test.com", "password123")

	// Générer une clé API
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

	// Utiliser la clé API pour accéder à un endpoint protégé
	req := httptest.NewRequest("GET", "/api/repos", nil)
	req.Header.Set("X-API-Key", keyResp.Data.Key)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("accès via API key: attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Lister les clés API
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("GET", "/api/auth/apikeys", nil, token))
	if rec.Code != http.StatusOK {
		t.Fatalf("liste clés API: attendu 200, reçu %d", rec.Code)
	}
}

func TestRepositoryLifecycle(t *testing.T) {
	router := setupTestRouter(t)
	token := registerAndLogin(t, router, "repo-admin", "repo-admin@test.com", "password123")
	makeAdmin(t, "repo-admin")
	// Re-login pour obtenir un token avec le rôle admin
	token = registerAndLogin(t, router, "repo-admin", "repo-admin@test.com", "password123")

	repoName := fmt.Sprintf("repo-lifecycle-%d", testing.AllocsPerRun(0, func() {}))

	// Créer un dépôt
	body, _ := json.Marshal(map[string]string{
		"name":          repoName,
		"description":   "Dépôt de test",
		"artefact_type": "generic",
		"storage_mode":  "local",
	})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("POST", "/api/repos", bytes.NewReader(body), token))

	if rec.Code != http.StatusCreated {
		t.Fatalf("création dépôt: attendu 201, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Lister les dépôts
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("GET", "/api/repos", nil, token))
	if rec.Code != http.StatusOK {
		t.Fatalf("liste dépôts: attendu 200, reçu %d", rec.Code)
	}

	// Détail du dépôt
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("GET", "/api/repos/"+repoName, nil, token))
	if rec.Code != http.StatusOK {
		t.Fatalf("détail dépôt: attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Supprimer le dépôt
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("DELETE", "/api/repos/"+repoName, nil, token))
	if rec.Code != http.StatusOK {
		t.Fatalf("suppression dépôt: attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}
}

func TestRepositoryCreationRequiresAdmin(t *testing.T) {
	router := setupTestRouter(t)
	token := registerAndLogin(t, router, "nonadmin-repo", "nonadmin-repo@test.com", "password123")

	body, _ := json.Marshal(map[string]string{
		"name":          "should-fail-repo",
		"artefact_type": "generic",
	})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("POST", "/api/repos", bytes.NewReader(body), token))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("création par non-admin: attendu 403, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}
}

func TestRepositoryPermissions(t *testing.T) {
	router := setupTestRouter(t)

	// Admin
	adminToken := registerAndLogin(t, router, "perm-admin", "perm-admin@test.com", "password123")
	makeAdmin(t, "perm-admin")
	adminToken = registerAndLogin(t, router, "perm-admin", "perm-admin@test.com", "password123")

	// Utilisateur normal
	userToken := registerAndLogin(t, router, "perm-user", "perm-user@test.com", "password123")

	// Récupérer l'ID de l'utilisateur
	var user models.User
	database.DB.First(&user, "username = ?", "perm-user")

	// Créer un dépôt
	repoBody, _ := json.Marshal(map[string]string{
		"name": "perm-test-repo", "artefact_type": "generic",
	})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("POST", "/api/repos", bytes.NewReader(repoBody), adminToken))
	if rec.Code != http.StatusCreated && rec.Code != http.StatusBadRequest {
		t.Fatalf("création dépôt: attendu 201, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Sans permission → 403 sur les artefacts
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("GET", "/api/repos/perm-test-repo/artefacts", nil, userToken))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("sans permission: attendu 403, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Accorder la permission read
	permBody, _ := json.Marshal(map[string]interface{}{
		"user_id": user.ID.String(),
		"level":   "read",
	})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("PUT", "/api/repos/perm-test-repo/permissions", bytes.NewReader(permBody), adminToken))
	if rec.Code != http.StatusOK {
		t.Fatalf("set permission: attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Avec permission read → 200
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("GET", "/api/repos/perm-test-repo/artefacts", nil, userToken))
	if rec.Code != http.StatusOK {
		t.Fatalf("avec permission read: attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}
}

func TestArtefactUploadAndDownload(t *testing.T) {
	router := setupTestRouter(t)

	// Admin crée un dépôt et donne les droits write
	adminToken := registerAndLogin(t, router, "art-admin", "art-admin@test.com", "password123")
	makeAdmin(t, "art-admin")
	adminToken = registerAndLogin(t, router, "art-admin", "art-admin@test.com", "password123")

	userToken := registerAndLogin(t, router, "art-user", "art-user@test.com", "password123")

	var user models.User
	database.DB.First(&user, "username = ?", "art-user")

	repoName := "artefact-test-repo"
	repoBody, _ := json.Marshal(map[string]string{"name": repoName, "artefact_type": "generic"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("POST", "/api/repos", bytes.NewReader(repoBody), adminToken))

	// Permission write
	permBody, _ := json.Marshal(map[string]interface{}{"user_id": user.ID.String(), "level": "write"})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("PUT", "/api/repos/"+repoName+"/permissions", bytes.NewReader(permBody), adminToken))

	// Upload
	content := []byte("contenu de test pour artefact")
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.bin")
	part.Write(content)
	writer.WriteField("name", "mon-artefact")
	writer.WriteField("version", "1.0.0")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/repos/"+repoName+"/artefacts", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("upload: attendu 201, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}

	var uploadResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	json.NewDecoder(rec.Body).Decode(&uploadResp)
	artefactID := uploadResp.Data.ID

	if artefactID == "" {
		t.Fatal("upload: ID artefact vide")
	}

	// Lister les artefacts du dépôt
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("GET", "/api/repos/"+repoName+"/artefacts", nil, userToken))
	if rec.Code != http.StatusOK {
		t.Fatalf("liste artefacts: attendu 200, reçu %d", rec.Code)
	}

	// Télécharger le contenu
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("GET", "/api/repos/"+repoName+"/artefacts/"+artefactID+"/content", nil, userToken))
	if rec.Code != http.StatusOK {
		t.Fatalf("download: attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}
	downloaded, _ := io.ReadAll(rec.Body)
	if !bytes.Equal(downloaded, content) {
		t.Errorf("contenu différent: attendu %q, reçu %q", content, downloaded)
	}

	// Supprimer l'artefact
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("DELETE", "/api/repos/"+repoName+"/artefacts/"+artefactID, nil, userToken))
	if rec.Code != http.StatusOK {
		t.Fatalf("suppression: attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}
}

func TestArtefactUploadWithoutPermission(t *testing.T) {
	router := setupTestRouter(t)

	adminToken := registerAndLogin(t, router, "noperm-admin", "noperm-admin@test.com", "password123")
	makeAdmin(t, "noperm-admin")
	adminToken = registerAndLogin(t, router, "noperm-admin", "noperm-admin@test.com", "password123")

	userToken := registerAndLogin(t, router, "noperm-user", "noperm-user@test.com", "password123")

	repoName := "noperm-repo"
	repoBody, _ := json.Marshal(map[string]string{"name": repoName, "artefact_type": "generic"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("POST", "/api/repos", bytes.NewReader(repoBody), adminToken))

	// Upload sans permission → 403
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.bin")
	part.Write([]byte("test"))
	writer.WriteField("name", "blocked")
	writer.WriteField("version", "1.0.0")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/repos/"+repoName+"/artefacts", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("upload sans permission: attendu 403, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}
}

func TestUnauthenticatedAccess(t *testing.T) {
	router := setupTestRouter(t)

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/repos"},
		{"POST", "/api/repos"},
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
