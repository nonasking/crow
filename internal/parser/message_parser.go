package parser

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrInvalidCardMessageFormat = errors.New("메시지 형식이 잘못되었습니다")
	ErrUnsupportedCardCompany   = errors.New("지원하지 않는 카드사입니다")
)

// 공통 금액 추출 함수
func extractAmount(msg string, pattern string) (int, string, error) {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(msg)
	if len(match) < 2 {
		return 0, "", ErrInvalidCardMessageFormat
	}
	amountStr := strings.ReplaceAll(match[1], ",", "")
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return 0, "", err
	}
	return amount, match[0], nil
}

// 신한카드 파싱
func parseShinhanCard(msg string) (int, string, string, error) {
	amount, matchedText, err := extractAmount(msg, `\s([\d,]+)원`)
	if err != nil {
		return 0, "", "", err
	}

	splits := strings.SplitN(msg, matchedText, 2)
	if len(splits) < 2 {
		return 0, "", "", ErrInvalidCardMessageFormat
	}
	location := strings.TrimSpace(splits[1])
	if idx := strings.Index(location, "누적"); idx != -1 {
		location = strings.TrimSpace(location[:idx])
	}

	return amount, location, "신한카드", nil
}

// 우리카드 파싱
func parseWooriCard(msg string) (int, string, string, error) {
	amount, _, err := extractAmount(msg, `(?m)^([\d,]+)원`)
	if err != nil {
		return 0, "", "", err
	}

	lines := strings.Split(strings.TrimSpace(msg), "\n")
	if len(lines) == 0 {
		return 0, "", "", ErrInvalidCardMessageFormat
	}
	location := strings.TrimSpace(lines[len(lines)-1])

	return amount, location, "우리카드", nil
}

// 카드사 자동 감지 파싱
func ParseWebhookAuto(msg string) (int, string, string, error) {
	switch {
	case strings.Contains(msg, "신한카드"):
		return parseShinhanCard(msg)
	case strings.Contains(msg, "우리카드"):
		return parseWooriCard(msg)
	default:
		return 0, "", "", ErrUnsupportedCardCompany
	}
}
