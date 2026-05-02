package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-jcklk/crow/internal/constants"
	"github.com/go-jcklk/crow/internal/parser"
)

// CardRecorder는 파싱된 카드 결제 내역을 외부 저장소에 기록하는 동작을 정의한다.
type CardRecorder interface {
	CreateCardRecord(amount int, place, cardCompany, paymentDate string) error
}

type WebhookHandler struct {
	recorder CardRecorder
}

func NewWebhookHandler(recorder CardRecorder) gin.HandlerFunc {
	h := &WebhookHandler{recorder: recorder}
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

	amount, place, cardCompany, paymentDate, err := parser.ParseWebhookAuto(body.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrMsgParseFailed + err.Error()})
		return
	}

	if err := h.recorder.CreateCardRecord(amount, place, cardCompany, paymentDate); err != nil {
		log.Println(constants.ErrMsgNotionFailed, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrMsgNotionFailed})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
