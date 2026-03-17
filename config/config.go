package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DBUser       string
	DBPassword   string
	DBHost       string
	DBPort       string
	DBName       string
	RedisHost    string
	RedisPort    string
	RedisUser    string
	RedisPassword string
	JWTSecret         string
	GroqAPIKey        string
	FirebaseCredentialsPath string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		Port:       getEnv("PORT", "8080"),
		DBUser:     getEnv("DB_USER", "api_user"),
		DBPassword: getEnv("DB_PASSWORD", "apipassword"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBName:     getEnv("DB_NAME", "parejas_db"),
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisUser:     getEnv("REDIS_USER", "default"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		JWTSecret:         getEnv("JWT_SECRET", "super-secret-key-change-me"),
		GroqAPIKey:        getEnv("GROQ_API_KEY", ""),
		FirebaseCredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
