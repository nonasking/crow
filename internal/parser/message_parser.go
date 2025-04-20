package parser

import (
    "errors"
	"regexp"
	"strconv"
	"strings"
)


func ParseShinhanCardWebhook(msg string) (int, string, string, error) {
	amountRegex := regexp.MustCompile(`\s([\d,]+)원`)
	amountMatch := amountRegex.FindStringSubmatch(msg)
	if len(amountMatch) < 2 {
		return 0, "", "", ErrInvalidFormat
	}
	amountStr := strings.ReplaceAll(amountMatch[1], ",", "")
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return 0, "", "", err
	}

	afterAmount := strings.Split(msg, amountMatch[0])
	if len(afterAmount) < 2 {
		return 0, "", "", ErrInvalidFormat
	}

	location := strings.TrimSpace(afterAmount[1])
	if strings.Contains(location, "누적") {
		location = strings.Split(location, "누적")[0]
	}

	return amount, strings.TrimSpace(location), "신한카드", nil
}


func ParseWooriCardWebhook(msg string) (int, string, string, error) {
	amountRegex := regexp.MustCompile(`(?m)^([\d,]+)원`)
	amountMatch := amountRegex.FindStringSubmatch(msg)
	if len(amountMatch) < 2 {
		return 0, "", "", ErrInvalidFormat
	}
	amountStr := strings.ReplaceAll(amountMatch[1], ",", "")
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return 0, "", "", err
	}

	lines := strings.Split(msg, "\n")
	lastLine := strings.TrimSpace(lines[len(lines)-1])

	return amount, lastLine, "우리카드", nil
}


// 커스텀 에러
var ErrInvalidFormat = &ParseError{"메시지 형식이 잘못되었습니다"}

type ParseError struct {
	Msg string
}

func (e *ParseError) Error() string {
	return e.Msg
}


func ParseWebhookAuto(msg string) (int, string, string, error) {
	switch {
	case strings.Contains(msg, "신한카드"):
		return ParseShinhanCardWebhook(msg)
	case strings.Contains(msg, "우리카드"):
		return ParseWooriCardWebhook(msg)
	default:
		return 0, "", "", errors.New("지원하지 않는 카드사입니다")
	}
}

