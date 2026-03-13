package api

import (
	"net/http"

	"omnifacts/internal/db"
	authMiddleware "omnifacts/internal/middleware/auth"
	"omnifacts/internal/models"
	"omnifacts/internal/repotype"
	"omnifacts/internal/service"
	"omnifacts/internal/utils"

	"github.com/gorilla/mux"
)

type Services struct {
	ArtefactService   service.ArtefactService
	AuthService       service.AuthService
	RepositoryService service.RepositoryService
}

type DBs struct {
	PermissionDB db.PermissionDB
	RepositoryDB db.RepositoryDB
	APIKeyDB     db.APIKeyDB
}

func SetupRoutes(services Services, dbs DBs, registry *repotype.Registry) *mux.Router {
	router := mux.NewRouter()

	artefactHandler := NewArtefactHandler(services.ArtefactService)
	authHandler := NewAuthHandler(services.AuthService)
	repoHandler := NewRepositoryHandler(services.RepositoryService, registry)

	mid := authMiddleware.NewMiddleware(dbs.APIKeyDB)

	api := router.PathPrefix("/api").Subrouter()

	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteSuccess(w, 200, map[string]string{"status": "OK"})
	}).Methods("GET")

	authenticated := api.PathPrefix("").Subrouter()
	authenticated.Use(mid.RequireAuth)

	authenticated.HandleFunc("/auth/apikeys", authHandler.GenerateAPIKey).Methods("POST")
	authenticated.HandleFunc("/auth/apikeys", authHandler.ListAPIKeys).Methods("GET")

	authenticated.HandleFunc("/repos/types", repoHandler.ListTypes).Methods("GET")

	adminRepos := authenticated.PathPrefix("/repos").Subrouter()
	adminRepos.Use(mid.RequireAdmin)
	adminRepos.HandleFunc("/{repoType}", repoHandler.CreateRepository).Methods("POST")
	adminRepos.HandleFunc("/{repoType}/{repoName}", repoHandler.DeleteRepository).Methods("DELETE")
	adminRepos.HandleFunc("/{repoType}/{repoName}/permissions", repoHandler.SetPermission).Methods("PUT")

	authenticated.HandleFunc("/repos", repoHandler.ListRepositories).Methods("GET")
	authenticated.HandleFunc("/repos/{repoType}/{repoName}", repoHandler.GetRepository).Methods("GET")
	authenticated.HandleFunc("/repos/{repoType}/{repoName}/permissions", repoHandler.ListPermissions).Methods("GET")

	repoRead := authenticated.PathPrefix("/repos/{repoType}/{repoName}").Subrouter()
	repoRead.Use(mid.RequireRepoPermission(dbs.PermissionDB, dbs.RepositoryDB, models.PermRead))
	repoRead.HandleFunc("/artefacts", artefactHandler.GetRepoArtefacts).Methods("GET")
	repoRead.HandleFunc("/artefacts/{id}/content", artefactHandler.GetArtefactContent).Methods("GET")

	repoWrite := authenticated.PathPrefix("/repos/{repoType}/{repoName}").Subrouter()
	repoWrite.Use(mid.RequireRepoPermission(dbs.PermissionDB, dbs.RepositoryDB, models.PermWrite))
	repoWrite.HandleFunc("/artefacts", artefactHandler.CreateArtefact).Methods("POST")
	repoWrite.HandleFunc("/artefacts/{id}", artefactHandler.DeleteArtefact).Methods("DELETE")

	authenticated.HandleFunc("/artefacts", artefactHandler.GetAllArtefacts).Methods("GET")

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))

	return router
}
