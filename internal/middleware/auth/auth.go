package auth

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	"omnifacts/internal/db"
	"omnifacts/internal/models"
	"omnifacts/internal/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type contextKey string

const (
	ContextKeyUser contextKey = "user"
)

// AuthenticatedUser représente l'utilisateur extrait du token/API key
type AuthenticatedUser struct {
	ID       uuid.UUID
	Username string
	Role     models.Role
}

// GetUser extrait l'utilisateur authentifié du contexte
func GetUser(ctx context.Context) *AuthenticatedUser {
	user, ok := ctx.Value(ContextKeyUser).(*AuthenticatedUser)
	if !ok {
		return nil
	}
	return user
}

// Middleware gère l'authentification JWT et par API key
type Middleware struct {
	apiKeyDB db.APIKeyDB
}

func NewMiddleware(apiKeyDB db.APIKeyDB) *Middleware {
	return &Middleware{apiKeyDB: apiKeyDB}
}

// RequireAuth vérifie qu'un utilisateur est authentifié (JWT ou API key)
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := m.authenticate(r)
		if err != nil {
			utils.WriteError(w, http.StatusUnauthorized, "Authentification requise")
			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin vérifie que l'utilisateur est admin
func (m *Middleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r.Context())
		if user == nil {
			utils.WriteError(w, http.StatusUnauthorized, "Authentification requise")
			return
		}

		if user.Role != models.RoleAdmin {
			utils.WriteError(w, http.StatusForbidden, "Droits administrateur requis")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireRepoPermission vérifie que l'utilisateur a la permission requise sur le dépôt
func (m *Middleware) RequireRepoPermission(permissionDB db.PermissionDB, repoDB db.RepositoryDB, level models.PermissionLevel) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUser(r.Context())
			if user == nil {
				utils.WriteError(w, http.StatusUnauthorized, "Authentification requise")
				return
			}

			// Les admins ont tous les droits
			if user.Role == models.RoleAdmin {
				next.ServeHTTP(w, r)
				return
			}

			vars := mux.Vars(r)
			repoName := vars["repoName"]
			if repoName == "" {
				utils.WriteError(w, http.StatusBadRequest, "Nom du dépôt requis")
				return
			}

			repo, err := repoDB.FindByName(repoName)
			if err != nil {
				utils.WriteError(w, http.StatusNotFound, "Dépôt non trouvé")
				return
			}

			perm, err := permissionDB.FindByUserAndRepo(user.ID, repo.ID)
			if err != nil {
				utils.WriteError(w, http.StatusForbidden, "Accès refusé à ce dépôt")
				return
			}

			switch level {
			case models.PermRead:
				if !perm.Level.CanRead() {
					utils.WriteError(w, http.StatusForbidden, "Permission de lecture requise")
					return
				}
			case models.PermWrite:
				if !perm.Level.CanWrite() {
					utils.WriteError(w, http.StatusForbidden, "Permission d'écriture requise")
					return
				}
			case models.PermAdmin:
				if !perm.Level.CanAdmin() {
					utils.WriteError(w, http.StatusForbidden, "Permission d'administration requise")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) authenticate(r *http.Request) (*AuthenticatedUser, error) {
	// 1. Essayer le header Authorization: Bearer <token>
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			return nil, err
		}
		return &AuthenticatedUser{
			ID:       claims.UserID,
			Username: claims.Username,
			Role:     models.Role(claims.Role),
		}, nil
	}

	// 2. Essayer le header X-API-Key
	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "" {
		return m.authenticateAPIKey(apiKey)
	}

	return nil, fmt.Errorf("aucune méthode d'authentification trouvée")
}

func (m *Middleware) authenticateAPIKey(rawKey string) (*AuthenticatedUser, error) {
	keyHash := hashAPIKey(rawKey)
	key, err := m.apiKeyDB.FindByHash(keyHash)
	if err != nil {
		return nil, fmt.Errorf("clé API invalide")
	}

	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("clé API expirée")
	}

	return &AuthenticatedUser{
		ID:       key.User.ID,
		Username: key.User.Username,
		Role:     key.User.Role,
	}, nil
}

func hashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", h)
}
