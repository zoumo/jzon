package jzon

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

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

const (
	// for number
	flagIsNegative   flag = 1
	flagIsFloat      flag = 1 << 1
	flagIsScientific flag = 1 << 2

	// for object and array
	flagNeedKey   flag = 1
	flagNeedColon flag = 1 << 1
	flagNeedComma flag = 1 << 2
	flagNeedValue flag = 1 << 3
	flagNeedStart flag = 1 << 4
	flagNeedEnd   flag = 1 << 5
)

var (
	trueBytes  = []byte("true")
	falseBytes = []byte("false")
	nullBytes  = []byte("null")
)

type JSON struct {
	data   []byte
	offset int
	head   int
	tail   int
	err    error
}

func FromBytes(data []byte) *JSON {
	json := &JSON{
		data:   data,
		offset: 0,
		head:   0,
		tail:   len(data),
	}
	return json
}

func FromReader(r io.Reader) (*JSON, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	json := &JSON{
		data:   buf,
		offset: 0,
		head:   0,
		tail:   len(buf),
	}
	return json, nil
}

func (json *JSON) Reuse(data []byte) *JSON {
	json.data = data
	json.offset = 0
	json.head = 0
	json.tail = len(data)
	json.nextToken()
	return json
}

// nextToken returns the byte read at index i, move offset to i
// if a valid byte is found
func (json *JSON) nextToken() (byte, bool) {
	if json.offset >= len(json.data) {
		return 0, false
	}
	for i, c := range json.data[json.offset:] {
		switch c {
		case ' ', '\n', '\t', '\r':
			continue
		}
		json.offset += i
		return json.data[json.offset], true
	}
	json.offset = len(json.data)
	return 0, false
}

func (json *JSON) Predict() Kind {
	c, ok := json.nextToken()
	if !ok {
		return Invalid
	}
	switch c {
	case '{':
		return Object
	case '[':
		return Array
	case '"':
		return String
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return Number
	case 't', 'f':
		return Bool
	case 'n':
		return Null
	default:
		return Invalid
	}
}

func (json *JSON) unsafeValueEnd() (int, Kind) {
	switch kind := json.Predict(); kind {
	case Object:
		return json.unsafeObjectEnd(), kind
	case Array:
		return json.unsafeArrayEnd(), kind
	case Number:
		return json.validNumberEnd(), kind
	case String:
		return json.validStringEnd(), kind
	case Bool, Null:
		return json.validLiteralValueEnd(), kind
	default:
		json.err = SchemaError{Invalid, json.offset, json.data}
		return -1, Invalid
	}
}

func (json *JSON) validValueEnd() (int, Kind) {
	switch kind := json.Predict(); kind {
	case Object:
		return json.validObjectEnd(), kind
	case Array:
		return json.validArrayEnd(), kind
	case Number:
		return json.validNumberEnd(), kind
	case String:
		return json.validStringEnd(), kind
	case Bool, Null:
		return json.validLiteralValueEnd(), kind
	default:
		json.err = SchemaError{Invalid, json.offset, json.data}
		return -1, Invalid
	}
}

func (json *JSON) validStringEnd() int {

	validHexDigit := func(data []byte) bool {
		if len(data) < 4 {
			return false
		}
		for i := 0; i < 4; i++ {
			c := data[i]
			if ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F') {
				return true
			}
		}
		return false
	}

	validEnd := func(data []byte) int {
		// https://tools.ietf.org/html/rfc7159#section-7
		n := len(data)
		// escaped := false
		for i := 0; i < n; {
			switch data[i] {
			case '\\':
				// look one more byte
				if i+1 >= n {
					return -i - 1
				}
				// escaped = true
				switch data[i+1] {
				case 'u':
					// maybe \uXXXX
					// need 4 hexdemical digit
					if !validHexDigit(data[i+2:]) {
						// error string format
						return -i - 1
					}
					i += 6
				case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
					// \", \\, \/, \b, \f, \n, \r, \t
					i += 2
				default:
					// error
					return -i - 1
				}

			case '"':
				// match end
				return i + 1
			default:
				i++
			}
		}
		// can not find string end
		return -n - 1
	}

	// if json.data[json.offset] != '"' {
	// 	return -1
	// }

	end := validEnd(json.data[json.offset+1:])
	if end < 0 {
		json.err = SchemaError{String, json.offset + 1 - (end + 1), json.data}
		return -1
	}

	return json.offset + 1 + end
}

