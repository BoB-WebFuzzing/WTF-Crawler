package filter

import (
	"testing"

	"github.com/Qianlitp/crawlergo/pkg/config"
	"github.com/Qianlitp/crawlergo/pkg/model"

	"github.com/stretchr/testify/assert"
)

var (
	requstUrls = []string{
		"http://witcher.kro.kr/wtfadmin/index.php?route=common/dashboard&user_token=986c780cade567dd8e4e280cb3d6f388",
		"http://witcher.kro.kr/wtfadmin/index.php?route=user/profile&user_token=986c780cade567dd8e4e280cb3d6f388",
		"http://witcher.kro.kr/wtfadmin/index.php?route=common/logout&user_token=986c780cade567dd8e4e280cb3d6f388",
		"http://witcher.kro.kr/wtfadmin/index.php?route=common/dashboard&user_token=986c780cade567dd8e4e280cb3d6f388", // 중복
		"http://witcher.kro.kr/wtfadmin/index.php?route=catalog/category&user_token=986c780cade567dd8e4e280cb3d6f388",
		"http://witcher.kro.kr/wtfadmin/index.php?route=catalog/product&user_token=986c780cade567dd8e4e280cb3d6f388",
		"http://witcher.kro.kr/wtfadmin/index.php?route=catalog/subscription_plan&user_token=986c780cade567dd8e4e280cb3d6f388",
		"http://witcher.kro.kr/wtfadmin/index.php?route=catalog/filter&user_token=986c780cade567dd8e4e280cb3d6f388",
		"http://witcher.kro.kr/wtfadmin/index.php?route=catalog/attribute&user_token=986c780cade567dd8e4e280cb3d6f388",
		"http://witcher.kro.kr/wtfadmin/index.php?route=catalog/option&user_token=986c780cade567dd8e4e280cb3d6f388",
		"http://witcher.kro.kr/wtfadmin/index.php?route=catalog/attribute_group&user_token=986c780cade567dd8e4e280cb3d6f388",
		"http://witcher.kro.kr/wtfadmin/index.php?route=catalog/option&user_token=986c780cade567dd8e4e280cb3d6f388", // 중복
		"http://witcher.kro.kr/wtfadmin/index.php?route=catalog/manufacturer&user_token=986c780cade567dd8e4e280cb3d6f388",
		"http://witcher.kro.kr/wtfadmin/index.php?route=catalog/download&user_token=986c780cade567dd8e4e280cb3d6f388",
		"http://witcher.kro.kr/wtfadmin/index.php?route=catalog/filter&user_token=986c780cade123dd8e4e280cb3d6f388",
	}
)

func TestDoFilter(t *testing.T) {
	smart := NewSmartFilter(NewSimpleFilter("witcher.kro.kr"), false)

	var targets []model.Request
	for _, url := range requstUrls {
		url, err := model.GetUrl(url)
		assert.Nil(t, err)
		targets = append(targets, model.GetRequest(config.GET, url))
	}

	assert.Equal(t, smart.DoFilter(&targets[0]), false)
	assert.Equal(t, smart.DoFilter(&targets[1]), false)
	assert.Equal(t, smart.DoFilter(&targets[2]), false)
	assert.Equal(t, smart.DoFilter(&targets[3]), true)
	assert.Equal(t, smart.DoFilter(&targets[4]), false)
	assert.Equal(t, smart.DoFilter(&targets[5]), false)
	assert.Equal(t, smart.DoFilter(&targets[6]), false)
	assert.Equal(t, smart.DoFilter(&targets[7]), false)
	assert.Equal(t, smart.DoFilter(&targets[8]), false)
	assert.Equal(t, smart.DoFilter(&targets[9]), false)
	assert.Equal(t, smart.DoFilter(&targets[10]), false) // 개선 : false
	assert.Equal(t, smart.DoFilter(&targets[11]), true)  // 개선 : false
	assert.Equal(t, smart.DoFilter(&targets[12]), false)
	assert.Equal(t, smart.DoFilter(&targets[13]), false) // 개선 : false
	assert.Equal(t, smart.DoFilter(&targets[14]), true)
}
