package jzon

func isDigit(c byte) bool {
	switch c {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}

type stack []*JSON

func (s *stack) Push(j ...*JSON) {
	*s = append(*s, j...)
}

func (s *stack) Pop() *JSON {

	if len(*s)-1 < 0 {
		return nil
	}

	j := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return j
}

func (s *stack) Peek() *JSON {
	if len(*s)-1 < 0 {
		return nil
	}
	return (*s)[len(*s)-1]
}
