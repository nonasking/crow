package notion

import (
	"fmt"
	"os"
	"time"
	"log"

	"github.com/go-resty/resty/v2"
)

type NotionClient struct {
	client      *resty.Client
	DatabaseID  string
	AccessToken string
}

func NewClient() *NotionClient {
	return &NotionClient{
		client:      resty.New(),
		DatabaseID:  os.Getenv("NOTION_DATABASE_ID"),
		AccessToken: os.Getenv("NOTION_TOKEN"),
	}
}

func (n *NotionClient) CreateCardRecord(
	amount int,
	place string,
	cardCompany string,
) error {
	payload := map[string]interface{}{
		"parent": map[string]interface{}{
			"database_id": n.DatabaseID,
		},
		"properties": map[string]interface{}{
			"항목": map[string]interface{}{
				"title": []map[string]interface{}{
					{
						"text": map[string]interface{}{
							"content": place,
						},
					},
				},
			},
			"소분류": map[string]interface{}{
				"select": map[string]interface{}{
					"name": "미정산",
				},
			},
			"결제방식": map[string]interface{}{
				"select": map[string]interface{}{
					"name": fmt.Sprintf("%s카드", cardCompany),
				},
			},
			"날짜": map[string]interface{}{
				"date": map[string]string{
					"start": time.Now().Format("2006-01-02"),
				},
			},
			"수입": map[string]interface{}{
				"number": 0,
			},
			"지출": map[string]interface{}{
				"number": amount,
			},
			"대분류": map[string]interface{}{
				"select": map[string]interface{}{
					"name": "기타",
				},
			},
			"비고": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"text": map[string]interface{}{
							"content": "비고",
						},
					},
				},
			},
		},
	}

	resp, err := n.client.R().
		SetHeader("Authorization", "Bearer "+n.AccessToken).
		SetHeader("Notion-Version", "2022-06-28").
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post("https://api.notion.com/v1/pages")

	if err != nil || resp.StatusCode() >= 400 {
		return fmt.Errorf("Notion API 에러: %v", err)
	}

	return nil
}

