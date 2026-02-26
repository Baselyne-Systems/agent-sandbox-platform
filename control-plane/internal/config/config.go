package config

import "os"

type Config struct {
	DatabaseURL string
	GRPCPort    string
	LogLevel    string
}

func LoadFromEnv() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/sandbox?sslmode=disable"),
		GRPCPort:    getEnv("GRPC_PORT", "50051"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
