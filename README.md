# crow

A lightweight Go webhook server that automatically parses Korean card transaction messages and logs them to a Notion database.

## Overview

Crow receives SMS/web notification messages from Korean banks (Shinhan Card, Woori Bank) via a webhook, extracts transaction details using regex-based parsing, and creates expense records in a Notion database — eliminating manual expense tracking.

```
Card notification → POST /webhook → Parse → Notion DB record
```

## Features

- Auto-detects card type from raw notification text
- Supports multiple Shinhan Card message formats
- Supports Woori Bank notification format
- Extracts amount, merchant, payment method, and date
- Creates structured records in Notion with mapped properties

## Requirements

- Go 1.24+
- A [Notion integration](https://www.notion.so/my-integrations) with write access to your database
- A Notion database with the properties described below

## Installation

```bash
git clone https://github.com/go-jcklk/crow.git
cd crow
go mod download
```

## Configuration

Create a `.env` file in the project root:

```env
NOTION_TOKEN=secret_xxxxxxxxxxxxxxxxxxxx
NOTION_DATABASE_ID=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
NOTION_VERSION=2022-06-28
PORT=8080  # optional, defaults to 8080
```

| Variable             | Description                          | Required |
|----------------------|--------------------------------------|----------|
| `NOTION_TOKEN`       | Notion integration secret token      | Yes      |
| `NOTION_DATABASE_ID` | Target Notion database ID            | Yes      |
| `NOTION_VERSION`     | Notion API version                   | Yes      |
| `PORT`               | Server port (default: `8080`)        | No       |

## Running

```bash
# Development
go run ./cmd/server/main.go

# Build and run
go build -o crow ./cmd/server
./crow
```

## API

### `POST /webhook`

Receives a raw card notification message and creates a Notion record.

**Request**

```json
{
  "message": "<raw card notification text>"
}
```

**Response — success**

```json
{ "status": "success" }
```

**Response — error**

```json
{ "error": "<error description>" }
```

| Status | Meaning                                      |
|--------|----------------------------------------------|
| 200    | Record created successfully                  |
| 400    | Invalid JSON or unsupported message format   |
| 500    | Notion API error                             |

## Supported Message Formats

### Shinhan Card (Format 1)

```
[Web발신]
신한카드(4557)승인 강*성 2,400원(일시불)04/20 16:56 세븐일레븐영 누적1,427,265원
```

### Shinhan Card (Format 2)

```
1차 민생회복 신한(4557)승인 강*성 16,980원 08/24 17:50 땀땀 잔액 0원
```

### Woori Bank

```
[Web발신]
우리은행통장에서
출금되었습니다
12,000원
04/21 14:30
스타벅스
```

## Notion Database Schema

The following properties must exist in your Notion database:

| Property   | Type   | Description                       |
|------------|--------|-----------------------------------|
| `항목`     | Title  | Merchant name                     |
| `소분류`   | Select | Sub-category (default: `미정산`)  |
| `결제방식` | Select | Payment method (card company)     |
| `날짜`     | Date   | Transaction date (`YYYY-MM-DD`)   |
| `수입`     | Number | Income amount (default: `0`)      |
| `지출`     | Number | Expense amount                    |
| `대분류`   | Select | Main category (default: `미정산`) |
| `비고`     | Text   | Notes (default: empty)            |

## Project Structure

```
crow/
├── cmd/server/
│   └── main.go               # Entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Environment variable loading
│   ├── constants/
│   │   └── constants.go      # Card names, error messages
│   ├── handler/
│   │   └── webhook.go        # HTTP handler
│   ├── notion/
│   │   └── client.go         # Notion API client
│   └── parser/
│       └── message_parser.go # Card message parsing
├── go.mod
└── go.sum
```

## License

MIT
