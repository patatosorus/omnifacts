package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Configuration du registre OCI (backend de stockage)
	RegistryURL       string // URL du registre OCI (ex: localhost:5000)
	RegistryNamespace string // Namespace par défaut pour les artefacts (ex: "omnifacts")
	RegistryPlainHTTP bool   // Utiliser HTTP au lieu de HTTPS (pour le développement)
}

func Load() *Config {
	// Charger le fichier .env
	err := godotenv.Load()
	if err != nil {
		slog.Error("Fichier .env non trouvé, utilisation des variables d'environnement")
	}

	return &Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "postgres"),

		RegistryURL:       getEnv("REGISTRY_URL", "localhost:5000"),
		RegistryNamespace: getEnv("REGISTRY_NAMESPACE", "omnifacts"),
		RegistryPlainHTTP: getEnv("REGISTRY_PLAIN_HTTP", "true") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