func (json *JSON) validLiteralValueEnd() int {
	// https://github.com/golang/go/commit/69cd91a5981c49eaaa59b33196bdb5586c18d289
	n := len(json.data)

	validTokenEnd := func(index int) bool {
		if index == n {
			return true
		}
		switch json.data[index] {
		case ' ', '\n', '\r', '\t', ',', '}', ']':
			return true
		}
		return false
	}

	var kind Kind
	switch json.data[json.offset] {
	case 't':
		end := json.offset + 4
		if end <= n &&
			bytes.Equal(json.data[json.offset:end], trueBytes) &&
			validTokenEnd(end) {
			return end
		}
		kind = Bool
	case 'f':
		end := json.offset + 5
		if end <= n &&
			bytes.Equal(json.data[json.offset:end], falseBytes) &&
			validTokenEnd(end) {
			return end
		}
		kind = Bool
	case 'n':
		end := json.offset + 4
		if end <= n &&
			bytes.Equal(json.data[json.offset:end], nullBytes) &&
			validTokenEnd(end) {
			return end
		}
		kind = Null
	default:
		kind = Invalid
	}
	json.err = SchemaError{kind, json.offset, json.data}

	return -1
}

func (json *JSON) validNumberEnd() int {

	validEnd := func(data []byte) (int, *flag) {
		n := len(data)
		flag := new(flag)
		if n == 0 {
			return -1, flag
		}

		lookOneMore := func(j int) (index int, digit bool, ok bool) {
			next := j + 1
			if next >= n {
				return j, false, false
			}

			return next, isDigit(data[next]), true
		}

		i := 0
		// first of all
		if data[i] == '-' {
			flag.add(flagIsNegative)
			i++
		}
		switch data[i] {
		case '0':
			// if first digit is 0, we should make sure that next byte should not be digit
			_, isdigit, more := lookOneMore(i)
			if !more {
				// only 0
				return i + 1, flag
			}
			if isdigit {
				// invalid format: number should not start with 0, like 01
				return -i - 1, flag
			}
			// start with 0, next byte is not digit
			i++
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			i++
		default:
			return -i - 1, flag
		}

		for i < n {
			switch data[i] {
			case ' ', '\n', '\r', '\t', ',', '}', ']':
				// here is the valid end
				return i, flag
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				// digit can be self-cycle
				i++
			case 'e', 'E':
				// https://en.wikipedia.org/wiki/Scientific_notation
				if flag.contains(flagIsScientific) {
					// 123ee
					return -i - 1, flag
				}
				flag.add(flagIsScientific)

				// +1
				j, isdigit, more := lookOneMore(i)
				if !more {
					// invalid format: 1.23e
					return -i - 1, flag
				}

				if data[j] == '-' || data[j] == '+' {
					// +2
					_, isdigit2, _ := lookOneMore(j)
					if isdigit2 {
						// 1.23e+3
						i += 3 // +3
						continue
					}
					// invalid format: 1.23e-a or 1.23e-
					return -i - 3, flag
				} else if !isdigit {
					// invalid format: 1.23ea
					return -i - 2, flag
				}
				// 1.23e3
				i += 2
			case '.':
				if flag.contains(flagIsFloat) || flag.contains(flagIsScientific) {
					// 1. get more than one dot
					// 2. or already get one e or E, only digit is allowed after e in scientific notation
					return -i - 1, flag
				}
				flag.add(flagIsFloat)
				_, isdigit, more := lookOneMore(i)
				if !more || !isdigit {
					// 1223.
					// 123.a
					return -i - 1, flag
				}

				i += 2
			default:
				return -i - 1, flag
			}
		}
		// meet string end
		return i, flag
	}
	end, _ := validEnd(json.data[json.offset:])

	if end < 0 {
		json.err = SchemaError{Number, json.offset - end, json.data}
		return -1
	}
	return json.offset + end
}

