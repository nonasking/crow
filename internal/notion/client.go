package notion

import (
    "fmt"
    "github.com/go-resty/resty/v2"
    "github.com/go-jcklk/crow/internal/config"
)

type NotionClient struct {
    resty  *resty.Client
    config *config.Config
}

func NewNotionClient(cfg *config.Config) *NotionClient {
    client := resty.New()
    return &NotionClient{
        resty:  client,
        config: cfg,
    }
}

func (n *NotionClient) CreatePage(title, message string) error {
    body := map[string]interface{}{
        "parent": map[string]interface{}{
            "database_id": n.config.NotionDatabaseID,
        },
        "properties": map[string]interface{}{
            "Name": map[string]interface{}{
                "title": []map[string]interface{}{
                    {
                        "text": map[string]interface{}{
                            "content": title,
                        },
                    },
                },
            },
            "Message": map[string]interface{}{
                "rich_text": []map[string]interface{}{
                    {
                        "text": map[string]interface{}{
                            "content": message,
                        },
                    },
                },
            },
        },
    }

    resp, err := n.resty.R().
        SetHeader("Authorization", fmt.Sprintf("Bearer %s", n.config.NotionToken)).
        SetHeader("Content-Type", "application/json").
        SetHeader("Notion-Version", n.config.NotionVersion).
        SetBody(body).
        Post("https://api.notion.com/v1/pages")

    if err != nil {
        return err
    }

    if resp.IsError() {
        return fmt.Errorf("Notion API error: %s", resp.String())
    }

    return nil
}
