package api

import (
	"net/http"

	"omnifacts/internal/db"
	authMiddleware "omnifacts/internal/middleware/auth"
	"omnifacts/internal/models"
	"omnifacts/internal/service"
	"omnifacts/internal/utils"

	"github.com/gorilla/mux"
)

// Services regroupe tous les services nécessaires au routeur
type Services struct {
	ArtefactService   service.ArtefactService
	AuthService       service.AuthService
	RepositoryService service.RepositoryService
}

// DBs regroupe les accès DB nécessaires au middleware RBAC
type DBs struct {
	PermissionDB db.PermissionDB
	RepositoryDB db.RepositoryDB
	APIKeyDB     db.APIKeyDB
}

func SetupRoutes(services Services, dbs DBs) *mux.Router {
	router := mux.NewRouter()

	artefactHandler := NewArtefactHandler(services.ArtefactService)
	authHandler := NewAuthHandler(services.AuthService)
	repoHandler := NewRepositoryHandler(services.RepositoryService)

	mid := authMiddleware.NewMiddleware(dbs.APIKeyDB)

	api := router.PathPrefix("/api").Subrouter()

	// Routes publiques (pas d'auth)
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteSuccess(w, 200, map[string]string{"status": "OK"})
	}).Methods("GET")

	// Routes authentifiées
	authenticated := api.PathPrefix("").Subrouter()
	authenticated.Use(mid.RequireAuth)

	// API keys
	authenticated.HandleFunc("/auth/apikeys", authHandler.GenerateAPIKey).Methods("POST")
	authenticated.HandleFunc("/auth/apikeys", authHandler.ListAPIKeys).Methods("GET")

	// Gestion des dépôts (admin uniquement pour créer/supprimer)
	adminRepos := authenticated.PathPrefix("/repos").Subrouter()
	adminRepos.Use(mid.RequireAdmin)
	adminRepos.HandleFunc("", repoHandler.CreateRepository).Methods("POST")
	adminRepos.HandleFunc("/{repoName}", repoHandler.DeleteRepository).Methods("DELETE")
	adminRepos.HandleFunc("/{repoName}/permissions", repoHandler.SetPermission).Methods("PUT")

	// Lecture des dépôts (tout utilisateur authentifié)
	authenticated.HandleFunc("/repos", repoHandler.ListRepositories).Methods("GET")
	authenticated.HandleFunc("/repos/{repoName}", repoHandler.GetRepository).Methods("GET")
	authenticated.HandleFunc("/repos/{repoName}/permissions", repoHandler.ListPermissions).Methods("GET")

	// Artefacts dans un dépôt — avec vérification RBAC par dépôt
	repoRead := authenticated.PathPrefix("/repos/{repoName}").Subrouter()
	repoRead.Use(mid.RequireRepoPermission(dbs.PermissionDB, dbs.RepositoryDB, models.PermRead))
	repoRead.HandleFunc("/artefacts", artefactHandler.GetRepoArtefacts).Methods("GET")
	repoRead.HandleFunc("/artefacts/{id}/content", artefactHandler.GetArtefactContent).Methods("GET")

	repoWrite := authenticated.PathPrefix("/repos/{repoName}").Subrouter()
	repoWrite.Use(mid.RequireRepoPermission(dbs.PermissionDB, dbs.RepositoryDB, models.PermWrite))
	repoWrite.HandleFunc("/artefacts", artefactHandler.CreateArtefact).Methods("POST")
	repoWrite.HandleFunc("/artefacts/{id}", artefactHandler.DeleteArtefact).Methods("DELETE")

	// Listing global des artefacts (tout utilisateur authentifié)
	authenticated.HandleFunc("/artefacts", artefactHandler.GetAllArtefacts).Methods("GET")

	// Fallback pour le frontend (doit être en dernier)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))

	return router
}
