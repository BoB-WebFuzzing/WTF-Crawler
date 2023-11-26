package filter

import (
	"strings"

	"github.com/Qianlitp/crawlergo/pkg/config"
	"github.com/Qianlitp/crawlergo/pkg/model"
	mapset "github.com/deckarep/golang-set"
)

type SimpleFilter struct {
	UniqueSet       mapset.Set
	HostLimit       string
	staticSuffixSet mapset.Set
}

func NewSimpleFilter(host string) *SimpleFilter {
	staticSuffixSet := config.StaticSuffixSet.Clone()

	for _, suffix := range []string{"js", "css", "json"} {
		staticSuffixSet.Add(suffix)
	}
	s := &SimpleFilter{UniqueSet: mapset.NewSet(), staticSuffixSet: staticSuffixSet, HostLimit: host}
	return s
}

/*
*
필터링이 필요한 경우 true를 반환하고 필터링이 필요하지 않은 경우 false를 반환합니다.
*/
func (s *SimpleFilter) DoFilter(req *model.Request) bool {
	if s.UniqueSet == nil {
		s.UniqueSet = mapset.NewSet()
	}
	// 먼저 도메인명을 필터링할지 여부를 결정
	if s.HostLimit != "" && s.DomainFilter(req) {
		return true
	}
	// UniqueId를 바탕으로 중복 제거
	if s.UniqueFilter(req) {
		return true
	}
	// 정적 리소스 필터링
	if s.StaticFilter(req) {
		return true
	}
	return false
}

/*
*
중복 제거 요청
*/
func (s *SimpleFilter) UniqueFilter(req *model.Request) bool {
	if s.UniqueSet == nil {
		s.UniqueSet = mapset.NewSet()
	}
	if s.UniqueSet.Contains(req.UniqueId()) {
		return true
	} else {
		s.UniqueSet.Add(req.UniqueId())
		return false
	}
}

/*
*
정적 리소스 필터링
*/
func (s *SimpleFilter) StaticFilter(req *model.Request) bool {
	if s.UniqueSet == nil {
		s.UniqueSet = mapset.NewSet()
	}
	// Slice를 Map으로 변환

	if req.URL.FileExt() == "" {
		return false
	}
	if s.staticSuffixSet.Contains(req.URL.FileExt()) {
		return true
	}
	return false
}

/*
*
지정된 도메인으로 연결되는 URL만 유지
*/
func (s *SimpleFilter) DomainFilter(req *model.Request) bool {
	if s.UniqueSet == nil {
		s.UniqueSet = mapset.NewSet()
	}
	if req.URL.Host == s.HostLimit || req.URL.Hostname() == s.HostLimit {
		return false
	}
	if strings.HasSuffix(s.HostLimit, ":80") && req.URL.Port() == "" && req.URL.Scheme == "http" {
		if req.URL.Hostname()+":80" == s.HostLimit {
			return false
		}
	}
	if strings.HasSuffix(s.HostLimit, ":443") && req.URL.Port() == "" && req.URL.Scheme == "https" {
		if req.URL.Hostname()+":443" == s.HostLimit {
			return false
		}
	}
	return true
}
