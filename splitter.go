package godotted

import "strings"

type attributeSplitter struct {
	sep     string
	index   int
	remain  string
	hasMore bool
}

func newAttributeSplitter(s, sep string) *attributeSplitter {
	return &attributeSplitter{sep: sep, index: -1, remain: s, hasMore: true}
}

func (s *attributeSplitter) HasMore() bool {
	return s.hasMore
}

func (s *attributeSplitter) Next() (string, int) {
	var remain string
	if !s.hasMore {
		return "", -1
	}
	remain = s.remain
	index := strings.Index(remain, s.sep)
	if index == -1 {
		s.hasMore = false
		return s.remain, s.index + 1
	}
	s.index++
	s.remain = remain[index+1:]
	return remain[:index], s.index
}
