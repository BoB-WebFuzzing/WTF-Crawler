package gremlins

import (
	"github.com/Qianlitp/crawlergo/pkg"
	"github.com/Qianlitp/crawlergo/pkg/logger"
	"github.com/Qianlitp/crawlergo/pkg/model"
	"testing"
)

var (
	taskConfig pkg.TaskConfig
	ReqList    []*model.Request
)

func TestExerciseTarget(t *testing.T) {
	logger.Logger.Info("[TEST CODE] Start Gremlins Test")

	url, _ := model.GetUrl("http://127.0.0.1:5000")
	req := model.GetRequest("GET", url)
	ReqList = append(ReqList, &req)

	gremlinTask, err := GremlinTestCode(taskConfig, ReqList)
	if err != nil {
		logger.Logger.Error("create gremlin task failed.")
		t.Fail()
	}
	_, _ = gremlinTask.Run()
	ReqList = gremlinTask.Result
}
