package jzon

import "fmt"

// A SyntaxError occurs when parsing json syntax
type SyntaxError struct {
	kind   Kind
	offset int
	data   []byte
}

func (e SyntaxError) Error() string {
	start := e.offset - 5
	if start < 0 {
		start = 0
	}
	end := e.offset + 5
	if end > len(e.data) {
		end = len(e.data)
	}

	return fmt.Sprintf("Json syntax error when parsing kind(%s), context near: |%s|", e.kind, string(e.data[start:end]))
}

// A KindError occurs when a JSON method is invoked on
// a JSON kind that does not support it. Such cases are documented
// in the description of each method.
type KindError struct {
	Method string
	Kind   Kind
}

func (e *KindError) Error() string {
	if e.Kind == 0 {
		return "jzon: call of " + e.Method + " on invalid JSON"
	}
	return "jzon: call of " + e.Method + " on " + e.Kind.String() + " JSON"
}