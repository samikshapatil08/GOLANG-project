package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	Env          string
	DatabaseURL  string
	JWTSecret    string
	GeminiAPIKey string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	dbURL := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	geminiKey := os.Getenv("GEMINI_API_KEY")

	return &Config{
		Port:         port,
		Env:          env,
		DatabaseURL:  dbURL,
		JWTSecret:    jwtSecret,
		GeminiAPIKey: geminiKey,
	}, nil
}
