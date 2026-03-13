package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"omnifacts/internal/api"
	"omnifacts/internal/config"
	"omnifacts/internal/db"
	"omnifacts/internal/repotype"
	_ "omnifacts/internal/repotype/plugins"
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

	registry := repotype.NewRegistry()
	if err := repotype.LoadPlugins(registry, cfg.EnabledPlugins); err != nil {
		log.Fatalf("Impossible de charger les plugins : %v", err)
	}

	artefactDB := db.NewArtefactDB(database.DB)
	userDB := db.NewUserDB(database.DB)
	repoDB := db.NewRepositoryDB(database.DB)
	permissionDB := db.NewPermissionDB(database.DB)
	apiKeyDB := db.NewAPIKeyDB(database.DB)

	artefactService := service.NewArtefactService(artefactDB, repoDB, storageBackend, cfg.RegistryNamespace, registry)
	authService := service.NewAuthService(userDB, apiKeyDB)
	repoService := service.NewRepositoryService(repoDB, permissionDB, registry)

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
		registry,
	)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	log.Printf("Serveur démarré sur le port %s", cfg.Port)
	log.Printf("Registre OCI : %s (namespace: %s)", cfg.RegistryURL, cfg.RegistryNamespace)
	log.Fatal(server.ListenAndServe())
}
