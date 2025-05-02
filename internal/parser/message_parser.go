package parser

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-jcklk/crow/internal/constants"
)

// 공통 금액 추출 함수
func extractAmount(msg string, pattern string) (int, string, error) {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(msg)
	if len(match) < 2 {
		return 0, "", errors.New(constants.ErrInvalidCardMessageFormat)
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
		return 0, "", "", errors.New(constants.ErrInvalidCardMessageFormat)
	}
	location := strings.TrimSpace(splits[1])
	if idx := strings.Index(location, "누적"); idx != -1 {
		location = strings.TrimSpace(location[:idx])
	}

	return amount, location, constants.CardCompanyShinhan, nil
}

// 우리카드 파싱
func parseWooriCard(msg string) (int, string, string, error) {
	if !strings.Contains(msg, "출금") {
		return 0, "", "", errors.New(constants.ErrInvalidCardMessageFormat)
	}

	// 출금 금액 추출 (출금이라는 단어와 함께 있는 줄에서 추출)
	re := regexp.MustCompile(`출금\s*([\d,]+)원`)
	match := re.FindStringSubmatch(msg)
	if len(match) < 2 {
		return 0, "", "", errors.New(constants.ErrInvalidCardMessageFormat)
	}

	amountStr := strings.ReplaceAll(match[1], ",", "")
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return 0, "", "", err
	}

	// 장소는 보통 마지막 줄 (잔액 제외한 줄) 혹은 출금 줄 다음 줄
	lines := strings.Split(strings.TrimSpace(msg), "\n")
	if len(lines) < 5 {
		return 0, "", "", errors.New(constants.ErrInvalidCardMessageFormat)
	}

	location := strings.TrimSpace(lines[4])
	return amount, location, constants.CardCompanyWoori, nil
}

// 카드사 자동 감지 파싱
func ParseWebhookAuto(msg string) (int, string, string, error) {
	switch {
	case strings.Contains(msg, "[Web발신]\n신한카드"):
		return parseShinhanCard(msg)
	case strings.Contains(msg, "[Web발신]\n우리"):
		return parseWooriCard(msg)
	default:
		return 0, "", "", errors.New(constants.ErrUnsupportedCardCompany)
	}
}
