package config

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    NotionToken     string
    NotionDatabaseID string
    NotionVersion    string
}

var AppConfig *Config

func LoadConfig() {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using environment variables.")
    }

    AppConfig = &Config{
        NotionToken:     os.Getenv("NOTION_TOKEN"),
        NotionDatabaseID: os.Getenv("NOTION_DATABASE_ID"),
        NotionVersion:    os.Getenv("NOTION_VERSION"),
    }

    if AppConfig.NotionToken == "" || AppConfig.NotionDatabaseID == "" {
        log.Fatal("Required environment variables are missing.")
    }
}
