package notion

import (
	"fmt"
	"testing"
	"time"
)

func TestFormatPaymentDate(t *testing.T) {
	currentYear := time.Now().Year()
	today := time.Now().Format("2006-01-02")

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "정상 MM/DD",
			in:   "08/24",
			want: fmt.Sprintf("%04d-08-24", currentYear),
		},
		{
			name: "0 패딩 없는 한자리 월일",
			in:   "3/5",
			want: fmt.Sprintf("%04d-03-05", currentYear),
		},
		{
			name: "빈 문자열은 오늘 날짜",
			in:   "",
			want: today,
		},
		{
			name: "포맷 깨짐 → 오늘 날짜로 폴백",
			in:   "abc",
			want: today,
		},
		{
			name: "구분자 부족 → 오늘 날짜로 폴백",
			in:   "0824",
			want: today,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatPaymentDate(tc.in)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
