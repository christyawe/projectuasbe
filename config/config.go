package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	MongoURI string
	MongoDB  string

	JWTSecret string
}

var AppConfig Config

func LoadConfig() {
	// Load file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ .env file not found — using system environment variables")
	}

	AppConfig = Config{
		AppPort: os.Getenv("APP_PORT"),

		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresPort:     os.Getenv("POSTGRES_PORT"),
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresDB:       os.Getenv("POSTGRES_DB"),

		MongoURI: os.Getenv("MONGO_URI"),
		MongoDB:  os.Getenv("MONGO_DB"),

		JWTSecret: os.Getenv("JWT_SECRET"),
	}
}
