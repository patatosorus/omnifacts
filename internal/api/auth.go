package api

import (
	"encoding/json"
	"net/http"

	authMiddleware "omnifacts/internal/middleware/auth"
	"omnifacts/internal/service"
	"omnifacts/internal/utils"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Format JSON invalide")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		utils.WriteError(w, http.StatusBadRequest, "Nom d'utilisateur, email et mot de passe requis")
		return
	}

	if len(req.Password) < 8 {
		utils.WriteError(w, http.StatusBadRequest, "Le mot de passe doit contenir au moins 8 caractères")
		return
	}

	user, err := h.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		utils.WriteError(w, http.StatusConflict, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Format JSON invalide")
		return
	}

	if req.Username == "" || req.Password == "" {
		utils.WriteError(w, http.StatusBadRequest, "Nom d'utilisateur et mot de passe requis")
		return
	}

	token, user, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) GenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	user := authMiddleware.GetUser(r.Context())
	if user == nil {
		utils.WriteError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Format JSON invalide")
		return
	}

	if req.Name == "" {
		utils.WriteError(w, http.StatusBadRequest, "Le nom de la clé est requis")
		return
	}

	rawKey, apiKey, err := h.authService.GenerateAPIKey(user.ID, req.Name)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, map[string]interface{}{
		"key":     rawKey,
		"api_key": apiKey,
	})
}

func (h *AuthHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	user := authMiddleware.GetUser(r.Context())
	if user == nil {
		utils.WriteError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	keys, err := h.authService.ListAPIKeys(user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, keys)
}
