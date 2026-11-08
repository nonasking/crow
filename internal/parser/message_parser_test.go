package parser

import (
	"strings"
	"testing"

	"github.com/go-jcklk/crow/internal/constants"
)

// 실제 수신된 SMS 샘플을 픽스처로 모아둔다.
// 카드사가 포맷을 바꾸면 여기에 케이스를 추가하는 것으로 회귀를 잡는다.
const (
	// 신한 구 포맷 — 금액 뒤에 (일시불)/(할부N개월) 마커가 붙는다.
	msgShinhanOldLump = "[Web발신]\n신한카드(4557)승인 강*성 2,400원(일시불)04/20 16:56 세븐일레븐영 누적1,427,265원"
	msgShinhanOldInst = "[Web발신]\n신한카드(4557)승인 강*성 120,000원(할부3개월)06/03 20:52 이마트 누적2,000,000원"

	// 신한 신 포맷 — "신한(" 으로 시작하고 말미에 "잔액"이 붙는다.
	msgShinhanNew      = "[Web발신]\n1차 민생회복 신한(4557)승인 강*성 16,980원 08/24 17:50 땀땀 잔액 0원"
	msgShinhanNewNoBal = "[Web발신]\n신한(4557)승인 강*성 5,000원 01/02 09:10 김밥천국"
)

// 우리카드는 줄바꿈 구조에 의존한다 (장소 = 마지막에서 두 번째 줄, 최소 5줄).
var (
	msgWooriOK = strings.Join([]string{
		"[Web발신]",
		"우리카드 승인",
		"03/15 14:30",
		"12,000원",
		"스타벅스 강남점",
		"누적 1,234,567원",
	}, "\n")

	// 줄 수가 5줄 미만이면 장소를 특정할 수 없다.
	msgWooriTooFewLines = strings.Join([]string{
		"[Web발신]",
		"우리카드 승인",
		"12,000원",
		"스타벅스 강남점",
	}, "\n")

	// "N원" 패턴이 없으면 금액 추출 실패.
	msgWooriNoAmount = strings.Join([]string{
		"[Web발신]",
		"우리카드 승인",
		"03/15 14:30",
		"금액 미상",
		"스타벅스 강남점",
		"누적",
	}, "\n")
)

