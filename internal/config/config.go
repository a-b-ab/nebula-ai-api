package config

import (
	"os"
)

type Config struct {
	// 服务器配置
	ServerPort string

	// 数据库配置
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Redis 配置
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// AI 配置
	DeepSeekAPIKey string

	// 发布器配置
	PublisherURL string
}

func Load() *Config {
	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "nebula"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		DeepSeekAPIKey: getEnv("DEEPSEEK_API_KEY", ""),

		PublisherURL: getEnv("PUBLISHER_URL", "http://localhost:3001"),
	}
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}
