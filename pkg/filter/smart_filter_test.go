package filter

import (
	"testing"

	"github.com/Qianlitp/crawlergo/pkg/config"
	"github.com/Qianlitp/crawlergo/pkg/model"

	"github.com/stretchr/testify/assert"
)

var (
	requstUrls = []string{
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=a&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=b&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=c&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=d&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=e&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=f&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=g&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=h&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=i&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=j&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=k&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=l&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=m&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=n&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=o&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=p&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=q&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=r&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=s&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=t&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=u&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=v&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=w&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=x&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=y&wr_id=6",
		"https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=z&wr_id=6",
		// "https://demo.sir.kr/gnuboard5/",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=free",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=qa",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=notice",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=gallery",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=job",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=market",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=humor",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=free&wr_id=1",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=free&wr_id=2",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=free&wr_id=3",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=free&wr_id=4",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=free&wr_id=5",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=free&wr_id=6",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=free&wr_id=7",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=free&wr_id=8",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=free&wr_id=9",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=free&wr_id=10",
		// "https://demo.sir.kr/gnuboard5/bbs/search.php?sfl=wr_subject&sop=and&stx=test",
		// "https://demo.sir.kr/gnuboard5/bbs/search.php?sfl=wr_content&sop=and&stx=test",
		// "https://demo.sir.kr/gnuboard5/bbs/search.php?sfl=wr_subject||wr_content&sop=and&stx=test",
		// "https://demo.sir.kr/gnuboard5/bbs/search.php?sfl=wr_subject||wr_content&sop=or&stx=test",
		// "https://demo.sir.kr/gnuboard5/bbs/search.php?sfl=wr_subject||wr_content&sop=and&stx=%27%20or%20%271%27=%271",
		// "https://demo.sir.kr/gnuboard5/bbs/search.php?sfl=wr_subject||wr_content&sop=and&stx=%27%20or%20%271%27=%271",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=free&wr_id=4456",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=free&wr_id=4457",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=free&wr_id=4458",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=notice&wr_id=1164",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=notice&wr_id=1165",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=notice&wr_id=1166",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=notice&wr_id=1167",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=gallery&wr_id=678",
		// "https://demo.sir.kr/gnuboard5/bbs/board.php?bo_table=gallery&wr_id=679",
	}

	smart = NewSmartFilter(NewSimpleFilter(""), true)
)

func TestDoFilter(t *testing.T) {
	targets := []model.Request{}
	for _, url := range requstUrls {
		url, err := model.GetUrl(url)
		assert.Nil(t, err)
		targets = append(targets, model.GetRequest(config.GET, url))
	}

	for _, target := range targets[:7] {
		assert.Equal(t, smart.DoFilter(&target), false)
	}
	assert.Equal(t, smart.DoFilter(&targets[8]), true)
	assert.Equal(t, smart.DoFilter(&targets[9]), true)
	assert.Equal(t, smart.DoFilter(&targets[10]), true)
}
