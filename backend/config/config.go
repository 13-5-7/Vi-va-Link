package config

import (
	"log"
	"os"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	OSRMBaseURL string
	Port        string
	AdminKey    string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	osrmURL := os.Getenv("OSRM_BASE_URL")
	if osrmURL == "" {
		osrmURL = "https://router.project-osrm.org"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required but not set")
	}
	adminKey := os.Getenv("ADMIN_KEY")
	if adminKey == "" {
		log.Println("WARNING: ADMIN_KEY is not set. Admin endpoints will be disabled.")
	}
	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   jwtSecret,
		OSRMBaseURL: osrmURL,
		Port:        port,
		AdminKey:    adminKey,
	}
}
