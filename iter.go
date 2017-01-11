package jzon

import "runtime"

// A ValueError occurs when a Value method is invoked on
// a Value that does not support it. Such cases are documented
// in the description of each method.
type ValueError struct {
	Method string
	Kind   Kind
}

func (e *ValueError) Error() string {
	if e.Kind == 0 {
		return "jzon: call of " + e.Method + " on zero Value"
	}
	return "jzon: call of " + e.Method + " on " + e.Kind.String() + " Value"
}

// methodName returns the name of the calling method,
// assumed to be two stack frames above.
func methodName() string {
	pc, _, _, _ := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "unknown method"
	}
	return f.Name()
}

type ObjectIter struct {
	*JSON
	Key   string
	Value *JSON
}

type ArrayIter struct {
	*JSON
	Index int
	Value *JSON
}

func (json *JSON) mustBe(expected Kind) {
	kind := json.Predict()
	if kind != expected {
		panic(&ValueError{methodName(), kind})
	}
}

func (json *JSON) Object() *ObjectIter {
	json.mustBe(Object)
	end := json.validObjectEnd()
	if end == -1 {
		return nil
	}
	return &ObjectIter{
		&JSON{
			data:   json.data[json.offset:end],
			offset: 0,
		},
		"",
		nil,
	}
}

func (json *JSON) Array() *ArrayIter {
	json.mustBe(Array)
	end := json.validArrayEnd()
	if end == -1 {
		return nil
	}
	return &ArrayIter{
		&JSON{
			data:   json.data[json.offset+1 : end-1],
			offset: 0,
		},
		-1,
		nil,
	}
}

func (iter *ObjectIter) Reset() {
	iter.offset = 0
}

func (iter *ObjectIter) Next() bool {
	// var key string
	// var value *JSON
Loop:
	for {
		c, ok := iter.nextToken()
		if !ok {
			break Loop
		}
		switch c {
		case '{':
			iter.offset++
		case '"':
			end := iter.validStringEnd()
			iter.Key = string(iter.data[iter.offset+1 : end-1])
			iter.offset = end
		case ':':
			iter.offset++
			end, _ := iter.unsafeValueEnd()
			iter.head = iter.offset
			iter.tail = end
			iter.Value = iter.JSON
			iter.offset = end
			break Loop
		case ',':
			iter.offset++
		case '}':
			iter.offset++
			return false
			// break Loop
		}

	}
	return true
}

func (iter *ArrayIter) Next() bool {
Loop:
	for {
		c, ok := iter.nextToken()
		if !ok {
			return false
		}
		switch c {
		case ',':
			iter.offset++
		default:
			end, _ := iter.unsafeValueEnd()
			iter.Index++
			iter.head = iter.offset
			iter.tail = end
			iter.Value = iter.JSON
			iter.offset = end

			break Loop
		}

	}
	return true
}
