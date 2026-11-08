package parser

import (
    "errors"
    "regexp"
    "strconv"
    "strings"

    "github.com/go-jcklk/crow/internal/constants"
)

// ErrCancelledPayment는 승인취소 문자를 받았을 때 반환된다.
// 취소 문자도 승인과 같은 형식이라 그대로 파싱하면 환불이 지출로 한 번 더 기록된다.
var ErrCancelledPayment = errors.New(constants.ErrCancelledPayment)

// 승인취소 판별: "승인취소" 또는 카드번호 괄호 바로 뒤의 "취소".
// 문자 전체에서 "취소"를 찾으면 가맹점명("취소불가상점")이나 적요("예약취소수수료")에
// 걸려 정상 지출이 기록되지 않는다.
var cancellationPattern = regexp.MustCompile(`승인\s*취소|\(\d+\)\s*취소`)

// 신한 신 포맷 가맹점: "HH:MM" 뒤부터 "잔액" 앞(또는 끝)까지.
// [^잔액] 문자 클래스로 쓰면 '잔'·'액' 글자 하나하나를 제외해서
// "잔치국수"가 빈 문자열, "카페액자"가 "카페"로 잘린다.
var shinhanNewLocationPattern = regexp.MustCompile(`\d{2}:\d{2}\s+(.+?)\s*(?:잔액|$)`)

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
        locationMatch := shinhanNewLocationPattern.FindStringSubmatch(strings.TrimSpace(msg))
        var location string
        if len(locationMatch) > 1 {
            location = strings.TrimSpace(locationMatch[1])
        }
        // 가맹점명 없이 바로 "잔액"이 오면 정규식이 "잔액…"을 통째로 잡는다.
        if location == "" || strings.HasPrefix(location, "잔액") {
            return 0, "", "", "", errors.New(constants.ErrInvalidCardMessageFormat)
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
    if cancellationPattern.MatchString(msg) {
       return 0, "", "", "", ErrCancelledPayment
    }
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