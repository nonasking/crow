package main

import (
    "github.com/gin-gonic/gin"
    "github.com/go-jcklk/crow/internal/config"
    "github.com/go-jcklk/crow/internal/handler"
    "github.com/go-jcklk/crow/internal/notion"
)

func main() {
    config.LoadConfig()

    notionClient := notion.NewNotionClient(config.AppConfig)
    webhookHandler := handler.NewWebhookHandler(notionClient)

    r := gin.Default()
    r.POST("/webhook", webhookHandler.HandleWebhook)
    r.Run(":8080")
}
