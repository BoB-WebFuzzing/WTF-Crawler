package pkg

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/Qianlitp/crawlergo/pkg/config"
	engine2 "github.com/Qianlitp/crawlergo/pkg/engine"
	"github.com/Qianlitp/crawlergo/pkg/filter"
	filter2 "github.com/Qianlitp/crawlergo/pkg/filter"
	"github.com/Qianlitp/crawlergo/pkg/logger"
	"github.com/Qianlitp/crawlergo/pkg/model"

	"github.com/panjf2000/ants/v2"
)

type CrawlerTask struct {
	Browser       *engine2.Browser
	RootDomain    string               // 하위 도메인 수집을 위한 현재 크롤링 루트 도메인
	Targets       []*model.Request     // 크롤링 대상
	Result        *Result              // 크롤링 결과
	Config        *TaskConfig          // 설정 정보
	filter        filter.FilterHandler // 필터 개체
	Pool          *ants.Pool           // 크롤링 작업 풀 (고루틴 풀 동시성 제어)
	taskWG        sync.WaitGroup       // 협업 풀의 모든 작업이 완료될 때까지 대기
	crawledCount  int                  // 크롤링 작업 수
	taskCountLock sync.Mutex           // 총 작업 수를 제어하는 데 사용되는 뮤텍스
	Start         time.Time            // 크롤링 시작 시간

}

type Result struct {
	ReqList       []*model.Request // 동일한 도메인에 대한 Request
	AllReqList    []*model.Request // 모든 도메인에 대한 Request
	AllDomainList []string         // 모든 도메인 목록
	SubDomainList []string         // 하위 도메인 목록
	resultLock    sync.Mutex       // 결과를 제어하는 데 사용되는 뮤텍스
}

type tabTask struct {
	crawlerTask *CrawlerTask
	browser     *engine2.Browser
	req         *model.Request
}

// 새로운 크롤러 작업 생성
func NewCrawlerTask(targets []*model.Request, taskConf TaskConfig) (*CrawlerTask, error) {
	crawlerTask := CrawlerTask{
		Result: &Result{},
		Config: &taskConf,
	}

	baseFilter := filter.NewSimpleFilter(targets[0].URL.Host)

	if taskConf.FilterMode == config.SmartFilterMode {
		crawlerTask.filter = filter.NewSmartFilter(baseFilter, false)

	} else if taskConf.FilterMode == config.StrictFilterMode {
		crawlerTask.filter = filter.NewSmartFilter(baseFilter, true)

	} else {
		crawlerTask.filter = baseFilter
	}

	if len(targets) == 1 {
		_newReq := *targets[0]
		newReq := &_newReq
		_newURL := *_newReq.URL
		newReq.URL = &_newURL
		if targets[0].URL.Scheme == "http" {
			newReq.URL.Scheme = "https"
		} else {
			newReq.URL.Scheme = "http"
		}
		targets = append(targets, newReq)
	}
	crawlerTask.Targets = targets[:]

	for _, req := range targets {
		req.Source = config.FromTarget
	}

	// 비즈니스 코드와 데이터 코드를 분리하고 일부 기본 구성을 초기화합니다.
	// function option과 프록시를 사용하여 taskConf 구성을 초기화합니다.
	for _, fn := range []TaskConfigOptFunc{
		WithTabRunTimeout(config.TabRunTimeout),
		WithMaxTabsCount(config.MaxTabsCount),
		WithMaxCrawlCount(config.MaxCrawlCount),
		WithDomContentLoadedTimeout(config.DomContentLoadedTimeout),
		WithEventTriggerInterval(config.EventTriggerInterval),
		WithBeforeExitDelay(config.BeforeExitDelay),
		WithEventTriggerMode(config.DefaultEventTriggerMode),
		WithIgnoreKeywords(config.DefaultIgnoreKeywords),
	} {
		fn(&taskConf)
	}

	// 사용자 정의 헤더가 Unmarshal 되는지 확인
	if taskConf.ExtraHeadersString != "" {
		err := json.Unmarshal([]byte(taskConf.ExtraHeadersString), &taskConf.ExtraHeaders)
		if err != nil {
			logger.Logger.Error("custom headers can't be Unmarshal.")
			return nil, err
		}
	}

	if len(taskConf.ChromiumWSUrl) > 0 {
		crawlerTask.Browser = engine2.ConnectBrowser(taskConf.ChromiumWSUrl, taskConf.ExtraHeaders)
	} else {
		crawlerTask.Browser = engine2.InitBrowser(taskConf.ChromiumPath, taskConf.ExtraHeaders, taskConf.Proxy, taskConf.NoHeadless)
	}
	crawlerTask.RootDomain = targets[0].URL.RootDomain()

	// 동시 풀 생성
	p, _ := ants.NewPool(taskConf.MaxTabsCount)
	crawlerTask.Pool = p

	return &crawlerTask, nil
}

/*
*
요청 목록을 기반으로 tabTask 공동 프로그래밍 작업 목록 생성
*/
func (t *CrawlerTask) generateTabTask(req *model.Request) *tabTask {
	task := tabTask{
		crawlerTask: t,
		browser:     t.Browser,
		req:         req,
	}
	return &task
}

