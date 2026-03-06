package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"omnifacts/internal/api"
	"omnifacts/internal/config"
	"omnifacts/internal/db"
	"omnifacts/internal/service"
	"omnifacts/internal/storage"
	"omnifacts/pkg/database"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	cfg := config.Load()

	database.Connect(cfg)
	database.Migrate()

	storageBackend := storage.NewORASStorage(cfg.RegistryURL, cfg.RegistryPlainHTTP)

	// Initialiser les couches DB
	artefactDB := db.NewArtefactDB(database.DB)
	userDB := db.NewUserDB(database.DB)
	repoDB := db.NewRepositoryDB(database.DB)
	permissionDB := db.NewPermissionDB(database.DB)
	apiKeyDB := db.NewAPIKeyDB(database.DB)

	// Initialiser les services
	artefactService := service.NewArtefactService(artefactDB, repoDB, storageBackend, cfg.RegistryNamespace)
	authService := service.NewAuthService(userDB, apiKeyDB)
	repoService := service.NewRepositoryService(repoDB, permissionDB)

	// Configurer les routes
	router := api.SetupRoutes(
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

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	log.Printf("Serveur démarré sur le port %s", cfg.Port)
	log.Printf("Registre OCI : %s (namespace: %s)", cfg.RegistryURL, cfg.RegistryNamespace)
	log.Fatal(server.ListenAndServe())
}
