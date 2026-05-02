package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeRecorder struct {
	called     bool
	gotAmount  int
	gotPlace   string
	gotCompany string
	gotDate    string
	returnErr  error
}

func (f *fakeRecorder) CreateCardRecord(amount int, place, company, date string) error {
	f.called = true
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

func postJSON(r *gin.Engine, payload string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestWebhook_Success(t *testing.T) {
	rec := &fakeRecorder{}
	r := setupRouter(rec)

	msg := "[Web발신]\n1차 민생회복 신한(4557)승인 강*성 16,980원 08/24 17:50 땀땀 잔액 0원"
	body, _ := json.Marshal(map[string]string{"message": msg})
	w := postJSON(r, string(body))

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200, body=%s", w.Code, w.Body.String())
	}
	if !rec.called {
		t.Fatal("recorder.CreateCardRecord 호출되지 않음")
	}
	if rec.gotAmount != 16980 || rec.gotPlace != "땀땀" || rec.gotDate != "08/24" {
		t.Errorf("recorder args mismatch: amount=%d, place=%q, date=%q",
			rec.gotAmount, rec.gotPlace, rec.gotDate)
	}
	if !strings.Contains(w.Body.String(), `"status":"success"`) {
		t.Errorf("response body: %s", w.Body.String())
	}
}

func TestWebhook_InvalidJSON(t *testing.T) {
	rec := &fakeRecorder{}
	r := setupRouter(rec)

	w := postJSON(r, "not a json")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
	if rec.called {
		t.Error("JSON 파싱 실패 시 recorder가 호출되면 안 됨")
	}
}

func TestWebhook_ParseError(t *testing.T) {
	rec := &fakeRecorder{}
	r := setupRouter(rec)

	body, _ := json.Marshal(map[string]string{"message": "지원하지 않는 카드사 메시지"})
	w := postJSON(r, string(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
	if rec.called {
		t.Error("파싱 실패 시 recorder가 호출되면 안 됨")
	}
}

func TestWebhook_RecorderError(t *testing.T) {
	rec := &fakeRecorder{returnErr: errors.New("notion down")}
	r := setupRouter(rec)

	msg := "[Web발신]\n1차 민생회복 신한(4557)승인 강*성 16,980원 08/24 17:50 땀땀 잔액 0원"
	body, _ := json.Marshal(map[string]string{"message": msg})
	w := postJSON(r, string(body))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
	if !rec.called {
		t.Error("recorder는 호출된 뒤 에러를 반환했어야 함")
	}
}
