package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Qianlitp/crawlergo/pkg"
	"github.com/Qianlitp/crawlergo/pkg/config"
	"github.com/Qianlitp/crawlergo/pkg/logger"
	model2 "github.com/Qianlitp/crawlergo/pkg/model"
	"github.com/Qianlitp/crawlergo/pkg/tools"
	"github.com/Qianlitp/crawlergo/pkg/tools/requests"
	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

/**
命令行调用适配器

用于生成开源的二进制程序
*/

type Result struct {
	// ReqList       []Request `json:"req_list"`
	// AllReqList    []Request `json:"all_req_list"`
	// AllDomainList []string  `json:"all_domain_list"`
	// SubDomainList []string  `json:"sub_domain_list"`
	RequestsFound map[string]Request `json:"requestsFound"`
	InputSet      []string           `json:"inputSet"`
}

type Request struct {
	Url     string                 `json:"url"`
	Method  string                 `json:"method"`
	Headers map[string]interface{} `json:"headers"`
	Data    string                 `json:"data"`
	Source  string                 `json:"source"`
}

type ProxyTask struct {
	req       *model2.Request
	pushProxy string
}

const (
	DefaultMaxPushProxyPoolMax = 10
	DefaultLogLevel            = "Info"
)

var (
	taskConfig              pkg.TaskConfig
	outputMode              string
	postData                string
	signalChan              chan os.Signal
	ignoreKeywords          = cli.NewStringSlice(config.DefaultIgnoreKeywords...)
	customFormTypeValues    = cli.NewStringSlice()
	customFormKeywordValues = cli.NewStringSlice()
	pushAddress             string
	pushProxyPoolMax        int
	pushProxyWG             sync.WaitGroup
	outputJsonPath          string
	logLevel                string
	Version                 string
)

func main() {
	author := cli.Author{
		Name:  "9ian1i",
		Email: "9ian1itp@gmail.com",
	}

	//ignoreKeywords = cli.NewStringSlice(config.DefaultIgnoreKeywords...)
	//customFormTypeValues = cli.NewStringSlice()
	//customFormKeywordValues = cli.NewStringSlice()

	app := &cli.App{
		Name:      "crawlergo",
		Usage:     "A powerful browser crawler for web vulnerability scanners",
		UsageText: "crawlergo [global options] url1 url2 url3 ... (must be same host)",
		Version:   Version,
		Authors:   []*cli.Author{&author},
		Flags:     cliFlags, // 프로그램 실행 시 사용할 수 있는 옵션 정의
		Action:    run,      // 프로그램이 실행될 때 호출되는 함수
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Logger.Fatal(err)
	}
}

func run(c *cli.Context) error {
	signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	if c.Args().Len() == 0 { // 인자로 url이 주어졌는지 검사
		logger.Logger.Error("url must be set")
		return errors.New("url must be set")
	}

	// 로그 출력 레벨 설정
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logger.Logger.Fatal(err)
	}
	logger.Logger.SetLevel(level)

	var targets []*model2.Request

	// 인자로 입력받은 URL에 대해 Request 객체 생성
	for _, _url := range c.Args().Slice() {
		var req model2.Request
		url, err := model2.GetUrl(_url)
		if err != nil {
			logger.Logger.Error("parse url failed, ", err)
			continue
		}
		if postData != "" {
			req = model2.GetRequest(config.POST, url, getOption())
		} else {
			req = model2.GetRequest(config.GET, url, getOption())
		}
		req.Proxy = taskConfig.Proxy
		targets = append(targets, &req)
	}
	taskConfig.IgnoreKeywords = ignoreKeywords.Value()
	if taskConfig.Proxy != "" {
		logger.Logger.Info("request with proxy: ", taskConfig.Proxy)
	}

	if len(targets) == 0 {
		logger.Logger.Fatal("no validate target.")
	}

	// 사용자 정의 양식 매개변수 처리
	taskConfig.CustomFormValues, err = parseCustomFormValues(customFormTypeValues.Value())
	if err != nil {
		logger.Logger.Fatal(err)
	}
	taskConfig.CustomFormKeywordValues, err = keywordStringToMap(customFormKeywordValues.Value())
	if err != nil {
		logger.Logger.Fatal(err)
	}

	// 크롤러 작업 시작
	task, err := pkg.NewCrawlerTask(targets, taskConfig)
	if err != nil {
		logger.Logger.Error("create crawler task failed.")
		os.Exit(-1)
	}
	if len(targets) != 0 {
		logger.Logger.Infof("Init crawler task, host: %s, max tab count: %d, max crawl count: %d, max runtime: %ds",
			targets[0].URL.Host, taskConfig.MaxTabsCount, taskConfig.MaxCrawlCount, taskConfig.MaxRunTime)
		logger.Logger.Info("filter mode: ", taskConfig.FilterMode)
	}

	// 사용자 지정 양식
	if len(taskConfig.CustomFormValues) > 0 {
		logger.Logger.Info("Custom form values, " + tools.MapStringFormat(taskConfig.CustomFormValues))
	}
	// 사용자 지정 양식 채우기
	if len(taskConfig.CustomFormKeywordValues) > 0 {
		logger.Logger.Info("Custom form keyword values, " + tools.MapStringFormat(taskConfig.CustomFormKeywordValues))
	}
	if _, ok := taskConfig.CustomFormValues["default"]; !ok {
		logger.Logger.Info("If no matches, default form input text: " + config.DefaultInputText)
		taskConfig.CustomFormValues["default"] = config.DefaultInputText
	}

	go handleExit(task)
	logger.Logger.Info("Start crawling.")
	task.Run()
	result := task.Result

	logger.Logger.Infof("Task finished, %d results, %d requests, %d subdomains, %d domains found, runtime: %d",
		len(result.ReqList), len(result.AllReqList), len(result.SubDomainList), len(result.AllDomainList), time.Now().Unix()-task.Start.Unix())

	// 内置请求代理
	if pushAddress != "" {
		logger.Logger.Info("pushing results to ", pushAddress, ", max pool number:", pushProxyPoolMax)
		Push2Proxy(result.ReqList)
	}

	// 输出结果
	outputResult(result)

	return nil
}

