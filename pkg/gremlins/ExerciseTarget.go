package gremlins

import (
	"context"
	"github.com/Qianlitp/crawlergo/pkg"
	"github.com/Qianlitp/crawlergo/pkg/config"
	"github.com/Qianlitp/crawlergo/pkg/engine"
	"github.com/Qianlitp/crawlergo/pkg/filter"
	"github.com/Qianlitp/crawlergo/pkg/logger"
	"github.com/Qianlitp/crawlergo/pkg/model"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"log"
	"strings"
)

type Browser struct {
	Ctx          context.Context
	Cancel       context.CancelFunc
	ExtraHeaders map[string]interface{}
}

type GremlinTest struct {
	Browser *Browser
	Result  []*model.Request // Crawlergo에서 수집한 URL
	filter  filter.FilterHandler
	Config  *pkg.TaskConfig
}

func InitBrowser(chromiumPath string, extraHeaders map[string]interface{}, noHeadless bool) *Browser {
	var bro Browser
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", !noHeadless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-images", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-xss-auditor", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("allow-running-insecure-content", true),
		chromedp.Flag("disable-webgl", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.WindowSize(1920, 1080),
		chromedp.ExecPath(chromiumPath),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// 새로운 브라우저 생성
	bctx, _ := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	if err := chromedp.Run(bctx); err != nil {
		logger.Logger.Fatal("[GremlinTest] Chromedp 실행 오류 : ", err.Error())
	}
	bro.Ctx = bctx
	bro.Cancel = cancel
	bro.ExtraHeaders = extraHeaders
	return &bro
}

func (bro *Browser) Close() {
	logger.Logger.Info("[GremlinTest] Browser를 종료합니다.")
	if err := browser.Close().Do(bro.Ctx); err != nil {
		logger.Logger.Debug(err)
	}
	bro.Cancel()
}

func GremlinTestCode(taskConf pkg.TaskConfig, crawlergoResult []*model.Request) (*GremlinTest, error) {
	gremlinTest := GremlinTest{
		Config: &taskConf,
	}

	baseFilter := filter.NewSimpleFilter(crawlergoResult[0].URL.Host)

	gremlinTest.filter = filter.NewSmartFilter(baseFilter, false)
	gremlinTest.Browser = InitBrowser(taskConf.ChromiumPath, taskConf.ExtraHeaders, taskConf.NoHeadless)

	for _, req := range crawlergoResult {
		if gremlinTest.filter.DoFilter(req) {
			continue
		}
		if req.Method != "GET" {
			req.GremlinTesting = true
		}
		gremlinTest.Result = append(gremlinTest.Result, req)
	}

	return &gremlinTest, nil
}

func (gt *GremlinTest) Run() ([]*model.Request, error) {
	defer gt.Browser.Close()

	var tempGremlinURL []*model.Request          //	Gremlin에서 수집한 모든 URL (필터링 전)
	var beforeSmartFilteringURL []*model.Request // Gremlin에서 수집한 URL 중 필터링 전 URL
	var collectedURL int

	addGremlinScript := `
		script = document.createElement('script');
		script.type = 'text/javascript';
		script.async = true;
		script.src = 'https://unpkg.com/gremlins.js';
		document.getElementsByTagName('head')[0].appendChild(script);
	
		var signal = document.createElement('div');
		signal.id = 'gremlin-script-added';
		document.body.appendChild(signal);
		console.log("[WTFuzz] Gremlin Script Added")
		`

	addFormScript := `
		let formEntries = {}
		const repeativeHorde = () => {
			let all_submitable =  [...document.getElementsByTagName("form"),
                ...document.querySelectorAll('[type="submit"]')];

            let randomArr = all_submitable;

            for(let i = 0; i < all_submitable.length; i++) {
                let submitable_item = randomArr[i];
                if(typeof submitable_item.submit === 'function') {
                    submitable_item.submit();
                } else if(typeof submitable_item.requestSubmit === 'function') {
                    try{
                        submitable_item.requestSubmit();
                    } catch (e){
                        console.log(e.stack)
                    }
                }
                if(typeof submitable_item.click === 'function') {
                    submitable_item.click()
                }
            }
		}
		const triggerSimulatedOnChange = (element, newValue, prototype) => {
            const lastValue = element.value;
            element.value = newValue;

            const nativeInputValueSetter = Object.getOwnPropertyDescriptor(prototype, 'value').set;
            nativeInputValueSetter.call(element, newValue);
            const event = new Event('input', { bubbles: true });

            // React 15
            event.simulated = true;
            // React >= 16
            let tracker = element._valueTracker;
            if (tracker) {
                tracker.setValue(lastValue);
            }
            element.dispatchEvent(event);
        };
        const fillTextAreaElement = (element) => {
            let rnd =  Math.random();
            let value = "2";
            if (rnd > 0.7){
                value = "Witcher";
            } else if (rnd > 0.3) {
                value =  "127.0.0.1";
            }
            triggerSimulatedOnChange(element, value, window.HTMLTextAreaElement.prototype);

            return value;
        };
        const fillNumberElement = (element) => {
            const number = randomizer.character({ pool: '0123456789' });
            const newValue = element.value + number;
            triggerSimulatedOnChange(element, newValue, window.HTMLInputElement.prototype);

            return number;
        };
        const fillSelect = (element) => {
            const options = element.querySelectorAll('option');
            if (options.length === 0) return;
            const randomOption = randomizer.pick(options);
            options.forEach((option) => {
                option.selected = option.value === randomOption.value;
            });

            let event = new Event('change');
            element.dispatchEvent(event);

            return randomOption.value;
        };
        const fillRadio = (element) => {
            // using mouse events to trigger listeners
            const evt = document.createEvent('MouseEvents');
            evt.initMouseEvent('click', true, true, window, 0, 0, 0, 0, 0, false, false, false, false, 0, null);
            element.dispatchEvent(evt);

            return element.value;
        };
        const fillCheckbox = (element) => {
            // using mouse events to trigger listeners
            const evt = document.createEvent('MouseEvents');
            evt.initMouseEvent('click', true, true, window, 0, 0, 0, 0, 0, false, false, false, false, 0, null);
            element.dispatchEvent(evt);

            return element.value;
        };
        const fillEmail = (element) => {
            const email = "test@test.com";
            triggerSimulatedOnChange(element, email, window.HTMLInputElement.prototype);

            return email;
        };
        const fillTextElement = (element) => {
            if (!element){
                console.log('[*] fillTextElement : Element is null')
                return 0;
            }
            let oldDateYearFirst = "1998-10-11";
            let oldDateMonthFirst = "11-12-1997";

            let rnd =  Math.random()
            let current_value = element.value;
            let desc = element.id;
            if (!desc){
                desc = element.name;
            }
            // let's leave it the default value for a little while.
            if (current_value && current_value > "" && desc > ""){
                if (desc in formEntries){
                    if (formEntries[desc]["inc"] < 5){
                        formEntries[desc]["inc"] += 1;
                        return current_value;
                    }
                } else {
                    formEntries[desc] = {origingal_value: current_value, inc:1};
                    return current_value;
                }
            }

            let value = "2";

            if (rnd > .2 && element.placeholder && (element.placeholder.match(/[Yy]{4}.[Mm]{2}.[Dd]{2}/) || element.placeholder.match(/[Mm]{2}.[Dd]{2}.[Yy]{4}/))){
                let yearfirst = element.placeholder.match(/[Yy]{4}(.)[Mm]{2}.[Dd]{2}/);
                let sep = "-";
                if (yearfirst)
                    sep = yearfirst[1]
                else {
                    let monthfirst = element.placeholder.match(/[Mm]{2}(.)[Dd]{2}.[Yy]{4}/)
                    if (monthfirst){
                        sep = monthfirst[1];
                    } else {
                        console.log("[WC] this should never occur, couldn't find the separator, defaulting to -")
                    }
                }

                if (element.placeholder.match(/[Yy]{4}.[Mm]{2}.[Dd]{2}/)) {
                    value = rnd > .8 ? currentDateYearFirst.replace("-",sep) : oldDateYearFirst.replace("-",sep);
                } else if (element.placeholder.match(/[Mm]{2}.[Dd]{2}.[Yy]{4}/)){
                    value = rnd > .8 ? currentDateMonthFirst.replace("-",sep) : oldDateMonthFirst.replace("-",sep);
                }
            } else if (rnd > .5 && element.name && (element.name.search(/dob/i) !== -1 || element.name.search(/birth/i) !== -1 )){
                value = rnd > .75 ? oldDateMonthFirst : oldDateYearFirst;
            } else if (rnd > .5 && element.name && (element.name.search(/date/i) !== -1)){
                value = rnd > .75 ? currentDateMonthFirst : currentDateYearFirst;
            } else if (rnd > .5 && element.name && (element.name.search(/time/i) !== -1)){
                value = element.name.search(/start/i) !== -1 ? "8:01" : "11:11";
            } else if (rnd > 0.4) {
                value = "127.0.0.1";
            } else if (rnd > .3){
                value = "WTFCrawler";
            } else if (rnd > 0.2) {
                value = value = rnd > .35 ? currentDateYearFirst : oldDateYearFirst;
            } else if (rnd > 0.1) {
                value = rnd > .45 ? currentDateYearFirst : oldDateYearFirst;
            } else if (rnd > 0.0){
                value = current_value;
            }
            element.value = value;
            if (Math.random() > 0.80){
                repeativeHorde();
            }
            return value;
        };
        const fillPassword = (element) => {
            let rnd =  Math.random()
            if (rnd < 0.8) {
                element.value = "password";
            } else {
                element.value = "passwor";
            }
            return element.value;
        };
        const clickSub = (element) => {
            element.click();
            return element.value
        }

		let wFormElementMapTypes = {
            textarea: fillTextAreaElement,
            'input[type="text"]': fillTextElement,
            'input[type="password"]': fillPassword,
            'input[type="number"]': fillNumberElement,
            select: fillSelect,
            'input[type="radio"]': fillRadio,
            'input[type="checkbox"]': fillCheckbox,
            'input[type="email"]': fillEmail,
            'input[type="submit"]' : clickSub,
            'button' : clickSub,
            'input:not([type])': fillTextElement,
        }

		let randomizer = new gremlins.Chance();

		var signal = document.createElement('div');
		signal.id = 'gremlin-form-script-added';
		document.body.appendChild(signal);

		console.log("[WTFuzz] Gremlin Form Script Added")
		`

	gremlinSettings := `
		let noChance = new gremlins.Chance();
		noChance.character = function(options) {
				if (options != null){
					return "2";
				} else {
					let rnd =  Math.random()
					if (rnd > 0.7){
						return "WTFCrawlergo";
					} else if (rnd > 0.3){
						return "127.0.0.1";
					} else {
						return "2"
					}
				}
			};

		let ff = window.gremlins.species.formFiller({elementMapTypes:wFormElementMapTypes, randomizer:noChance});
		const distributionStrategy = gremlins.strategies.distribution({
			distribution: [0.80, 0.15, 0.05],
			delay: 20,
		});

		var signal = document.createElement('div');
		signal.id = 'gremlin-setting-script-added';
		document.body.appendChild(signal);

		console.log("[WTFuzz] Gremlin Setting Script Added")
		`

	coolHorde := `
		async function runGremlin() {
			for (let i = 0; i < 2; i++) {
				console.log("Gremlin TEST START : " + (i+1)/5);
	
				console.log("Gremlin TESTING : formFiller()");
				await gremlins.createHorde({
					species: [ff],
					mogwais: [gremlins.mogwais.alert(),gremlins.mogwais.gizmo()],
					strategies: [gremlins.strategies.allTogether({ nb: 1000 })],
					randomizer: noChance
				}).unleash();
	
				console.log("Gremlin TESTING : clicker(), formFiller(), scroller()");
				await gremlins.createHorde({
					species: [gremlins.species.clicker(), ff, gremlins.species.scroller()],
					mogwais: [gremlins.mogwais.alert(),gremlins.mogwais.gizmo()],
					strategies: [distributionStrategy],
					randomizer: noChance
				}).unleash();
	
				console.log("Gremlin TESTING : clicker(), typer()");
				await gremlins.createHorde({
					species: [gremlins.species.clicker(), gremlins.species.typer()],
					mogwais: [gremlins.mogwais.alert(),gremlins.mogwais.gizmo()],
					strategies: [gremlins.strategies.allTogether({ nb: 1000 })],
					randomizer: noChance
				}).unleash();
			}
			console.log("[WTFuzz] Gremlin TEST Complete");
			var signal = document.createElement('div');
			signal.id = 'gremlin-complete';
			document.body.appendChild(signal);
		}
		runGremlin()
	`

	// copy gt.Result to tempGremlinURL
	for _, req := range gt.Result {
		tempGremlinURL = append(tempGremlinURL, req)
	}

	collectedURL = len(tempGremlinURL)

	for {
		if collectedURL == 0 {
			break
		}
		for _, req := range gt.Result {
			if req.GremlinTesting {
				continue
			}
			if req.URL.Scheme == "https" {
				continue
			}
			collectedURL = 0
			targetURL := req.URL.String()
			logger.Logger.Info("Gremlin Test URL : ", targetURL)

			ctx, cancel := chromedp.NewContext(gt.Browser.Ctx)
			defer cancel()

			initPageLoading := true

			chromedp.ListenTarget(ctx, func(ev interface{}) {
				switch ev := ev.(type) {
				case *runtime.EventConsoleAPICalled:
					for _, arg := range ev.Args {
						if strings.Contains(string([]byte(arg.Value)), "WTFuzz") {
							logger.Logger.Info("[console] " + string([]byte(arg.Value)))
						}
						if strings.Contains(string([]byte(arg.Value)), "Gremlin Script Added") {
							initPageLoading = false
						}
					}
				case *network.EventResponseReceived:
					c := chromedp.FromContext(ctx)
					ctx := cdp.WithExecutor(ctx, c.Target)
					res, err := network.GetResponseBody(ev.RequestID).Do(ctx)
					if err != nil {
						logger.Logger.Debug("[GremlinTest] Response Parsing ERROR : ", err)
						return
					}
					resStr := string(res)

					if ev.Response.Status >= 400 {
						logger.Logger.Info("[GremlinTest] Response Status : ", ev.Response.Status, ", Skipped ", ev.Response.URL)
						cancel()
						return
					}
					switch ev.Response.Headers["Content-Type"] {
					case "application/javascript", "application/json", "text/javascript", "text/json":
						logger.Logger.Info("[GremlinTest] Response Status : ", ev.Response.Status, ", Skipped ", ev.Response.URL)
						cancel()
						return
					}

					if len(resStr) < 20 {
						logger.Logger.Info("[GremlinTest] Response Text is Too Short. Skipped ", ev.Response.URL)
						cancel()
						return
					}

					if !strings.Contains(resStr, "<body") || !strings.Contains(resStr, "<form") || !strings.Contains(resStr, "<frameset") {
						logger.Logger.Info("[GremlinTest] Response Text is not HTML. Skipped ", ev.Response.URL)
						cancel()
						return
					}
				case *fetch.EventRequestPaused:
					go func() {
						c := chromedp.FromContext(ctx)
						ctx := cdp.WithExecutor(ctx, c.Target)
						logger.Logger.Debug("[GremlinTest] Intercepted Request : ", ev.Request.URL)

						requestURL, err := model.GetUrl(ev.Request.URL)
						if err != nil {
							logger.Logger.Error("[GremlinTest] URL Parse Error : ", err)
							return
						}

						_option := model.Options{
							Headers:  ev.Request.Headers,
							PostData: ev.Request.PostData,
						}
						req := model.GetRequest(ev.Request.Method, requestURL, _option)

						var simpleFilter filter.SimpleFilter
						if !simpleFilter.UniqueFilter(&req) {
							req.Source = config.FromFetch
							//gt.GremlinResult = append(gt.GremlinResult, &req)
							beforeSmartFilteringURL = append(beforeSmartFilteringURL, &req)
						}

						if engine.IsIgnoredByKeywordMatch(req, gt.Config.IgnoreKeywords) {
							_ = fetch.FailRequest(ev.RequestID, network.ErrorReasonAborted).Do(ctx)
							return
						}

						if strings.HasSuffix(requestURL.Path, ".css") || strings.HasSuffix(requestURL.Path, ".js") {
							_ = fetch.ContinueRequest(ev.RequestID).Do(ctx)
							return
						}

						if ev.Request.Method == "POST" {
							_ = fetch.ContinueRequest(ev.RequestID).Do(ctx)
							return
						}

						if ev.Request.URL == targetURL {
							if initPageLoading {
								_ = fetch.ContinueRequest(ev.RequestID).Do(ctx)
							} else {
								_ = fetch.FailRequest(ev.RequestID, network.ErrorReasonAborted).Do(ctx)
							}
							return
						} else {
							tempURL, err := model.GetUrl(targetURL)
							if err != nil {
								logger.Logger.Error("[GremlinTest] URL Parse Error : ", err)
								return
							}
							tempURL.Scheme = "https"
							if ev.Request.URL == tempURL.String() {
								if initPageLoading {
									_ = fetch.ContinueRequest(ev.RequestID).Do(ctx)
								} else {
									_ = fetch.FailRequest(ev.RequestID, network.ErrorReasonAborted).Do(ctx)
								}
								return
							}
							if strings.Contains(requestURL.String(), "gremlins") {
								_ = fetch.ContinueRequest(ev.RequestID).Do(ctx)
								return
							} else {
								_ = fetch.FailRequest(ev.RequestID, network.ErrorReasonAborted).Do(ctx)
								return
							}
						}
					}()
				}
			})

			if err := chromedp.Run(ctx,
				fetch.Enable(),
				network.SetExtraHTTPHeaders(gt.Browser.ExtraHeaders),
				chromedp.Navigate(targetURL),
				chromedp.WaitVisible(`body`, chromedp.ByQuery),
				chromedp.Evaluate(addGremlinScript, nil),
				chromedp.WaitVisible(`#gremlin-script-added`, chromedp.ByQuery),
				chromedp.Evaluate(addFormScript, nil),
				chromedp.WaitVisible(`#gremlin-form-script-added`, chromedp.ByQuery),
				chromedp.Evaluate(gremlinSettings, nil),
				chromedp.WaitVisible(`#gremlin-setting-script-added`, chromedp.ByQuery),
				chromedp.Evaluate(coolHorde, nil),
				chromedp.WaitVisible(`#gremlin-complete`, chromedp.ByQuery),
			); err != nil {
				logger.Logger.Error("[GremlinTest] Error : ", err)
				continue
			}

			req.GremlinTesting = true

			for _, req := range beforeSmartFilteringURL {
				if gt.filter.DoFilter(req) {
					logger.Logger.Debugf("[GremlinTest] filter req: " + req.URL.RequestURI())
					continue
				}
				if req.Method != "GET" {
					req.GremlinTesting = true
				}
				tempGremlinURL = append(tempGremlinURL, req)
				logger.Logger.Info("[GremlinTest] Collected NEW URL : " + req.Method + " " + req.URL.String() + " " + req.PostData)
				collectedURL += 1
			}
		}
		for _, req := range tempGremlinURL {
			gt.Result = append(gt.Result, req)
		}
	}
	logger.Logger.Info("[GremlinTest] Test Complete")
	return gt.Result, nil
}
