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

// 날짜 추출 함수 (MM/DD 형식)
func extractPaymentDate(msg string) string {
    // MM/DD HH:MM 패턴 매칭
    re := regexp.MustCompile(`(\d{2}/\d{2})\s+\d{2}:\d{2}`)
    match := re.FindStringSubmatch(msg)
    if len(match) > 1 {
        return match[1]
    }
    return ""
}

// 신한카드 파싱 (기존 형식 + 새로운 형식 지원, 결제일 추가)
func parseShinhanCard(msg string) (int, string, string, string, error) {
    paymentDate := extractPaymentDate(msg)

    // 패턴 1: 기존 형식 - "신한카드(4557)승인 강*성 2,400원(일시불)04/20 16:56 세븐일레븐영 누적1,427,265원"
    if strings.Contains(msg, "신한카드") {
        amount, matchedText, err := extractAmount(msg, `\s([\d,]+)원`)
        if err != nil {
            return 0, "", "", "", err
        }

        splits := strings.SplitN(msg, matchedText, 2)
        if len(splits) < 2 {
            return 0, "", "", "", errors.New(constants.ErrInvalidCardMessageFormat)
        }
        location := strings.TrimSpace(splits[1])
        if idx := strings.Index(location, "누적"); idx != -1 {
            location = strings.TrimSpace(location[:idx])
        }
        // 시간 패턴 제거 (HH:MM 형식)
        timePattern := regexp.MustCompile(`\d{2}:\d{2}\s+`)
        location = timePattern.ReplaceAllString(location, "")
        location = strings.TrimSpace(location)

        return amount, location, constants.CardCompanyShinhan, paymentDate, nil
    }

    // 패턴 2: 새로운 형식 - "1차 민생회복 신한(4557)승인 강*성 16,980원 08/24 17:50 땀땀 잔액 0원"
    if strings.Contains(msg, "신한(") {
        amount, _, err := extractAmount(msg, `\s([\d,]+)원`)
        if err != nil {
            return 0, "", "", "", err
        }

        // 장소 추출: 시간 이후부터 "잔액" 전까지
        locationPattern := regexp.MustCompile(`\d{2}:\d{2}\s+([^잔액]+)`)
        locationMatch := locationPattern.FindStringSubmatch(msg)
        var location string
        if len(locationMatch) > 1 {
            location = strings.TrimSpace(locationMatch[1])
        }

        return amount, location, constants.CardCompanyShinhan, paymentDate, nil
    }

    return 0, "", "", "", errors.New(constants.ErrInvalidCardMessageFormat)
}

// 우리카드 파싱 (결제일 추가)
func parseWooriCard(msg string) (int, string, string, string, error) {
    paymentDate := extractPaymentDate(msg)
    // 금액 추출 (원 단위로 된 숫자 찾기)
    re := regexp.MustCompile(`([\d,]+)원`)
    match := re.FindStringSubmatch(msg)
    if len(match) < 2 {
       return 0, "", "", "", errors.New(constants.ErrInvalidCardMessageFormat)
    }
    amountStr := strings.ReplaceAll(match[1], ",", "")
    amount, err := strconv.Atoi(amountStr)
    if err != nil {
       return 0, "", "", "", err
    }
    // 장소는 마지막 줄의 전 줄
    lines := strings.Split(strings.TrimSpace(msg), "\n")
    if len(lines) < 5 {
       return 0, "", "", "", errors.New(constants.ErrInvalidCardMessageFormat)
    }
    location := strings.TrimSpace(lines[len(lines)-2])
    return amount, location, constants.CardCompanyWoori, paymentDate, nil
}

// 카드사 자동 감지 파싱 (결제일 포함)
func ParseWebhookAuto(msg string) (int, string, string, string, error) {
    switch {
    case strings.Contains(msg, "[Web발신]\n신한카드") || strings.Contains(msg, "신한("):
       return parseShinhanCard(msg)
    case strings.Contains(msg, "[Web발신]\n우리"):
       return parseWooriCard(msg)
    default:
       return 0, "", "", "", errors.New(constants.ErrUnsupportedCardCompany)
    }
}

// 기존 API 호환성을 위한 함수 (결제일 제외)
func ParseWebhookAutoLegacy(msg string) (int, string, string, error) {
    amount, location, company, _, err := ParseWebhookAuto(msg)
    return amount, location, company, err
}