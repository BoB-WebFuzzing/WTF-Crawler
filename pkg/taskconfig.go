package pkg

import "time"

type TaskConfig struct {
	MaxCrawlCount           int    // 최대 크롤링 횟수
	FilterMode              string // simple、smart、strict
	ExtraHeaders            map[string]interface{}
	ExtraHeadersString      string
	AllDomainReturn         bool // 전체 도메인 수집
	SubDomainReturn         bool // 하위 도메인 수집
	NoHeadless              bool // headless 모드
	DomContentLoadedTimeout time.Duration
	TabRunTimeout           time.Duration     // 단일 탭 크롤링 제한 시간
	PathByFuzz              bool              // dictionary를 사용하여 경로를 찾을지 여부
	FuzzDictPath            string            // dictionary 경로
	PathFromRobots          bool              // robots.txt를 사용하여 경로를 찾을지 여부
	MaxTabsCount            int               // 열 수 있는 최대 탭 수, 즉 동시 크롤링 횟수
	ChromiumPath            string            // Chromium 경로
	ChromiumWSUrl           string            // 실행 중인 크롬 세션에 대한 WebSocket 디버깅 URL
	EventTriggerMode        string            // 이벤트 트리거 방식 : 비동기 / 동기
	EventTriggerInterval    time.Duration     // 이벤트 트리거 간격
	BeforeExitDelay         time.Duration     // 종료 전 대기시간, DOM 랜더링 대기, XHR에서 캡처 발행 대기
	EncodeURLWithCharset    bool              // 감지된 문자 집합을 사용하여 URL 자동 인코딩
	IgnoreKeywords          []string          // 무시된 키워드는 더 이상 검색되지 않으며, 검색이 완료된 후에도 요청이 전송되지 않음
	Proxy                   string            // 에이전트 요청
	CustomFormValues        map[string]string // 사용자 지정 양식 매개 변수
	CustomFormKeywordValues map[string]string // 사용자 지정 양식 keyword 매개 변수
	MaxRunTime              int64             // 최대 크롤링 시간(초), 시간 초과로 작업 종료, 원활한 종료(예시: 아직 처리되지 않은 URL은 종료할 수 없으며, 전체 작업을 종료하려면 완료해야 하는 요청이 있어야 함)
}

type TaskConfigOptFunc func(*TaskConfig)

func NewTaskConfig(optFuncs ...TaskConfigOptFunc) *TaskConfig {
	conf := &TaskConfig{}
	for _, fn := range optFuncs {
		fn(conf)
	}
	return conf
}

func WithMaxCrawlCount(maxCrawlCount int) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.MaxCrawlCount == 0 {
			tc.MaxCrawlCount = maxCrawlCount
		}
	}
}

func WithFilterMode(gen string) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.FilterMode == "" {
			tc.FilterMode = gen
		}
	}
}

func WithExtraHeaders(gen map[string]interface{}) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.ExtraHeaders == nil {
			tc.ExtraHeaders = gen
		}
	}
}

func WithExtraHeadersString(gen string) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.ExtraHeadersString == "" {
			tc.ExtraHeadersString = gen
		}
	}
}

func WithAllDomainReturn(gen bool) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if !tc.AllDomainReturn {
			tc.AllDomainReturn = gen
		}
	}
}
func WithSubDomainReturn(gen bool) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if !tc.SubDomainReturn {
			tc.SubDomainReturn = gen
		}
	}
}

func WithNoHeadless(gen bool) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if !tc.NoHeadless {
			tc.NoHeadless = gen
		}
	}
}

func WithDomContentLoadedTimeout(gen time.Duration) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.DomContentLoadedTimeout == 0 {
			tc.DomContentLoadedTimeout = gen
		}
	}
}

func WithTabRunTimeout(gen time.Duration) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.TabRunTimeout == 0 {
			tc.TabRunTimeout = gen
		}
	}
}
func WithPathByFuzz(gen bool) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if !tc.PathByFuzz {
			tc.PathByFuzz = gen
		}
	}
}
func WithFuzzDictPath(gen string) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.FuzzDictPath == "" {
			tc.FuzzDictPath = gen
		}
	}
}
func WithPathFromRobots(gen bool) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if !tc.PathFromRobots {
			tc.PathFromRobots = gen
		}
	}
}
func WithMaxTabsCount(gen int) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.MaxTabsCount == 0 {
			tc.MaxTabsCount = gen
		}
	}
}
func WithChromiumPath(gen string) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.ChromiumPath == "" {
			tc.ChromiumPath = gen
		}
	}
}
func WithEventTriggerMode(gen string) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.EventTriggerMode == "" {
			tc.EventTriggerMode = gen
		}
	}
}
func WithEventTriggerInterval(gen time.Duration) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.EventTriggerInterval == 0 {
			tc.EventTriggerInterval = gen
		}
	}
}
func WithBeforeExitDelay(gen time.Duration) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.BeforeExitDelay == 0 {
			tc.BeforeExitDelay = gen
		}
	}
}
func WithEncodeURLWithCharset(gen bool) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if !tc.EncodeURLWithCharset {
			tc.EncodeURLWithCharset = gen
		}
	}
}
func WithIgnoreKeywords(gen []string) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.IgnoreKeywords == nil || len(tc.IgnoreKeywords) == 0 {
			tc.IgnoreKeywords = gen
		}
	}
}
func WithProxy(gen string) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.Proxy == "" {
			tc.Proxy = gen
		}
	}
}
func WithCustomFormValues(gen map[string]string) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.CustomFormValues == nil || len(tc.CustomFormValues) == 0 {
			tc.CustomFormValues = gen
		}
	}
}
func WithCustomFormKeywordValues(gen map[string]string) TaskConfigOptFunc {
	return func(tc *TaskConfig) {
		if tc.CustomFormKeywordValues == nil || len(tc.CustomFormKeywordValues) == 0 {
			tc.CustomFormKeywordValues = gen
		}
	}
}
