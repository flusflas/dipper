package splitter

import "strings"

type Splitter struct {
	sep     string
	index   int
	remain  string
	hasMore bool
}

func NewSplitter(s, sep string) *Splitter {
	return &Splitter{sep: sep, index: -1, remain: s, hasMore: true}
}

func (s *Splitter) HasMore() bool {
	return s.hasMore
}

func (s *Splitter) Next() (string, int) {
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