func TestParseWebhookAuto(t *testing.T) {
	tests := []struct {
		name        string
		msg         string
		wantAmount  int
		wantPlace   string
		wantCompany string
		wantDate    string
		wantErrMsg  string // 비어 있으면 성공을 기대한다.
	}{
		// ── 신한 ─────────────────────────────────────────────
		{
			// 알려진 한계: 구 포맷은 금액 바로 뒤에 "(일시불)04/20" 이 붙어 있어
			// 장소 문자열에 마커/날짜 접두가 남는다. 다운스트림(crow-backend의
			// normalize_merchant)에서 정규화하는 것을 전제로, 현재 동작을 고정해 둔다.
			name:        "신한 구 포맷 (일시불)",
			msg:         msgShinhanOldLump,
			wantAmount:  2400,
			wantPlace:   "(일시불)04/20 세븐일레븐영",
			wantCompany: constants.CardCompanyShinhan,
			wantDate:    "04/20",
		},
		{
			name:        "신한 구 포맷 (할부) — 누적 금액에 낚이지 않는다",
			msg:         msgShinhanOldInst,
			wantAmount:  120000,
			wantPlace:   "(할부3개월)06/03 이마트",
			wantCompany: constants.CardCompanyShinhan,
			wantDate:    "06/03",
		},
		{
			name:        "신한 신 포맷 — 잔액 표기 제거",
			msg:         msgShinhanNew,
			wantAmount:  16980,
			wantPlace:   "땀땀",
			wantCompany: constants.CardCompanyShinhan,
			wantDate:    "08/24",
		},
		{
			name:        "신한 신 포맷 — 잔액 표기 없음",
			msg:         msgShinhanNewNoBal,
			wantAmount:  5000,
			wantPlace:   "김밥천국",
			wantCompany: constants.CardCompanyShinhan,
			wantDate:    "01/02",
		},
		{
			name:        "신한 신 포맷 — 가맹점명에 '잔'이 들어가도 잘리지 않는다",
			msg:         "[Web발신]\n신한(4557)승인 강*성 9,000원 05/05 12:00 잔치국수 잔액 1,000원",
			wantAmount:  9000,
			wantPlace:   "잔치국수",
			wantCompany: constants.CardCompanyShinhan,
			wantDate:    "05/05",
		},
		{
			name:        "신한 신 포맷 — 가맹점명에 '액'이 들어가도 잘리지 않는다",
			msg:         "[Web발신]\n신한(4557)승인 강*성 15,000원 05/06 13:00 카페액자 잔액 1,000원",
			wantAmount:  15000,
			wantPlace:   "카페액자",
			wantCompany: constants.CardCompanyShinhan,
			wantDate:    "05/06",
		},
		{
			name:       "신한 신 포맷 — 가맹점명 없이 바로 잔액",
			msg:        "[Web발신]\n신한(4557)승인 강*성 5,000원 01/02 09:10 잔액 0원",
			wantErrMsg: constants.ErrInvalidCardMessageFormat,
		},
		{
			name:       "신한 승인취소는 기록 대상이 아니다",
			msg:        "[Web발신]\n신한카드(4557)승인취소 강*성 2,400원(일시불)04/20 16:56 세븐일레븐영 누적1,424,865원",
			wantErrMsg: constants.ErrCancelledPayment,
		},
		{
			name:       "신한 괄호 뒤 취소도 기록 대상이 아니다",
			msg:        "[Web발신]\n신한(4557)취소 강*성 5,000원 01/02 09:10 김밥천국 잔액 5,000원",
			wantErrMsg: constants.ErrCancelledPayment,
		},
		{
			name:        "가맹점명에 '취소'가 들어간 정상 승인은 기록한다",
			msg:         "[Web발신]\n신한(4557)승인 강*성 5,000원 01/02 09:10 취소불가상점 잔액 0원",
			wantAmount:  5000,
			wantPlace:   "취소불가상점",
			wantCompany: constants.CardCompanyShinhan,
			wantDate:    "01/02",
		},
		{
			name:       "신한 — 금액 누락",
			msg:        "[Web발신]\n신한카드(4557)승인 강*성",
			wantErrMsg: constants.ErrInvalidCardMessageFormat,
		},

		// ── 우리 ─────────────────────────────────────────────
		{
			name:        "우리 정상",
			msg:         msgWooriOK,
			wantAmount:  12000,
			wantPlace:   "스타벅스 강남점",
			wantCompany: constants.CardCompanyWoori,
			wantDate:    "03/15",
		},
		{
			name:       "우리 — 줄 수 부족 (5줄 미만)",
			msg:        msgWooriTooFewLines,
			wantErrMsg: constants.ErrInvalidCardMessageFormat,
		},
		{
			name:       "우리 — 금액 누락",
			msg:        msgWooriNoAmount,
			wantErrMsg: constants.ErrInvalidCardMessageFormat,
		},

		// ── 디스패치 실패 ────────────────────────────────────
		{
			name:       "미지원 카드사",
			msg:        "[Web발신]\nKB국민카드 결제 알림 1,000원",
			wantErrMsg: constants.ErrUnsupportedCardCompany,
		},
		{
			name:       "빈 메시지",
			msg:        "",
			wantErrMsg: constants.ErrUnsupportedCardCompany,
		},
		{
			name:       "카드사 키워드 없는 잡음 문자",
			msg:        "[Web발신]\n[국제발신] 인증번호 123456",
			wantErrMsg: constants.ErrUnsupportedCardCompany,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			amount, place, company, date, err := ParseWebhookAuto(tc.msg)

			if tc.wantErrMsg != "" {
				if err == nil {
					t.Fatalf("에러 기대했으나 nil (amount=%d, place=%q)", amount, place)
				}
				if err.Error() != tc.wantErrMsg {
					t.Errorf("에러 메시지: got %q, want %q", err.Error(), tc.wantErrMsg)
				}
				// 실패 시에는 제로값이어야 한다 (부분 파싱 결과가 새어나가면 안 된다).
				if amount != 0 || place != "" || company != "" || date != "" {
					t.Errorf("실패 시 제로값 기대, got (%d, %q, %q, %q)", amount, place, company, date)
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

func TestExtractPaymentDate(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want string
	}{
		{name: "MM/DD HH:MM 매칭", msg: "승인 2,400원 04/20 16:56 세븐일레븐", want: "04/20"},
		{name: "시간이 없으면 빈 문자열", msg: "승인 2,400원 04/20 세븐일레븐", want: ""},
		{name: "날짜 자체가 없음", msg: "승인 2,400원 세븐일레븐", want: ""},
		{name: "빈 문자열", msg: "", want: ""},
		{name: "첫 번째 MM/DD HH:MM 만 취한다", msg: "01/02 03:04 그리고 05/06 07:08", want: "01/02"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractPaymentDate(tc.msg); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseWebhookAutoLegacy(t *testing.T) {
	tests := []struct {
		name        string
		msg         string
		wantAmount  int
		wantPlace   string
		wantCompany string
		wantErr     bool
	}{
		{
			name:        "신한 신 포맷 — 결제일만 빠진 4-튜플",
			msg:         msgShinhanNew,
			wantAmount:  16980,
			wantPlace:   "땀땀",
			wantCompany: constants.CardCompanyShinhan,
		},
		{
			name:        "우리 정상",
			msg:         msgWooriOK,
			wantAmount:  12000,
			wantPlace:   "스타벅스 강남점",
			wantCompany: constants.CardCompanyWoori,
		},
		{
			name:    "미지원 카드사는 에러를 그대로 전달",
			msg:     "[Web발신]\nKB국민카드 결제 알림 1,000원",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			amount, place, company, err := ParseWebhookAutoLegacy(tc.msg)
			if tc.wantErr {
				if err == nil {
					t.Fatal("에러 기대했으나 nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("예상치 못한 에러: %v", err)
			}
			if amount != tc.wantAmount || place != tc.wantPlace || company != tc.wantCompany {
				t.Errorf("got (%d, %q, %q), want (%d, %q, %q)",
					amount, place, company, tc.wantAmount, tc.wantPlace, tc.wantCompany)
			}
		})
	}
}
