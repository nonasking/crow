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

// SMS의 날짜·시간은 한국 시간이다. 서버(Render 등)는 UTC로 돌기 때문에
// time.Now()를 그대로 쓰면 KST 00:00~08:59 결제의 연도와 폴백 날짜가 어긋난다.
var kst = time.FixedZone("KST", 9*60*60)

// MM/DD 형식의 날짜를 YYYY-MM-DD 형식으로 변환
func formatPaymentDate(paymentDate string) string {
	return formatPaymentDateAt(paymentDate, time.Now().In(kst))
}

// formatPaymentDateAt은 now(KST)를 기준으로 MM/DD에 연도를 붙인다.
// SMS에는 연도가 없으므로 올해로 붙이되, 미래 날짜가 되면 작년 결제로 본다
// (예: 1월 1일에 받은 12/31 결제). 형식이 깨졌거나 존재하지 않는 날짜면 오늘 날짜를 쓴다.
func formatPaymentDateAt(paymentDate string, now time.Time) string {
	today := now.Format("2006-01-02")

	parts := strings.Split(paymentDate, "/")
	if len(parts) != 2 {
		return today
	}

	month, err1 := strconv.Atoi(parts[0])
	day, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return today
	}

	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, kst)
	for _, year := range []int{now.Year(), now.Year() - 1} {
		candidate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, kst)
		// time.Date는 13월·2/30 같은 값을 다음 달로 넘겨 버리므로 되돌려 확인한다.
		if int(candidate.Month()) != month || candidate.Day() != day {
			continue
		}
		if !candidate.After(todayDate) {
			return candidate.Format("2006-01-02")
		}
	}
	return today
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