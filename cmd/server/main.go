package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"

	"github.com/go-jcklk/crow/internal/handler"
	"github.com/go-jcklk/crow/internal/notion"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := gin.Default()
	notionClient := notion.NewClient()

	r.POST("/webhook", handler.WebhookHandler(notionClient))

	log.Println("서버 시작: http://localhost:" + port)
	err := r.Run(":" + port)
	if err != nil {
		log.Fatal(err)
	}
}
