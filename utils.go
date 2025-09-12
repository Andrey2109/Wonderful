package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	APIKey string
}

func loadEnvVariables() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config := Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	}

	if config.APIKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}
	return config
}
