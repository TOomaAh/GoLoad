package config

import (
	"os"
)

// Config contient les configurations globales de l'application
type Config struct {
	DatabasePath string
	IsDocker     bool
}

// DefaultConfig retourne la configuration par défaut
func DefaultConfig() Config {
	return Config{
		DatabasePath: "downloads.db",
		IsDocker:     os.Getenv("DOCKER") == "true",
	}
}

// GetConfig retourne la configuration actuelle
func GetConfig() Config {
	config := DefaultConfig()

	// Charger éventuellement d'autres configurations depuis des fichiers ou variables d'environnement
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		config.DatabasePath = dbPath
	}

	return config
}
