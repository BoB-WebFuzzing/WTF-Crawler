package config

import (
	"time"

	mapset "github.com/deckarep/golang-set"
)

const (
	DefaultUA               = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.0 Safari/537.36"
	MaxTabsCount            = 10
	TabRunTimeout           = 20 * time.Second
	DefaultInputText        = "Crawlergo" // 기본 텍스트 입력
	FormInputKeyword        = "Crawlergo" // 사용하지 않는 것 같음
	SuspectURLRegex         = `(?:"|')(((?:[a-zA-Z]{1,10}://|//)[^"'/]{1,}\.[a-zA-Z]{2,}[^"']{0,})|((?:/|\.\./|\./)[^"'><,;|*()(%%$^/\\\[\]][^"'><,;|()]{1,})|([a-zA-Z0-9_\-/]{1,}/[a-zA-Z0-9_\-/]{1,}\.(?:[a-zA-Z]{1,4}|action)(?:[\?|#][^"|']{0,}|))|([a-zA-Z0-9_\-/]{1,}/[a-zA-Z0-9_\-/]{3,}(?:[\?|#][^"|']{0,}|))|([a-zA-Z0-9_\-]{1,}\.(?:php|asp|aspx|jsp|json|action|html|js|txt|xml)(?:[\?|#][^"|']{0,}|)))(?:"|')`
	URLRegex                = `((https?|ftp|file):)?//[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]`
	AttrURLRegex            = ``
	DomContentLoadedTimeout = 5 * time.Second
	EventTriggerInterval    = 100 * time.Millisecond // 单位毫秒
	BeforeExitDelay         = 1 * time.Second
	DefaultEventTriggerMode = EventTriggerAsync
	MaxCrawlCount           = 200
	MaxRunTime              = 60 * 60
)

// 요청 방법
const (
	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
)

// 필터링 모드
const (
	SimpleFilterMode = "simple"
	SmartFilterMode  = "smart"
	StrictFilterMode = "strict"
)

// 이벤트 트리거 모드
const (
	EventTriggerAsync = "async"
	EventTriggerSync  = "sync"
)

// 요청의 출처
const (
	FromTarget      = "Target"     // 초기 입력 대상
	FromNavigation  = "Navigation" // 페이지 탐색 요청
	FromXHR         = "XHR"        // ajax 비동기 요청
	FromDOM         = "DOM"        // dom에 의해 파싱된 요청
	FromJSFile      = "JavaScript" // JS 스크립트에서 구문 분석
	FromFuzz        = "PathFuzz"   // 디렉토리 dictionary를 사용하여 경로를 찾음
	FromRobots      = "robots.txt" // robots.txt
	FromComment     = "Comment"    // HTML 주석
	FromWebSocket   = "WebSocket"
	FromEventSource = "EventSource"
	FromFetch       = "Fetch"
	FromHistoryAPI  = "HistoryAPI"
	FromOpenWindow  = "OpenWindow"
	FromHashChange  = "HashChange"
	FromStaticRes   = "StaticResource"
	FromStaticRegex = "StaticRegex"
)

// content-type
const (
	JSON       = "application/json"
	URLENCODED = "application/x-www-form-urlencoded"
	MULTIPART  = "multipart/form-data"
)

// 정적 파일 접미사, 이러한 접미사가 있는 모든 요청은 무시됩니다.
var (
	StaticSuffix = []string{
		"png", "gif", "jpg", "mp4", "mp3", "mng", "pct", "bmp", "jpeg", "pst", "psp", "ttf",
		"tif", "tiff", "ai", "drw", "wma", "ogg", "wav", "ra", "aac", "mid", "au", "aiff",
		"dxf", "eps", "ps", "svg", "3gp", "asf", "asx", "avi", "mov", "mpg", "qt", "rm",
		"wmv", "m4a", "bin", "xls", "xlsx", "ppt", "pptx", "doc", "docx", "odt", "ods", "odg",
		"odp", "exe", "zip", "rar", "tar", "gz", "iso", "rss", "pdf", "txt", "dll", "ico",
		"gz2", "apk", "crt", "woff", "map", "woff2", "webp", "less", "dmg", "bz2", "otf", "swf",
		"flv", "mpeg", "dat", "xsl", "csv", "cab", "exif", "wps", "m4v", "rmvb",
	}
	StaticSuffixSet mapset.Set
)

var (
	ScriptSuffix = []string{
		"php", "asp", "jsp", "asa",
	}
	ScriptSuffixSet mapset.Set
)

var DefaultIgnoreKeywords = []string{"logout", "quit", "exit"}
var AllowedFormName = []string{"default", "mail", "code", "phone", "username", "password", "qq", "id_card", "url", "date", "number"}

type ContinueResourceList []string

var InputTextMap = map[string]map[string]interface{}{
	"mail": {
		"keyword": []string{"mail"},
		"value":   "crawlergo@gmail.com",
	},
	"code": {
		"keyword": []string{"yanzhengma", "code", "ver", "captcha"},
		"value":   "123a",
	},
	"phone": {
		"keyword": []string{"phone", "number", "tel", "shouji"},
		"value":   "18812345678",
	},
	"username": {
		"keyword": []string{"name", "user", "id", "login", "account"},
		"value":   "crawlergo@gmail.com",
	},
	"password": {
		"keyword": []string{"pass", "pwd"},
		"value":   "Crawlergo6.",
	},
	"qq": {
		"keyword": []string{"qq", "wechat", "tencent", "weixin"},
		"value":   "123456789",
	},
	"IDCard": {
		"keyword": []string{"card", "shenfen"},
		"value":   "511702197409284963",
	},
	"url": {
		"keyword": []string{"url", "site", "web", "blog", "link"},
		"value":   "https://crawlergo.nice.cn/",
	},
	"date": {
		"keyword": []string{"date", "time", "year", "now"},
		"value":   "2018-01-01",
	},
	"number": {
		"keyword": []string{"day", "age", "num", "count"},
		"value":   "10",
	},
}

func init() {
	StaticSuffixSet = initSet(StaticSuffix)
	ScriptSuffixSet = initSet(ScriptSuffix)
}

func initSet(suffixs []string) mapset.Set {
	set := mapset.NewSet()
	for _, s := range suffixs {
		set.Add(s)
	}
	return set
}
