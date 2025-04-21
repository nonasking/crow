package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-jcklk/crow/internal/config"
	"github.com/go-jcklk/crow/internal/handler"
	"github.com/go-jcklk/crow/internal/notion"
)

func main() {
	config.LoadConfig()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := gin.Default()
	notionClient := notion.NewClient(config.AppConfig)

	r.POST("/webhook", handler.NewWebhookHandler(notionClient))

	log.Printf("서버 시작: http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
