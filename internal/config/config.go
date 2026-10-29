package config

import (
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    DatabaseURL string
    JWTSecret   string
    Port        string
    UploadDir   string
}

func Load() Config {
    _ = godotenv.Load() // abaikan error, misal saat deploy tanpa berkas .env

    return Config{
        DatabaseURL: os.Getenv("DATABASE_URL"),
        JWTSecret:   os.Getenv("JWT_SECRET"),
        Port:        getOrDefault("PORT", "8080"),
        UploadDir:   getOrDefault("UPLOAD_DIR", "./uploads"),
    }
}

func getOrDefault(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}