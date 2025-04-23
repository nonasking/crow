package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	NotionToken      string
	NotionDatabaseID string
	NotionVersion    string
}

var AppConfig *Config

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables.")
	}

	AppConfig = &Config{
		NotionToken:      os.Getenv("NOTION_TOKEN"),
		NotionDatabaseID: os.Getenv("NOTION_DATABASE_ID"),
		NotionVersion:    os.Getenv("NOTION_VERSION"),
	}

	if AppConfig.NotionToken == "" || AppConfig.NotionDatabaseID == "" || AppConfig.NotionVersion == "" {
		log.Fatal("환경변수(NOTION_TOKEN, NOTION_DATABASE_ID, NOTION_VERSION)가 누락되었습니다.")
	}
}
