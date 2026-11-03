package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-jcklk/crow/internal/constants"
)

// fakeRecorder 는 CardRecorder 를 구현하는 테스트 더블이다.
// 호출 횟수와 인자를 기록하고, returnErr 로 저장소 실패를 주입할 수 있다.
type fakeRecorder struct {
	calls      int
	gotAmount  int
	gotPlace   string
	gotCompany string
	gotDate    string
	returnErr  error
}

func (f *fakeRecorder) CreateCardRecord(amount int, place, company, date string) error {
	f.calls++
	f.gotAmount = amount
	f.gotPlace = place
	f.gotCompany = company
	f.gotDate = date
	return f.returnErr
}

func setupRouter(recorder CardRecorder) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/webhook", NewWebhookHandler(recorder))
	return r
}

// postRaw 는 본문을 가공 없이 그대로 POST 한다 (깨진 JSON도 보낼 수 있어야 하므로).
func postRaw(r *gin.Engine, payload string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// JSON 문자열 리터럴 안에 들어갈 SMS 픽스처 (\n 은 JSON escape 로 전달된다).
const (
	shinhanNewMsg = `[Web발신]\n1차 민생회복 신한(4557)승인 강*성 16,980원 08/24 17:50 땀땀 잔액 0원`
	shinhanOldMsg = `[Web발신]\n신한카드(4557)승인 강*성 2,400원(일시불)04/20 16:56 세븐일레븐영 누적1,427,265원`
	wooriMsg      = `[Web발신]\n우리카드 승인\n03/15 14:30\n12,000원\n스타벅스 강남점\n누적 1,234,567원`
)

func TestWebhookHandler(t *testing.T) {
	tests := []struct {
		name        string
		payload     string // 요청 본문 (raw)
		recorderErr error  // recorder 가 반환할 에러 (nil 이면 성공)

		wantStatus   int
		wantCalls    int    // recorder 호출 횟수
		wantBodyHas  string // 응답 본문에 포함되어야 할 문자열
		wantAmount   int
		wantPlace    string
		wantCompany  string
		wantDate     string
		checkRecArgs bool
	}{
		// ── 성공 경로: 파서 결과가 recorder 로 그대로 전달되는가 ──
		{
			name:         "신한 신 포맷 → 200, recorder 에 파싱값 전달",
			payload:      `{"message": "` + shinhanNewMsg + `"}`,
			wantStatus:   http.StatusOK,
			wantCalls:    1,
			wantBodyHas:  `"status":"success"`,
			wantAmount:   16980,
			wantPlace:    "땀땀",
			wantCompany:  constants.CardCompanyShinhan,
			wantDate:     "08/24",
			checkRecArgs: true,
		},
		{
			name:         "신한 구 포맷 → 200",
			payload:      `{"message": "` + shinhanOldMsg + `"}`,
			wantStatus:   http.StatusOK,
			wantCalls:    1,
			wantBodyHas:  `"status":"success"`,
			wantAmount:   2400,
			wantPlace:    "(일시불)04/20 세븐일레븐영",
			wantCompany:  constants.CardCompanyShinhan,
			wantDate:     "04/20",
			checkRecArgs: true,
		},
		{
			name:         "우리 → 200",
			payload:      `{"message": "` + wooriMsg + `"}`,
			wantStatus:   http.StatusOK,
			wantCalls:    1,
			wantBodyHas:  `"status":"success"`,
			wantAmount:   12000,
			wantPlace:    "스타벅스 강남점",
			wantCompany:  constants.CardCompanyWoori,
			wantDate:     "03/15",
			checkRecArgs: true,
		},

		// ── 바인딩 실패: 400, recorder 미호출 ──
		{
			name:        "JSON 아님 → 400",
			payload:     "not a json",
			wantStatus:  http.StatusBadRequest,
			wantCalls:   0,
			wantBodyHas: constants.ErrMsgInvalidJSON,
		},
		{
			name:        "빈 본문 → 400",
			payload:     "",
			wantStatus:  http.StatusBadRequest,
			wantCalls:   0,
			wantBodyHas: constants.ErrMsgInvalidJSON,
		},
		{
			name:        "message 타입 불일치 (숫자) → 400",
			payload:     `{"message": 12345}`,
			wantStatus:  http.StatusBadRequest,
			wantCalls:   0,
			wantBodyHas: constants.ErrMsgInvalidJSON,
		},
		{
			name:        "닫히지 않은 깨진 JSON → 400",
			payload:     `{"message": "신한(1234)승인"`,
			wantStatus:  http.StatusBadRequest,
			wantCalls:   0,
			wantBodyHas: constants.ErrMsgInvalidJSON,
		},

		// ── 파싱 실패: 400, recorder 미호출 ──
		{
			name:        "message 필드 누락 → 파싱 실패 400",
			payload:     `{}`,
			wantStatus:  http.StatusBadRequest,
			wantCalls:   0,
			wantBodyHas: constants.ErrUnsupportedCardCompany,
		},
		{
			name:        "빈 message → 파싱 실패 400",
			payload:     `{"message": ""}`,
			wantStatus:  http.StatusBadRequest,
			wantCalls:   0,
			wantBodyHas: constants.ErrUnsupportedCardCompany,
		},
		{
			name:        "미지원 카드사 → 400",
			payload:     `{"message": "[Web발신]\nKB국민카드 결제 알림 1,000원"}`,
			wantStatus:  http.StatusBadRequest,
			wantCalls:   0,
			wantBodyHas: constants.ErrUnsupportedCardCompany,
		},
		{
			name:        "신한이지만 금액 누락 → 400",
			payload:     `{"message": "[Web발신]\n신한카드(4557)승인 강*성"}`,
			wantStatus:  http.StatusBadRequest,
			wantCalls:   0,
			wantBodyHas: constants.ErrInvalidCardMessageFormat,
		},

		// ── 저장소 실패: 500, recorder 는 호출된 뒤 실패 ──
		{
			name:        "Notion 기록 실패 → 500",
			payload:     `{"message": "` + shinhanNewMsg + `"}`,
			recorderErr: errors.New("notion down"),
			wantStatus:  http.StatusInternalServerError,
			wantCalls:   1,
			wantBodyHas: constants.ErrMsgNotionFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := &fakeRecorder{returnErr: tc.recorderErr}
			w := postRaw(setupRouter(rec), tc.payload)

			if w.Code != tc.wantStatus {
				t.Fatalf("status: got %d, want %d (body=%s)", w.Code, tc.wantStatus, w.Body.String())
			}
			if rec.calls != tc.wantCalls {
				t.Errorf("recorder 호출 횟수: got %d, want %d", rec.calls, tc.wantCalls)
			}
			if tc.wantBodyHas != "" && !strings.Contains(w.Body.String(), tc.wantBodyHas) {
				t.Errorf("응답 본문에 %q 가 없음: %s", tc.wantBodyHas, w.Body.String())
			}
			if tc.checkRecArgs {
				if rec.gotAmount != tc.wantAmount || rec.gotPlace != tc.wantPlace ||
					rec.gotCompany != tc.wantCompany || rec.gotDate != tc.wantDate {
					t.Errorf("recorder 인자: got (%d, %q, %q, %q), want (%d, %q, %q, %q)",
						rec.gotAmount, rec.gotPlace, rec.gotCompany, rec.gotDate,
						tc.wantAmount, tc.wantPlace, tc.wantCompany, tc.wantDate)
				}
			}
		})
	}
}

