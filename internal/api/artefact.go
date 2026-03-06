package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"omnifacts/internal/service"
	"omnifacts/internal/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ArtefactHandler struct {
	artefactService service.ArtefactService
}

func NewArtefactHandler(artefactService service.ArtefactService) *ArtefactHandler {
	return &ArtefactHandler{artefactService: artefactService}
}

// CreateArtefact gère la création d'un artefact dans un dépôt via upload multipart
func (h *ArtefactHandler) CreateArtefact(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoName := vars["repoName"]

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Impossible de parser le formulaire multipart")
		return
	}

	name := r.FormValue("name")
	if name == "" {
		utils.WriteError(w, http.StatusBadRequest, "Le nom est requis")
		return
	}

	version := r.FormValue("version")
	if version == "" {
		version = "latest"
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Le fichier est requis")
		return
	}
	defer file.Close()

	var annotations map[string]string
	if annotationsJSON := r.FormValue("annotations"); annotationsJSON != "" {
		if err := json.Unmarshal([]byte(annotationsJSON), &annotations); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Format JSON invalide pour les annotations")
			return
		}
	}

	artefact, err := h.artefactService.CreateArtefact(r.Context(), repoName, name, version, file, annotations)
	if err != nil {
		slog.Error("Erreur lors de la création de l'artefact", "error", err)
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, artefact)
}

// GetRepoArtefacts retourne la liste des artefacts d'un dépôt
func (h *ArtefactHandler) GetRepoArtefacts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoName := vars["repoName"]

	artefacts, err := h.artefactService.RetrieveArtefactsByRepo(repoName)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, artefacts)
}

// GetAllArtefacts retourne tous les artefacts
func (h *ArtefactHandler) GetAllArtefacts(w http.ResponseWriter, r *http.Request) {
	artefactType := r.URL.Query().Get("type")

	var err error
	var artefacts interface{}

	if artefactType != "" {
		artefacts, err = h.artefactService.RetrieveArtefactsByType(artefactType)
	} else {
		artefacts, err = h.artefactService.RetrieveArtefacts()
	}

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, artefacts)
}

// GetArtefactContent télécharge le contenu binaire d'un artefact
func (h *ArtefactHandler) GetArtefactContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID invalide")
		return
	}

	content, err := h.artefactService.GetArtefactContent(r.Context(), id)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Artefact non trouvé")
		return
	}
	defer content.Close()

	if _, err := io.Copy(w, content); err != nil {
		slog.Error("Erreur lors de l'envoi du contenu", "error", err)
	}
}

// DeleteArtefact supprime un artefact
func (h *ArtefactHandler) DeleteArtefact(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "ID invalide")
		return
	}

	if err := h.artefactService.DeleteArtefact(r.Context(), id); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]string{"message": "Artefact supprimé"})
}
