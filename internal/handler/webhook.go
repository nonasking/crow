package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/go-jcklk/crow/internal/model"
    "github.com/go-jcklk/crow/internal/notion"

    "fmt"
)

type WebhookHandler struct {
    NotionClient *notion.NotionClient
}

func NewWebhookHandler(nc *notion.NotionClient) *WebhookHandler {
    return &WebhookHandler{
        NotionClient: nc,
    }
}

func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
    var payload model.WebhookPayload

    fmt.Println(payload)

// 웹훅 -> 노션 업로드

//     if err := c.ShouldBindJSON(&payload); err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
//         return
//     }

//     err := h.NotionClient.CreatePage(payload.Title, payload.Message)
//     if err != nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Notion page"})
//         return
//     }

    c.JSON(http.StatusOK, gin.H{"status": "success"})
}