// 바인딩/파싱이 실패하는 어떤 입력에서도 recorder 가 단 한 번도 호출되지 않아야 한다.
// (Notion 쓰기는 되돌릴 수 없으므로 이 불변식을 별도로 못박는다.)
func TestWebhookHandler_NoRecordOnFailure(t *testing.T) {
	badPayloads := []string{
		"not a json",
		"",
		`{"message": 12345}`,
		`{}`,
		`{"message": ""}`,
		`{"message": "[Web발신]\nKB국민카드 결제 알림 1,000원"}`,
		`{"message": "[Web발신]\n신한카드(4557)승인 강*성"}`,
	}

	rec := &fakeRecorder{}
	r := setupRouter(rec)

	for _, p := range badPayloads {
		if w := postRaw(r, p); w.Code == http.StatusOK {
			t.Errorf("payload %q: 200이 나오면 안 됨", p)
		}
	}

	if rec.calls != 0 {
		t.Errorf("실패 페이로드에서 recorder 가 %d 회 호출됨 (0이어야 함)", rec.calls)
	}
}

func TestWebhookHandler_CancellationIgnored(t *testing.T) {
	rec := &fakeRecorder{}
	r := setupRouter(rec)

	w := postRaw(r, `{"message": "[Web발신]\n신한카드(4557)승인취소 강*성 2,400원(일시불)04/20 16:56 세븐일레븐영 누적1,424,865원"}`)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), `"ignored"`) {
		t.Errorf("body: got %s, want status=ignored", w.Body.String())
	}
	if rec.calls != 0 {
		t.Errorf("승인취소는 기록하지 않아야 한다 (calls=%d)", rec.calls)
	}
}
