package jzon

import (
	"fmt"
	"runtime"
	"unsafe"

	"strconv"

	"github.com/davecgh/go-spew/spew"
)

// A ValueError occurs when a Value method is invoked on
// a Value that does not support it. Such cases are documented
// in the description of each method.
type ValueError struct {
	Method string
	Kind   Kind
}

func (e *ValueError) Error() string {
	if e.Kind == 0 {
		return "reflect: call of " + e.Method + " on zero Value"
	}
	return "reflect: call of " + e.Method + " on " + e.Kind.String() + " Value"
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

type JSON struct {
	data []byte
	Kind Kind
	ptr  unsafe.Pointer
}

func (j *JSON) mustBe(expected Kind) {
	if j.Kind != expected {
		panic(&ValueError{methodName(), j.Kind})
	}
}

func (j *JSON) Int64() (int64, error) {
	switch j.Kind {
	case Number:
		return strconv.ParseInt(string(j.data), 10, 64)
	}
	panic(&ValueError{"JSON.Int", j.Kind})

}

func (j *JSON) String() string {
	switch j.Kind {
	case Invalid:
		return "<invalid JSON>"
	case String:
		return string(j.data)
	case Number, Bool, Null:
		return fmt.Sprintf("<Kind: %s, Value: %s>", j.Kind, j.data)
	case Object:
		return spew.Sdump(*(*map[string]*JSON)(j.ptr))
	case Array:
		return spew.Sdump(*(*[]*JSON)(j.ptr))
	}
	// If you call String on a reflect.Value of other type, it's better to
	// print something than to panic. Useful in debugging.
	return fmt.Sprintf("<Kind: %s Value: %s>", j.Kind, j.data)
}

func (j *JSON) MapIndex(key string) (*JSON, bool) {
	j.mustBe(Object)
	jj, ok := (*(*map[string]*JSON)(j.ptr))[key]
	return jj, ok
}

func (j *JSON) Index(index int) (*JSON, error) {
	j.mustBe(Array)
	array := *(*[]*JSON)(j.ptr)
	if index < 0 || index >= len(array) {
		return nil, fmt.Errorf("array index out of range")
	}
	return array[index], nil
}

func (j *JSON) Array() []*JSON {
	j.mustBe(Array)
	return *(*[]*JSON)(j.ptr)
}

// func getPath(val interface{}, keys ...interface{}) (interface{}, error) {
// 	if len(keys) == 0 {
// 		return val, nil
// 	}
// 	switch key := keys[0].(type) {
// 	case string:
// 		nextVal, err := getFromMap(val, key)
// 		if err != nil {
// 			return nil, err
// 		}
// 		nextKeys := make([]interface{}, len(keys)-1)
// 		copy(nextKeys, keys[1:])
// 		return getPath(nextVal, nextKeys...)
// 	case int:
// 		nextVal, err := getFromArray(val, key)
// 		if err != nil {
// 			return nil, err
// 		}
// 		nextKeys := make([]interface{}, len(keys)-1)
// 		copy(nextKeys, keys[1:])
// 		return getPath(nextVal, nextKeys...)
// 	default:
// 		return nil, fmt.Errorf("%v is not string or int", keys[0])
// 	}
// 	return getPath(val, keys)
// }

func (j *JSON) Get(keys ...interface{}) (*JSON, error) {
	if len(keys) == 0 {
		return j, nil
	}

	switch key := keys[0].(type) {
	case string:
		if j.Kind != Object {
			return nil, fmt.Errorf("can not get %s from json type %s", key, j.Kind)
		}

		jj, ok := j.MapIndex(key)
		if !ok {
			return nil, fmt.Errorf("JSON does not have key %s", key)
		}

		return jj.Get(keys[1:]...)
	case int:
		if j.Kind != Array {
			return nil, fmt.Errorf("can not get %d from json type %s", key, j.Kind)
		}
		jj, err := j.Index(key)
		if err != nil {
			return nil, err
		}

		return jj.Get(keys[1:]...)
	default:
		return nil, fmt.Errorf("%v is not string or int", keys[0])
	}
}

type Kind uint

const (
	Invalid Kind = iota
	Object
	Array
	Number
	String
	Bool
	Null
)

func (k Kind) String() string {
	switch k {
	case Invalid:
		return "Invalid"
	case Object:
		return "Object"
	case Array:
		return "Array"
	case Number:
		return "Number"
	case String:
		return "String"
	case Bool:
		return "Bool"
	case Null:
		return "Null"
	default:
		return "Unkown"
	}
}
