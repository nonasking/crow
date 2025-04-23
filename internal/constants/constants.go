package constants

// Notion 관련 상수
const (
	NotionDefaultCategory  = "미정산"
	NotionDefaultSubCategory = "미정산"
	NotionDefaultMemo      = ""
)

// 카드사명
const (
	CardCompanyShinhan = "신한Big카드"
	CardCompanyWoori   = "우리은행통장"
)

// 에러 메시지
const (
	ErrMsgInvalidJSON   = "Invalid JSON"
	ErrMsgParseFailed   = "메시지 파싱 실패: "
	ErrMsgNotionFailed  = "Notion 업로드 실패"
	ErrInvalidCardMessageFormat = "메시지 형식이 잘못되었습니다"
	ErrUnsupportedCardCompany = "지원하지 않는 카드사입니다"
)