func (json *JSON) validArrayEnd() int {
	validEnd := func() int {
		flag := new(flag).add(flagNeedStart)
		for {
			c, ok := json.nextToken()
			if !ok {
				return -json.offset - 1
			}
			switch c {

			case ',':
				if !flag.contains(flagNeedComma) {
					return -json.offset - 1
				}
				flag.remove(flagNeedComma, flagNeedEnd).add(flagNeedValue)
				json.offset++
			case ']':
				if !flag.contains(flagNeedEnd) {
					return -json.offset - 1
				}
				json.offset++
				return json.offset
			case '[':
				if flag.contains(flagNeedValue) {
					// [[1,2], [3,4]]
					goto DefaultCase
				}

				if !flag.contains(flagNeedStart) {
					return -json.offset - 1
				}
				flag.remove(flagNeedStart).add(flagNeedValue, flagNeedEnd)
				json.offset++
				break
			DefaultCase: // label
				fallthrough
			default:
				if !flag.contains(flagNeedValue) {
					return -json.offset - 1
				}
				end, _ := json.validValueEnd()
				if end == -1 {
					return -1
				}
				flag.remove(flagNeedValue).add(flagNeedComma, flagNeedEnd)
				json.offset = end
			}
		}
	}

	now := json.offset
	end := validEnd()
	if end < 0 {
		if json.err == nil {
			json.err = SchemaError{Array, -(end + 1), json.data}
		}
		end = -1
	}

	json.offset = now
	return end
}

func (json *JSON) validObjectEnd() int {
	validEnd := func() int {
		flag := new(flag).add(flagNeedStart)
		for {
			c, ok := json.nextToken()
			if !ok {
				return -json.offset - 1
			}
			switch c {
			case '{':
				if !flag.contains(flagNeedStart) {
					return -json.offset - 1
				}
				flag.remove(flagNeedStart).add(flagNeedKey, flagNeedEnd)
				json.offset++
			case '"':
				if !flag.contains(flagNeedKey) {
					return -json.offset - 1
				}
				end := json.validStringEnd()
				if end == -1 {
					return -1
				}
				flag.remove(flagNeedKey, flagNeedEnd).add(flagNeedColon) // clean
				json.offset = end                                        // move to end
			case ':':
				if !flag.contains(flagNeedColon) {
					return -json.offset - 1
				}
				json.offset++
				end, _ := json.validValueEnd()
				if end == -1 {
					return -1
				}
				flag.remove(flagNeedColon).add(flagNeedComma, flagNeedEnd) // clean flagColon
				json.offset = end
			case ',':
				if !flag.contains(flagNeedComma) {
					return -json.offset - 1
				}
				flag.remove(flagNeedComma, flagNeedEnd).add(flagNeedKey)
				json.offset++
			case '}':
				if !flag.contains(flagNeedEnd) {
					return -json.offset - 1
				}
				json.offset++
				return json.offset
			}

		}
	}
	now := json.offset
	end := validEnd()
	if end < 0 {
		if json.err == nil {
			json.err = SchemaError{Object, -(end + 1), json.data}
		}
		end = -1
	}
	json.offset = now
	return end
}

func (json *JSON) unsafeBlockEnd(left, right byte) int {
	level := 0
	now := json.offset
	defer func() {
		json.offset = now
	}()

	n := len(json.data)

	for ; json.offset < n; json.offset++ {
		switch json.data[json.offset] {
		case left:
			level++
		case right:
			level--
			if level == 0 {
				return json.offset + 1
			}
		case '"':
			end := json.validStringEnd()
			if end == -1 {
				return -1
			}
			// json.offset will +1 always, so json.offset == end next time
			json.offset = end - 1
		}

	}

	return -1
}

func (json *JSON) unsafeArrayEnd() int {
	return json.unsafeBlockEnd('[', ']')
}

func (json *JSON) unsafeObjectEnd() int {
	return json.unsafeBlockEnd('{', '}')
}

type SchemaError struct {
	kind   Kind
	offset int
	data   []byte
}

func (e SchemaError) Error() string {
	start := e.offset - 5
	if start < 0 {
		start = 0
	}
	end := e.offset + 5
	if end > len(e.data) {
		end = len(e.data)
	}

	return fmt.Sprintf("Json schema error when parsing kind(%s), context near: |%s|", e.kind, string(e.data[start:end]))
}

func (json *JSON) Err() error {
	return json.err
}

