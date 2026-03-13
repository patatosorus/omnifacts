package database

import (
	"fmt"
	"log/slog"

	"omnifacts/internal/config"
	"omnifacts/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.Config) {
	var err error
	slog.Info("Connexion à la base de données",
		"host", cfg.DBHost,
		"user", cfg.DBUser,
		"name", cfg.DBName,
		"port", cfg.DBPort,
	)

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
	)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		slog.Error("Erreur de connexion à la base de données", "error", err.Error())
	}

	slog.Info("Connexion à la base de données établie")
}

func Migrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.APIKey{},
		&models.Layer{},
		&models.OCIManifest{},
		&models.Repository{},
		&models.Artefact{},
		&models.Permission{},
	)
	if err != nil {
		slog.Error("Une erreur est survenue pendant l'application des migrations", "error", err.Error())
	}
	slog.Info("Migrations effectuées")
}
