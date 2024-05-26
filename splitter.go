package dipper

// attributeSplitter offers methods to iterate the substrings of a string
// using a given separator.
type attributeSplitter struct {
	s            string
	sep          string
	index        int
	hasMore      bool
	scanIndex    int
	prevBrackets bool
}

// newAttributeSplitter returns a new attributeSplitter instance.
func newAttributeSplitter(s, sep string) *attributeSplitter {
	return &attributeSplitter{sep: sep, index: -1, s: s, hasMore: true}
}

// HasMore returns true if the iterated string has more fields.
func (s *attributeSplitter) HasMore() bool {
	return s.hasMore
}

// Next returns the next field of the iterated string and the position of the
// field in the string (or an empty string and -1 if the string does not have
// more fields).
func (s *attributeSplitter) Next() (string, int) {
	if !s.hasMore {
		return "", -1
	}

	remain := s.s[s.scanIndex:]
	index := -1
	enclosureCount := 0
	t := len(remain) - len(s.sep) + 1
	separatorLength := len(s.sep)

	for i := 0; i < t; i++ {
		if remain[i] == '[' {
			if !s.prevBrackets {
				index = i
				separatorLength = 0
				s.prevBrackets = true
				break
			}
			enclosureCount++
		} else if remain[i] == ']' {
			enclosureCount--
			s.prevBrackets = false
		}

		if enclosureCount == 0 && remain[i:i+separatorLength] == s.sep {
			index = i
			break
		}
	}

	if index == -1 {
		s.hasMore = false
		return remain, s.index + 1
	}
	s.index++
	s.scanIndex += index + separatorLength
	return remain[:index], s.index
}

// CountRemaining returns the number of remaining fields in the string.
func (s *attributeSplitter) CountRemaining() int {
	remain := s.s[s.scanIndex:]
	count := 0
	enclosureCount := 0
	t := len(remain) - len(s.sep) + 1
	separatorLength := len(s.sep)

	for i := 0; i < t; i++ {
		if remain[i] == '[' {
			enclosureCount++
			count++
			continue
		} else if remain[i] == ']' {
			enclosureCount--
		}

		if enclosureCount == 0 && remain[i:i+separatorLength] == s.sep {
			count++
		}
	}

	return count + 1
}
