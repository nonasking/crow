package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-jcklk/crow/internal/notion"
	"github.com/go-jcklk/crow/internal/parser"
	"github.com/go-jcklk/crow/internal/constants"
)

type WebhookHandler struct {
	notionClient *notion.NotionClient
}

func NewWebhookHandler(client *notion.NotionClient) gin.HandlerFunc {
	h := &WebhookHandler{notionClient: client}
	return h.handle
}

func (h *WebhookHandler) handle(c *gin.Context) {
	var body struct {
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrMsgInvalidJSON})
		return
	}

	amount, place, cardCompany, err := parser.ParseWebhookAuto(body.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrMsgParseFailed + err.Error()})
		return
	}

	if err := h.notionClient.CreateCardRecord(amount, place, cardCompany); err != nil {
		log.Println(constants.ErrMsgNotionFailed, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrMsgNotionFailed})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