func (json *JSON) ObjectIndex(key string) *JSON {
	json.mustBe(Object)
	validIndex := func() (int, *JSON) {
		match := false
		flag := new(flag).add(flagNeedStart)
		for {
			c, ok := json.nextToken()
			if !ok {
				return -json.offset - 1, nil
			}
			switch c {
			case '{':
				if !flag.contains(flagNeedStart) {
					return -json.offset - 1, nil
				}
				flag.remove(flagNeedStart).add(flagNeedKey, flagNeedEnd)
				json.offset++
			case '"':
				if !flag.contains(flagNeedKey) {
					return -json.offset - 1, nil
				}
				end := json.validStringEnd()
				if end == -1 {
					return -1, nil
				}
				if k := string(json.data[json.offset+1 : end-1]); k == key {
					match = true
				} else {
					match = false
				}

				flag.remove(flagNeedKey, flagNeedEnd).add(flagNeedColon) // clean
				json.offset = end                                        // move to end
			case ':':
				if !flag.contains(flagNeedColon) {
					return -json.offset - 1, nil
				}
				json.offset++
				var end int
				if !match {
					end, _ = json.unsafeValueEnd()
				} else {
					end, _ = json.validValueEnd()
				}

				if end == -1 {
					return -1, nil
				}

				if match {
					json.head = json.offset
					json.tail = end
					return json.offset, json
				}

				flag.remove(flagNeedColon).add(flagNeedComma, flagNeedEnd) // clean flagColon
				json.offset = end
			case ',':
				if !flag.contains(flagNeedComma) {
					return -json.offset - 1, nil
				}
				flag.remove(flagNeedComma, flagNeedEnd).add(flagNeedKey)
				json.offset++
			case '}':
				if !flag.contains(flagNeedEnd) {
					return -json.offset - 1, nil
				}
				json.offset++
				return json.offset, nil
			}

		}
	}
	now := json.offset
	defer func() {
		json.offset = now
	}()

	end, j := validIndex()
	if end < 0 {
		if json.err == nil {
			json.err = SchemaError{Object, -(end + 1), json.data}
		}
	} else if j == nil {
		json.err = fmt.Errorf("object has no such key %s", key)
	}

	return j
}

func (json *JSON) Index(index int) *JSON {
	json.mustBe(Array)
	validIndex := func() (int, *JSON) {
		flag := new(flag).add(flagNeedStart)
		i := 0
		for {
			c, ok := json.nextToken()
			if !ok {
				return -json.offset - 1, nil
			}
			switch c {

			case ',':
				if !flag.contains(flagNeedComma) {
					return -json.offset - 1, nil
				}
				flag.remove(flagNeedComma, flagNeedEnd).add(flagNeedValue)
				json.offset++
			case ']':
				if !flag.contains(flagNeedEnd) {
					return -json.offset - 1, nil
				}
				json.offset++
				return json.offset, nil
			case '[':
				if flag.contains(flagNeedValue) {
					goto DefaultCase
				}

				if !flag.contains(flagNeedStart) {
					return -json.offset - 1, nil
				}
				flag.remove(flagNeedStart).add(flagNeedValue, flagNeedEnd)
				json.offset++
				break
			DefaultCase: // label
				fallthrough
			default:
				if !flag.contains(flagNeedValue) {
					return -json.offset - 1, nil
				}
				var end int
				if i == index {
					end, _ = json.validValueEnd()
				} else {
					end, _ = json.unsafeValueEnd()
				}

				if end == -1 {
					return -1, nil
				}

				if i == index {
					json.head = json.offset
					json.tail = end
					return json.offset, json
				}
				i++
				flag.remove(flagNeedValue).add(flagNeedComma, flagNeedEnd)
				json.offset = end
			}
		}
	}

	now := json.offset
	defer func() {
		json.offset = now
	}()

	end, j := validIndex()
	if end < 0 {
		if json.err == nil {
			json.err = SchemaError{Object, -(end + 1), json.data}
		}
	} else if j == nil {
		json.err = fmt.Errorf("array out of range")
	}

	return j
}

func (json *JSON) Get(keys ...interface{}) (*JSON, error) {
	if len(keys) == 0 {
		return json, nil
	}
	kind := json.Predict()
	switch key := keys[0].(type) {
	case string:
		if kind != Object {
			return nil, fmt.Errorf("can not get %s from json type %s", key, kind)
		}
		jj := json.ObjectIndex(key)
		if jj == nil {
			return nil, json.Err()
		}
		return jj.Get(keys[1:]...)
	case int:
		if kind != Array {
			return nil, fmt.Errorf("can not get %d from json type %s", key, kind)
		}
		jj := json.Index(key)
		if jj == nil {
			return nil, json.Err()
		}
		return jj.Get(keys[1:]...)
	default:
		return nil, fmt.Errorf("%v is not string or int", keys[0])
	}
}

func (json *JSON) String() string {
	if json.head == 0 && json.tail == 0 {
		return string(json.data)
	}
	return string(json.data[json.head:json.tail])
}
