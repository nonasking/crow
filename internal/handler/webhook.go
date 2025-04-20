package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-jcklk/crow/internal/notion"
	"github.com/go-jcklk/crow/internal/parser"
)

func WebhookHandler(notionClient *notion.NotionClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Message string `json:"message"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		amount, place, cardCompany, err := parser.ParseWebhookAuto(body.Message)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "메시지 파싱 실패: " + err.Error()})
			return
		}

		if err := notionClient.CreateCardRecord(amount, place, cardCompany); err != nil {
			log.Println("Notion 업로드 실패:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Notion 업로드 실패"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}
