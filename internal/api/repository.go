package api

import (
	"encoding/json"
	"net/http"

	"omnifacts/internal/models"
	"omnifacts/internal/service"
	"omnifacts/internal/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type RepositoryHandler struct {
	repoService service.RepositoryService
}

func NewRepositoryHandler(repoService service.RepositoryService) *RepositoryHandler {
	return &RepositoryHandler{repoService: repoService}
}

func (h *RepositoryHandler) CreateRepository(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string             `json:"name"`
		Description  string             `json:"description"`
		ArtefactType string             `json:"artefact_type"`
		StorageMode  models.StorageMode `json:"storage_mode"`
		UpstreamURL  string             `json:"upstream_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Format JSON invalide")
		return
	}

	if req.Name == "" || req.ArtefactType == "" {
		utils.WriteError(w, http.StatusBadRequest, "Nom et type d'artefact requis")
		return
	}

	if req.StorageMode == "" {
		req.StorageMode = models.StorageLocal
	}

	repo, err := h.repoService.Create(req.Name, req.Description, req.ArtefactType, req.StorageMode, req.UpstreamURL)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, repo)
}

func (h *RepositoryHandler) ListRepositories(w http.ResponseWriter, r *http.Request) {
	repos, err := h.repoService.List()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, repos)
}

func (h *RepositoryHandler) GetRepository(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["repoName"]

	repo, err := h.repoService.FindByName(name)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Dépôt non trouvé")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, repo)
}

func (h *RepositoryHandler) DeleteRepository(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["repoName"]

	repo, err := h.repoService.FindByName(name)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Dépôt non trouvé")
		return
	}

	if err := h.repoService.Delete(repo.ID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]string{"message": "Dépôt supprimé"})
}

func (h *RepositoryHandler) SetPermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoName := vars["repoName"]

	var req struct {
		UserID uuid.UUID               `json:"user_id"`
		Level  models.PermissionLevel  `json:"level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Format JSON invalide")
		return
	}

	repo, err := h.repoService.FindByName(repoName)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Dépôt non trouvé")
		return
	}

	if err := h.repoService.SetPermission(req.UserID, repo.ID, req.Level); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]string{"message": "Permission mise à jour"})
}

func (h *RepositoryHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoName := vars["repoName"]

	repo, err := h.repoService.FindByName(repoName)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Dépôt non trouvé")
		return
	}

	perms, err := h.repoService.ListPermissions(repo.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, perms)
}
