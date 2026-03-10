package test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"omnifacts/pkg/database"
)

func TestArtefactUploadAndDownload(t *testing.T) {
	router := setupTestRouter(t)

	adminToken := registerAdmin(t, router, "art-admin", "art-admin@test.com", "password123")
	userToken := registerAndLogin(t, router, "art-user", "art-user@test.com", "password123")

	var user struct {
		ID string `json:"id"`
	}
	database.DB.Raw("SELECT id FROM users WHERE username = ?", "art-user").Scan(&user)

	repoName := "artefact-test-repo"
	repoBody, _ := json.Marshal(map[string]string{"name": repoName})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("POST", "/api/repos/generic", bytes.NewReader(repoBody), adminToken))

	permBody, _ := json.Marshal(map[string]interface{}{"user_id": user.ID, "level": "write"})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("PUT", "/api/repos/generic/"+repoName+"/permissions", bytes.NewReader(permBody), adminToken))

	var artefactID string

	t.Run("upload", func(t *testing.T) {
		content := []byte("contenu de test pour artefact")
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.bin")
		part.Write(content)
		writer.WriteField("name", "mon-artefact")
		writer.WriteField("version", "1.0.0")
		writer.Close()

		req := httptest.NewRequest("POST", "/api/repos/generic/"+repoName+"/artefacts", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+userToken)
		rec := httptest.NewRecorder()
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
		artefactID = uploadResp.Data.ID

		if artefactID == "" {
			t.Fatal("upload: ID artefact vide")
		}
	})

	t.Run("list", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, authReq("GET", "/api/repos/generic/"+repoName+"/artefacts", nil, userToken))
		if rec.Code != http.StatusOK {
			t.Fatalf("liste artefacts: attendu 200, reçu %d", rec.Code)
		}
	})

	t.Run("download", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, authReq("GET", "/api/repos/generic/"+repoName+"/artefacts/"+artefactID+"/content", nil, userToken))
		if rec.Code != http.StatusOK {
			t.Fatalf("download: attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
		}
		downloaded, _ := io.ReadAll(rec.Body)
		expected := []byte("contenu de test pour artefact")
		if !bytes.Equal(downloaded, expected) {
			t.Errorf("contenu différent: attendu %q, reçu %q", expected, downloaded)
		}
	})

	t.Run("delete", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, authReq("DELETE", "/api/repos/generic/"+repoName+"/artefacts/"+artefactID, nil, userToken))
		if rec.Code != http.StatusOK {
			t.Fatalf("suppression: attendu 200, reçu %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestArtefactUploadWithoutPermission(t *testing.T) {
	router := setupTestRouter(t)

	adminToken := registerAdmin(t, router, "noperm-admin", "noperm-admin@test.com", "password123")
	userToken := registerAndLogin(t, router, "noperm-user", "noperm-user@test.com", "password123")

	repoName := "noperm-repo"
	repoBody, _ := json.Marshal(map[string]string{"name": repoName})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authReq("POST", "/api/repos/generic", bytes.NewReader(repoBody), adminToken))

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.bin")
	part.Write([]byte("test"))
	writer.WriteField("name", "blocked")
	writer.WriteField("version", "1.0.0")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/repos/generic/"+repoName+"/artefacts", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("upload sans permission: attendu 403, reçu %d. Body: %s", rec.Code, rec.Body.String())
	}
}
