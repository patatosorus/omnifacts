package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	RegistryURL       string
	RegistryNamespace string
	RegistryPlainHTTP bool

	EnabledPlugins []string
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

		EnabledPlugins: parsePluginList(getEnv("ENABLED_PLUGINS", "docker,oci,helm,pypi,npm,terraform,maven")),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parsePluginList(raw string) []string {
	if raw == "" {
		return nil
	}
	var plugins []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			plugins = append(plugins, p)
		}
	}
	return plugins
}
