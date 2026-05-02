package parser

import (
	"strings"
	"testing"

	"github.com/go-jcklk/crow/internal/constants"
)

func TestParseWebhookAuto(t *testing.T) {
	tests := []struct {
		name        string
		msg         string
		wantAmount  int
		wantPlace   string
		wantCompany string
		wantDate    string
		wantErr     bool
	}{
		{
			name:        "신한카드 신규 포맷",
			msg:         "[Web발신]\n1차 민생회복 신한(4557)승인 강*성 16,980원 08/24 17:50 땀땀 잔액 0원",
			wantAmount:  16980,
			wantPlace:   "땀땀",
			wantCompany: constants.CardCompanyShinhan,
			wantDate:    "08/24",
		},
		{
			name: "우리은행 체크카드",
			msg: strings.Join([]string{
				"[Web발신]",
				"우리카드 승인",
				"03/15 14:30",
				"12,000원",
				"스타벅스 강남점",
				"누적 1,234,567원",
			}, "\n"),
			wantAmount:  12000,
			wantPlace:   "스타벅스 강남점",
			wantCompany: constants.CardCompanyWoori,
			wantDate:    "03/15",
		},
		{
			name:    "지원하지 않는 카드사",
			msg:     "[Web발신]\nKB국민카드 결제 알림 1,000원",
			wantErr: true,
		},
		{
			name:    "신한 포맷 일부 누락 (금액 없음)",
			msg:     "[Web발신]\n신한카드(4557)승인 강*성",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			amount, place, company, date, err := ParseWebhookAuto(tc.msg)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("에러 기대했으나 nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("예상치 못한 에러: %v", err)
			}
			if amount != tc.wantAmount {
				t.Errorf("amount: got %d, want %d", amount, tc.wantAmount)
			}
			if place != tc.wantPlace {
				t.Errorf("place: got %q, want %q", place, tc.wantPlace)
			}
			if company != tc.wantCompany {
				t.Errorf("company: got %q, want %q", company, tc.wantCompany)
			}
			if date != tc.wantDate {
				t.Errorf("date: got %q, want %q", date, tc.wantDate)
			}
		})
	}
}

func TestParseWebhookAutoLegacy(t *testing.T) {
	msg := "[Web발신]\n1차 민생회복 신한(4557)승인 강*성 16,980원 08/24 17:50 땀땀 잔액 0원"
	amount, place, company, err := ParseWebhookAutoLegacy(msg)
	if err != nil {
		t.Fatalf("예상치 못한 에러: %v", err)
	}
	if amount != 16980 || place != "땀땀" || company != constants.CardCompanyShinhan {
		t.Errorf("got (%d, %q, %q), want (16980, 땀땀, %q)", amount, place, company, constants.CardCompanyShinhan)
	}
}