// 크롤링 현재 작업 시작
func (t *CrawlerTask) Run() {
	defer t.Pool.Release()  // 동시 풀 해제
	defer t.Browser.Close() // 브라우저 종료

	t.Start = time.Now()
	if t.Config.PathFromRobots {
		reqsFromRobots := GetPathsFromRobots(*t.Targets[0])
		logger.Logger.Info("get paths from robots.txt: ", len(reqsFromRobots))
		t.Targets = append(t.Targets, reqsFromRobots...)
	}

	if t.Config.FuzzDictPath != "" {
		if t.Config.PathByFuzz {
			logger.Logger.Warn("`--fuzz-path` is ignored, using `--fuzz-path-dict` instead")
		}
		reqsByFuzz := GetPathsByFuzzDict(*t.Targets[0], t.Config.FuzzDictPath)
		t.Targets = append(t.Targets, reqsByFuzz...)
	} else if t.Config.PathByFuzz {
		reqsByFuzz := GetPathsByFuzz(*t.Targets[0])
		logger.Logger.Info("get paths by fuzzing: ", len(reqsByFuzz))
		t.Targets = append(t.Targets, reqsByFuzz...)
	}

	t.Result.AllReqList = t.Targets[:]

	var initTasks []*model.Request
	for _, req := range t.Targets {
		if t.filter.DoFilter(req) {
			logger.Logger.Debugf("filter req: " + req.URL.RequestURI())
			continue
		}
		initTasks = append(initTasks, req)
		t.Result.ReqList = append(t.Result.ReqList, req)
	}
	logger.Logger.Info("filter repeat, target count: ", len(initTasks))

	for _, req := range initTasks {
		if !engine2.IsIgnoredByKeywordMatch(*req, t.Config.IgnoreKeywords) {
			t.addTask2Pool(req)
		}
	}

	t.taskWG.Wait()

	// 对全部请求进行唯一去重
	todoFilterAll := make([]*model.Request, len(t.Result.AllReqList))
	copy(todoFilterAll, t.Result.AllReqList)

	t.Result.AllReqList = []*model.Request{}
	var simpleFilter filter2.SimpleFilter
	for _, req := range todoFilterAll {
		if !simpleFilter.UniqueFilter(req) {
			t.Result.AllReqList = append(t.Result.AllReqList, req)
		}
	}

	// 全部域名
	t.Result.AllDomainList = AllDomainCollect(t.Result.AllReqList)
	// 子域名
	t.Result.SubDomainList = SubDomainCollect(t.Result.AllReqList, t.RootDomain)
}

/*
*
동시 풀에 작업 추가
추가 전 실시간 필터링
*/
func (t *CrawlerTask) addTask2Pool(req *model.Request) {
	t.taskCountLock.Lock()
	if t.crawledCount >= t.Config.MaxCrawlCount {
		t.taskCountLock.Unlock()
		return
	} else {
		t.crawledCount += 1
	}

	if t.Start.Add(time.Second * time.Duration(t.Config.MaxRunTime)).Before(time.Now()) {
		t.taskCountLock.Unlock()
		return
	}
	t.taskCountLock.Unlock()

	t.taskWG.Add(1)
	task := t.generateTabTask(req)
	go func() {
		err := t.Pool.Submit(task.Task)
		if err != nil {
			t.taskWG.Done()
			logger.Logger.Error("addTask2Pool ", err)
		}
	}()
}

/*
*
单个运行的tab标签任务，实现了workpool的接口
*/
func (t *tabTask) Task() {
	defer t.crawlerTask.taskWG.Done()

	// 设置tab超时时间，若设置了程序最大运行时间， tab超时时间和程序剩余时间取小
	timeremaining := t.crawlerTask.Start.Add(time.Duration(t.crawlerTask.Config.MaxRunTime) * time.Second).Sub(time.Now())
	tabTime := t.crawlerTask.Config.TabRunTimeout
	if t.crawlerTask.Config.TabRunTimeout > timeremaining {
		tabTime = timeremaining
	}

	if tabTime <= 0 {
		return
	}

	tab := engine2.NewTab(t.browser, *t.req, engine2.TabConfig{
		TabRunTimeout:           tabTime,
		DomContentLoadedTimeout: t.crawlerTask.Config.DomContentLoadedTimeout,
		EventTriggerMode:        t.crawlerTask.Config.EventTriggerMode,
		EventTriggerInterval:    t.crawlerTask.Config.EventTriggerInterval,
		BeforeExitDelay:         t.crawlerTask.Config.BeforeExitDelay,
		EncodeURLWithCharset:    t.crawlerTask.Config.EncodeURLWithCharset,
		IgnoreKeywords:          t.crawlerTask.Config.IgnoreKeywords,
		CustomFormValues:        t.crawlerTask.Config.CustomFormValues,
		CustomFormKeywordValues: t.crawlerTask.Config.CustomFormKeywordValues,
	})
	tab.Start()

	// 收集结果
	t.crawlerTask.Result.resultLock.Lock()
	t.crawlerTask.Result.AllReqList = append(t.crawlerTask.Result.AllReqList, tab.ResultList...)
	t.crawlerTask.Result.resultLock.Unlock()

	for _, req := range tab.ResultList {
		if !t.crawlerTask.filter.DoFilter(req) {
			t.crawlerTask.Result.resultLock.Lock()
			t.crawlerTask.Result.ReqList = append(t.crawlerTask.Result.ReqList, req)
			t.crawlerTask.Result.resultLock.Unlock()
			if !engine2.IsIgnoredByKeywordMatch(*req, t.crawlerTask.Config.IgnoreKeywords) {
				t.crawlerTask.addTask2Pool(req)
			}
		}
	}
}
