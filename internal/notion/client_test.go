package notion

import (
	"testing"
	"time"
)

func TestFormatPaymentDateAt(t *testing.T) {
	// 2026-03-01 10:00 KST
	now := time.Date(2026, 3, 1, 10, 0, 0, 0, kst)

	tests := []struct {
		name string
		in   string
		now  time.Time
		want string
	}{
		{name: "정상 MM/DD", in: "02/24", now: now, want: "2026-02-24"},
		{name: "오늘 날짜는 올해", in: "03/01", now: now, want: "2026-03-01"},
		{name: "0 패딩 없는 한자리 월일", in: "2/5", now: now, want: "2026-02-05"},
		{name: "미래 날짜는 작년 결제로 본다", in: "12/31", now: now, want: "2025-12-31"},
		{
			name: "UTC로는 아직 12/31이어도 KST 새해 새벽이면 올해 1/1",
			in:   "01/01",
			// 2026-12-31 15:30 UTC == 2027-01-01 00:30 KST
			now:  time.Date(2026, 12, 31, 15, 30, 0, 0, time.UTC).In(kst),
			want: "2027-01-01",
		},
		{name: "올해·작년 모두 2/29가 없으면 오늘", in: "02/29", now: now, want: "2026-03-01"},
		{name: "존재하지 않는 날짜(13월)는 오늘", in: "13/45", now: now, want: "2026-03-01"},
		{name: "빈 문자열은 오늘 날짜", in: "", now: now, want: "2026-03-01"},
		{name: "포맷 깨짐 → 오늘 날짜로 폴백", in: "abc", now: now, want: "2026-03-01"},
		{name: "구분자 부족 → 오늘 날짜로 폴백", in: "0824", now: now, want: "2026-03-01"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatPaymentDateAt(tc.in, tc.now)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatPaymentDateUsesKST(t *testing.T) {
	want := time.Now().In(kst).Format("2006-01-02")
	if got := formatPaymentDate(""); got != want {
		t.Errorf("빈 날짜 폴백은 KST 오늘이어야 한다: got %q, want %q", got, want)
	}
}