func getOption() model2.Options {
	var option model2.Options
	if postData != "" {
		option.PostData = postData
	}
	if taskConfig.ExtraHeadersString != "" {
		err := json.Unmarshal([]byte(taskConfig.ExtraHeadersString), &taskConfig.ExtraHeaders)
		if err != nil {
			logger.Logger.Fatal("custom headers can't be Unmarshal.")
			panic(err)
		}
		option.Headers = taskConfig.ExtraHeaders
	}
	return option
}

func parseCustomFormValues(customData []string) (map[string]string, error) {
	parsedData := map[string]string{}
	for _, item := range customData {
		keyValue := strings.Split(item, "=")
		if len(keyValue) < 2 {
			return nil, errors.New("invalid form item: " + item)
		}
		key := keyValue[0]
		if !tools.StringSliceContain(config.AllowedFormName, key) {
			return nil, errors.New("not allowed form key: " + key)
		}
		value := keyValue[1]
		parsedData[key] = value
	}
	return parsedData, nil
}

func keywordStringToMap(data []string) (map[string]string, error) {
	parsedData := map[string]string{}
	for _, item := range data {
		keyValue := strings.Split(item, "=")
		if len(keyValue) < 2 {
			return nil, errors.New("invalid keyword format: " + item)
		}
		key := keyValue[0]
		value := keyValue[1]
		parsedData[key] = value
	}
	return parsedData, nil
}

func outputResult(result *pkg.Result) {
	// 输出结果
	if outputMode == "json" {
		fmt.Println("--[Mission Complete]--")
		resBytes := getJsonSerialize(result)
		fmt.Println(string(resBytes))
	} else if outputMode == "console" {
		for _, req := range result.ReqList {
			req.FormatPrint()
		}
	}
	if len(outputJsonPath) != 0 {
		resBytes := getJsonSerialize(result)
		tools.WriteFile(outputJsonPath, resBytes)
	}
}

/*
*
原生被动代理推送支持
*/
func Push2Proxy(reqList []*model2.Request) {
	pool, _ := ants.NewPool(pushProxyPoolMax)
	defer pool.Release()
	for _, req := range reqList {
		task := ProxyTask{
			req:       req,
			pushProxy: pushAddress,
		}
		pushProxyWG.Add(1)
		go func() {
			err := pool.Submit(task.doRequest)
			if err != nil {
				logger.Logger.Error("add Push2Proxy task failed: ", err)
				pushProxyWG.Done()
			}
		}()
	}
	pushProxyWG.Wait()
}

/*
*
协程池请求的任务
*/
func (p *ProxyTask) doRequest() {
	defer pushProxyWG.Done()
	_, _ = requests.Request(p.req.Method, p.req.URL.String(), tools.ConvertHeaders(p.req.Headers), []byte(p.req.PostData),
		&requests.ReqOptions{Timeout: 1, AllowRedirect: false, Proxy: p.pushProxy})
}

func handleExit(t *pkg.CrawlerTask) {
	<-signalChan
	fmt.Println("exit ...")
	t.Pool.Tune(1)
	t.Pool.Release()
	t.Browser.Close()
	os.Exit(-1)
}

// URL에 index.php가 없으면 추가
func addIndex(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	path := parsedURL.Path
	if !strings.HasSuffix(path, ".php") {
		if !strings.HasSuffix(path, "/") {
			path += "/index.php"
		} else {
			path += "index.php"
		}
	}
	parsedURL.Path = path
	return parsedURL.String()
}

func getJsonSerialize(result *pkg.Result) []byte {
	requestsFound := make(map[string]Request)
	inputSet := make([]string, 0)

	for _, _req := range result.ReqList {

		var req Request
		var key string
		req.Method = _req.Method
		req.Url = addIndex(_req.URL.String())
		req.Source = _req.Source
		req.Data = _req.PostData
		req.Headers = _req.Headers

		// URL에서 쿼리 파싱
		parsedURL, _ := url.Parse(req.Url)
		queryMap, _ := url.ParseQuery(parsedURL.RawQuery)
		for k, v := range queryMap {
			inputSet = append(inputSet, fmt.Sprintf("%s=%s", k, v[0]))
		}

		// postData 파싱
		if req.Data != "" {
			dataPairs := strings.Split(req.Data, "&")
			for _, pair := range dataPairs {
				inputSet = append(inputSet, pair)
			}
		}

		if req.Data == "" {
			key = fmt.Sprintf("%s %s", req.Method, req.Url)
		} else {
			key = fmt.Sprintf("%s %s %s", req.Method, req.Url, req.Data)
		}
		requestsFound[key] = req
	}

	resultJSON := Result{
		RequestsFound: requestsFound,
		InputSet:      inputSet,
	}

	resBytes, err := json.MarshalIndent(resultJSON, "", "    ")
	if err != nil {
		log.Fatal("Marshal result error")
	}

	replacer := strings.NewReplacer(
		"\\u0026", "&",
		"\\u003c", "<",
		"\\u003e", ">",
		"\\\\", "\"",
	)
	resultString := replacer.Replace(string(resBytes))

	return []byte(resultString)
}
