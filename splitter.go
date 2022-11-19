package godotted

import "strings"

// attributeSplitter offers methods to iterate the fields of a dot-separated
// string.
type attributeSplitter struct {
	sep     byte
	index   int
	remain  string
	hasMore bool
}

// newAttributeSplitter returns a new attributeSplitter instance.
func newAttributeSplitter(s string, sep byte) *attributeSplitter {
	return &attributeSplitter{sep: sep, index: -1, remain: s, hasMore: true}
}

// HasMore returns true if the iterated string has more fields.
func (s *attributeSplitter) HasMore() bool {
	return s.hasMore
}

// Next returns the next field of the iterated string and the position of the
// field in the string (or an empty string and -1 if the string does not have
// more fields).
func (s *attributeSplitter) Next() (string, int) {
	var remain string
	if !s.hasMore {
		return "", -1
	}
	remain = s.remain
	index := strings.IndexByte(remain, s.sep)
	if index == -1 {
		s.hasMore = false
		return s.remain, s.index + 1
	}
	s.index++
	s.remain = remain[index+1:]
	return remain[:index], s.index
}
