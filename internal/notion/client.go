package notion

import (
	"fmt"
	"log"
	"time"
	"strings"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/go-jcklk/crow/internal/config"
	"github.com/go-jcklk/crow/internal/constants"
)

type NotionClient struct {
	client      *resty.Client
	databaseID  string
	accessToken string
	version     string
}

func NewClient(cfg *config.Config) *NotionClient {
	return &NotionClient{
		client:      resty.New(),
		databaseID:  cfg.NotionDatabaseID,
		accessToken: cfg.NotionToken,
		version:     cfg.NotionVersion,
	}
}

// MM/DD 형식의 날짜를 YYYY-MM-DD 형식으로 변환
func formatPaymentDate(paymentDate string) string {
	if paymentDate == "" {
		return time.Now().Format("2006-01-02")
	}

	// MM/DD 형식 파싱
	parts := strings.Split(paymentDate, "/")
	if len(parts) != 2 {
		return time.Now().Format("2006-01-02")
	}

	month, err1 := strconv.Atoi(parts[0])
	day, err2 := strconv.Atoi(parts[1])

	if err1 != nil || err2 != nil {
		return time.Now().Format("2006-01-02")
	}

	// 현재 연도 사용
	currentYear := time.Now().Year()

	// YYYY-MM-DD 형식으로 포맷
	return fmt.Sprintf("%04d-%02d-%02d", currentYear, month, day)
}

func (n *NotionClient) CreateCardRecord(amount int, place, cardCompany, paymentDate string) error {
	// 결제일을 노션 API 형식으로 변환
	formattedDate := formatPaymentDate(paymentDate)

	payload := map[string]interface{}{
		"parent": map[string]interface{}{
			"database_id": n.databaseID,
		},
		"properties": map[string]interface{}{
			"항목": map[string]interface{}{
				"title": []map[string]interface{}{
					{"text": map[string]interface{}{"content": place}},
				},
			},
			"소분류": map[string]interface{}{
				"select": map[string]interface{}{"name": constants.NotionDefaultSubCategory},
			},
			"결제방식": map[string]interface{}{
				"select": map[string]interface{}{"name": cardCompany},
			},
			"날짜": map[string]interface{}{
				"date": map[string]string{"start": formattedDate},
			},
			"수입": map[string]interface{}{"number": 0},
			"지출": map[string]interface{}{"number": amount},
			"대분류": map[string]interface{}{
				"select": map[string]interface{}{"name": constants.NotionDefaultCategory},
			},
			"비고": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"text": map[string]interface{}{"content": constants.NotionDefaultMemo}},
				},
			},
		},
	}

	resp, err := n.client.R().
		SetHeader("Authorization", "Bearer "+n.accessToken).
		SetHeader("Notion-Version", n.version).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post("https://api.notion.com/v1/pages")

	if err != nil {
		return fmt.Errorf("Notion API 요청 실패: %v", err)
	}
	if resp.StatusCode() >= 400 {
		log.Printf("Notion 응답 오류: %s", resp.String())
		return fmt.Errorf("Notion 응답 코드: %d", resp.StatusCode())
	}

	return nil
}