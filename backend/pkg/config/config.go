package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	GeminiKey      string
	QdrantHost     string
	QdrantPort     int
	DBPath         string
	Port           string
	EmbedWorkers   int
	ChunkSize      int
	ChunkOverlap   int
	TopK           int
	MaxHistoryTurns int
	UploadDir      string
	CollectionName string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		GeminiKey:       mustEnv("GEMINI_API_KEY"),
		QdrantHost:      getEnv("QDRANT_HOST", "localhost"),
		QdrantPort:      getEnvInt("QDRANT_PORT", 6334),
		DBPath:          getEnv("DB_PATH", "./notebooklm.db"),
		Port:            getEnv("PORT", "8080"),
		EmbedWorkers:    getEnvInt("EMBED_WORKERS", 8),
		ChunkSize:       getEnvInt("CHUNK_SIZE", 1000),
		ChunkOverlap:    getEnvInt("CHUNK_OVERLAP", 200),
		TopK:            getEnvInt("TOP_K", 5),
		MaxHistoryTurns: getEnvInt("MAX_HISTORY_TURNS", 10),
		UploadDir:       getEnv("UPLOAD_DIR", "/tmp/notebooklm_uploads"),
		CollectionName:  getEnv("QDRANT_COLLECTION", "notebooklm"),
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}
