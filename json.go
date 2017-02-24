package jzon

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
	"strconv"
)

// Kind defines the type of JSON
type Kind uint

const (
	// Invalid is an invalid type of JSON
	Invalid Kind = iota
	// Object is an JSON object
	Object
	// Array is an JSON array
	Array
	// Number is an JSON number
	Number
	// String is an JSON string
	String
	// Bool is an JSON bool
	Bool
	// Null is an JSON null
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

// JSON is the basic struct representation
type JSON struct {
	data      []byte
	offset    int
	head      int
	tail      int
	limitHead int
	limitTail int
	err       error
}

// FromString returns an JSON from string
func FromString(data string) *JSON {
	return FromBytes([]byte(data))
}

// FromBytes returns a JSON from byte slice
func FromBytes(data []byte) *JSON {
	json := &JSON{
		data:      data,
		offset:    0,
		head:      0,
		tail:      len(data),
		limitHead: 0,
		limitTail: len(data),
	}
	return json
}

// FromReader reads bytes from reader, then build JSON from it
func FromReader(r io.Reader) (*JSON, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	json := &JSON{
		data:      buf,
		offset:    0,
		head:      0,
		tail:      len(buf),
		limitHead: 0,
		limitTail: len(buf),
	}
	return json, nil
}

// Reset resets JSON to be reused
func (json *JSON) Reset() *JSON {
	json.offset = 0
	json.head = 0
	json.tail = len(json.data)
	json.limitHead = len(json.data)
	json.limitTail = len(json.data)
	json.err = nil
	json.nextToken()
	return json
}

func (json *JSON) limit(head, tail int) {
	json.limitHead = head

	if tail > len(json.data) {
		tail = len(json.data)
	}
	json.limitTail = tail
	json.offset = json.limitHead
}

func (json *JSON) unlimit() {
	json.limitHead = 0
	json.limitTail = len(json.data)
}

func (json *JSON) CheckValid() error {
	end, _ := json.validValueEnd()
	if end == -1 {
		return json.err
	}
	return nil
}

// nextToken returns the byte read at index i, move offset to i
// if a valid byte is found
func (json *JSON) nextToken() (byte, bool) {

	if json.limitTail == 0 {
		json.limitTail = len(json.data)
	}

	if json.offset >= json.limitTail {
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
	json.offset = json.limitTail
	return 0, false
}

// readNextToken is like nextToken
func (json *JSON) readNextToken() (byte, bool) {

	c, ok := json.nextToken()
	if ok {
		json.offset++
	}
	return c, ok
}

// Predict predicts the type of json according to the next token
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

// unsafeValueEnd returns the end of json value,
// it is unsafe because it doesn't check syntax strictly
// it is efficient if you don't care what the value is
func (json *JSON) unsafeValueEnd() (int, Kind) {
	switch kind := json.Predict(); kind {
	case Object:
		return json.unsafeObjectEnd(), kind
	case Array:
		return json.unsafeArrayEnd(), kind
	case Number:
		return json.unsafeNumberEnd(), kind
	case String:
		return json.validStringEnd(), kind
	case Bool, Null:
		return json.validLiteralValueEnd(), kind
	default:
		json.err = SyntaxError{Invalid, json.offset, json.data}
		return -1, Invalid
	}
}

// validValueEnd return the end of json value,
// it is safe because it check syntax strictly
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
		json.err = SyntaxError{Invalid, json.offset, json.data}
		return -1, Invalid
	}
}

func (json *JSON) validStringEnd() int {

	if json.limitTail == 0 {
		json.limitTail = len(json.data)
	}

	n := json.limitTail

	validEnd := func() int {
		data := json.data
		// https://tools.ietf.org/html/rfc7159#section-7
		// escaped := false
		for i := json.offset; i < n; {
			switch c := data[i]; {
			case c == '\\':
				// look one more byte
				if i+1 >= n {
					return -i - 1
				}
				// escaped = true
				switch data[i+1] {
				case 'u':
					// maybe \uXXXX
					// need 4 hexdemical digit
					if getu4(data[i:]) == -1 {
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

			case c == '"':
				// match end
				return i + 1
			case c < ' ':
				// control characters are invalid
				return -i - 1
			default:
				i++
			}
		}
		// can not find string end
		return -n - 1
	}
	json.offset++
	end := validEnd()
	if end < 0 {
		json.err = SyntaxError{String, json.offset + 1 - (end + 1), json.data}
		return -1
	}
	json.offset--
	return end
}

func (json *JSON) validLiteralValueEnd() int {
	// https://github.com/golang/go/commit/69cd91a5981c49eaaa59b33196bdb5586c18d289
	if json.limitTail == 0 {
		json.limitTail = len(json.data)
	}

	n := json.limitTail

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

	if json.offset >= n {
		json.err = SyntaxError{Invalid, json.offset, json.data}
		return -1
	}

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

	json.err = SyntaxError{kind, json.offset, json.data}

	return -1
}

func (json *JSON) validNumberEnd() int {
	if json.limitTail == 0 {
		json.limitTail = len(json.data)
	}

	n := json.limitTail

	lookOneMore := func(j int) (index int, digit bool, ok bool) {
		next := j + 1
		if next >= n {
			return j, false, false
		}

		return next, isDigit(json.data[next]), true
	}

	validEnd := func() (int, flag) {
		var flag flag
		data := json.data
		i := json.offset

		if i >= n {
			return -1, flag
		}

		// first of all
		if data[i] == '-' {
			flag = add(flag, flagIsNegative)
			i++
		}

		if i >= n {
			// -
			return -1, flag
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
				if contains(flag, flagIsScientific) {
					// 123ee
					return -i - 1, flag
				}
				flag = add(flag, flagIsScientific)

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
				if contains(flag, flagIsFloat) || contains(flag, flagIsScientific) {
					// 1. get more than one dot
					// 2. or already get one e or E, only digit is allowed after e in scientific notation
					return -i - 1, flag
				}
				flag = add(flag, flagIsFloat)
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

	end, _ := validEnd()

	if end < 0 {
		json.err = SyntaxError{Number, -end, json.data}
		return -1
	}

	return end
}

func (json *JSON) validArrayEnd() int {
	validEnd := func() int {
		flag := flagNeedStart
		for {
			c, ok := json.nextToken()
			if !ok {
				return -json.offset - 1
			}
			switch c {

			case ',':
				if !contains(flag, flagNeedComma) {
					return -json.offset - 1
				}
				flag = remove(flag, flagNeedComma, flagNeedEnd)
				flag = add(flag, flagNeedValue)
				json.offset++
			case ']':
				if !contains(flag, flagNeedEnd) {
					return -json.offset - 1
				}
				json.offset++
				return json.offset
			case '[':
				if contains(flag, flagNeedValue) {
					// [[1,2], [3,4]]
					goto DefaultCase
				}

				if !contains(flag, flagNeedStart) {
					return -json.offset - 1
				}
				flag = remove(flag, flagNeedStart)
				flag = add(flag, flagNeedValue, flagNeedEnd)
				json.offset++
				break
			DefaultCase: // label
				fallthrough
			default:
				if !contains(flag, flagNeedValue) {
					return -json.offset - 1
				}
				end, _ := json.validValueEnd()
				if end == -1 {
					return -1
				}
				flag = remove(flag, flagNeedValue)
				flag = add(flag, flagNeedComma, flagNeedEnd)
				json.offset = end
			}
		}
	}

	now := json.offset
	end := validEnd()
	if end < 0 {
		if json.err == nil {
			json.err = SyntaxError{Array, -(end + 1), json.data}
		}
		end = -1
	}

	json.offset = now
	return end
}

func (json *JSON) validObjectEnd() int {
	validEnd := func() int {
		flag := flagNeedStart
		for {
			c, ok := json.nextToken()
			if !ok {
				return -json.offset - 1
			}
			switch c {
			case '{':
				if !contains(flag, flagNeedStart) {
					return -json.offset - 1
				}
				flag = remove(flag, flagNeedStart)
				flag = add(flag, flagNeedKey, flagNeedEnd)
				json.offset++
			case '"':
				if !contains(flag, flagNeedKey) {
					return -json.offset - 1
				}
				end := json.validStringEnd()
				if end == -1 {
					return -1
				}
				flag = remove(flag, flagNeedKey, flagNeedEnd)
				flag = add(flag, flagNeedColon)
				json.offset = end // move to end
			case ':':
				if !contains(flag, flagNeedColon) {
					return -json.offset - 1
				}
				json.offset++
				end, _ := json.validValueEnd()
				if end == -1 {
					return -1
				}
				flag = remove(flag, flagNeedColon)
				flag = add(flag, flagNeedComma, flagNeedEnd) // clean flagColon
				json.offset = end
			case ',':
				if !contains(flag, flagNeedComma) {
					return -json.offset - 1
				}
				flag = remove(flag, flagNeedComma, flagNeedEnd)
				flag = add(flag, flagNeedKey)
				json.offset++
			case '}':
				if !contains(flag, flagNeedEnd) {
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
			json.err = SyntaxError{Object, -(end + 1), json.data}
		}
		end = -1
	}
	json.offset = now
	return end
}

// unsafeBlockEnd finds end of the data structure, array or object.
// it is unsafe bucause it only check the nested symbol pair
func (json *JSON) unsafeBlockEnd(left, right byte) int {
	now := json.offset
	defer func() {
		json.offset = now
	}()

	if json.limitTail == 0 {
		json.limitTail = len(json.data)
	}

	n := json.limitTail

	level := 0
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
	var kind Kind
	if left == '{' {
		kind = Object
	} else if left == '[' {
		kind = Array
	}
	json.err = SyntaxError{kind, json.offset, json.data}
	return -1
}

func (json *JSON) unsafeArrayEnd() int {
	return json.unsafeBlockEnd('[', ']')
}

func (json *JSON) unsafeObjectEnd() int {
	return json.unsafeBlockEnd('{', '}')
}

func (json *JSON) unsafeNumberEnd() int {
	if json.limitTail == 0 {
		json.limitTail = len(json.data)
	}

	n := json.limitTail

	if json.offset == n {
		return -1
	}

	i := json.offset

	for ; i < n; i++ {
		switch json.data[i] {
		case ' ', '\n', '\r', '\t', ',', '}', ']':
			// here is the valid end
			return i
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '+', '-', 'e', 'E', '.':
			continue
		default:
			json.err = SyntaxError{Number, i, json.data}
			return -1
		}
	}
	return i
}

// ObjectIndex finds value index i by object key, then move offset to i,
// if not found, return error
// if occur syntax error, return error
func (json *JSON) ObjectIndex(key string) error {
	json.mustBe(Object)
	validIndex := func() int {
		match := false
		flag := flagNeedStart
		for {
			c, ok := json.nextToken()
			if !ok {
				return -json.offset - 1
			}
			switch c {
			case '{':
				if !contains(flag, flagNeedStart) {
					return -json.offset - 1
				}
				flag = remove(flag, flagNeedStart)
				flag = add(flag, flagNeedKey, flagNeedEnd)
				json.offset++
			case '"':
				if !contains(flag, flagNeedKey) {
					return -json.offset - 1
				}
				end := json.validStringEnd()
				if end == -1 {
					return -1
				}
				s, _ := unquote(json.data[json.offset:end])
				if k := string(s); k == key {
					match = true
				} else {
					match = false
				}

				flag = remove(flag, flagNeedKey, flagNeedEnd)
				flag = add(flag, flagNeedColon)
				json.offset = end // move to end
			case ':':
				if !contains(flag, flagNeedColon) {
					return -json.offset - 1
				}
				json.offset++
				var end int
				if !match {
					end, _ = json.unsafeValueEnd()
				} else {
					end, _ = json.validValueEnd()
				}

				if end == -1 {
					return -1
				}

				if match {
					json.head = json.offset
					json.tail = end
					return json.offset
				}

				flag = remove(flag, flagNeedColon)
				flag = add(flag, flagNeedComma, flagNeedEnd)
				json.offset = end
			case ',':
				if !contains(flag, flagNeedComma) {
					return -json.offset - 1
				}
				flag = remove(flag, flagNeedComma, flagNeedEnd)
				flag = add(flag, flagNeedKey)
				json.offset++
			case '}':
				if !contains(flag, flagNeedEnd) {
					return -json.offset - 1
				}
				json.offset++
				json.err = fmt.Errorf("object: key[%s] not found", key)
				return -json.offset
			}

		}
	}

	end := validIndex()
	if end < 0 && json.err == nil {
		json.err = SyntaxError{Object, -(end + 1), json.data}
	}

	return json.err
}

// Index finds value index i by array index, then move offset to i,
// if out of array range, return error
// if occur syntax error, return error
func (json *JSON) Index(index int) error {
	json.mustBe(Array)
	validIndex := func() int {
		flag := flagNeedStart
		i := 0
		for {
			c, ok := json.nextToken()
			if !ok {
				return -json.offset - 1
			}
			switch c {
			case ',':
				if !contains(flag, flagNeedComma) {
					return -json.offset - 1
				}
				flag = remove(flag, flagNeedComma, flagNeedEnd)
				flag = add(flag, flagNeedValue)
				json.offset++
			case ']':
				if !contains(flag, flagNeedEnd) {
					return -json.offset - 1
				}
				json.offset++
				json.err = fmt.Errorf("array: index[%d] out of range", index)
				return json.offset
			case '[':
				if contains(flag, flagNeedValue) {
					// [[1,2],[3,4]]
					goto DefaultCase
				}
				if !contains(flag, flagNeedStart) {
					return -json.offset - 1
				}
				flag = remove(flag, flagNeedStart)
				flag = add(flag, flagNeedValue, flagNeedEnd)
				json.offset++
				break
			DefaultCase: // label
				fallthrough
			default:
				if !contains(flag, flagNeedValue) {
					return -json.offset - 1
				}
				var end int
				if i == index {
					end, _ = json.validValueEnd()
				} else {
					end, _ = json.unsafeValueEnd()
				}

				if end == -1 {
					return -1
				}

				if i == index {
					json.head = json.offset
					json.tail = end
					return json.offset
				}

				flag = remove(flag, flagNeedValue)
				flag = add(flag, flagNeedComma, flagNeedEnd)
				json.offset = end
				i++
			}
		}
	}

	end := validIndex()
	if end < 0 && json.err == nil {
		json.err = SyntaxError{Object, -(end + 1), json.data}
	}

	return json.err
}

// Path moves offset to given keys path
func (json *JSON) Path(keys ...interface{}) error {
	if len(keys) == 0 {
		return nil
	}

	for _, key := range keys {
		kind := json.Predict()

		switch k := key.(type) {
		case string:
			if kind != Object {
				json.err = fmt.Errorf("can not get key[%s] from json type %s", key, kind)
				return json.err
			}
			err := json.ObjectIndex(k)
			if err != nil {
				return err
			}
		case int:
			if kind != Array {
				json.err = fmt.Errorf("can not get index[%d] from json type %s", key, kind)
				return json.err
			}
			err := json.Index(k)
			if err != nil {
				return err
			}
		default:
			json.err = fmt.Errorf("%v is not string or int", keys[0])
			return json.err
		}
	}
	return nil
}

// ParseInt64 parses an int Number json value to int64
func (json *JSON) ParseInt64() (int64, error) {
	if json.tail <= 0 {
		json.tail = len(json.data)
	}

	json.offset = json.head
	kind := json.Predict()
	if kind != Number {
		return 0, fmt.Errorf("ParseInt64: Can not parse %s JSON to int", kind)
	}

	return strconv.ParseInt(string(json.data[json.head:json.tail]), 10, 64)
}

// ParseFloat parses an float Number json value to float64
func (json *JSON) ParseFloat() (float64, error) {
	if json.tail <= 0 {
		json.tail = len(json.data)
	}

	json.offset = json.head
	kind := json.Predict()
	if kind != Number {
		return 0, fmt.Errorf("ParseFloat: Can not parse %s JSON to float", kind)
	}

	return strconv.ParseFloat(string(json.data[json.head:json.tail]), 64)
}

// ParseString parses an String json value to float64
func (json *JSON) ParseString() (string, error) {
	if json.tail <= 0 {
		json.tail = len(json.data)
	}
	json.offset = json.head
	kind := json.Predict()
	if kind != String {
		return "", fmt.Errorf("ParseString: Can not parse %s JSON to string", kind)
	}
	s, ok := unquote(json.data[json.head:json.tail])
	if !ok {
		return "", errors.New("ParseString: unquote string error")
	}
	return s, nil
}

// ParseBoolean parses an Bool json value to bool
func (json *JSON) ParseBoolean() (bool, error) {
	if json.tail <= 0 {
		json.tail = len(json.data)
	}
	json.offset = json.head
	kind := json.Predict()
	if kind != Bool {
		return false, fmt.Errorf("ParseBoolean: Can not parse %s JSON to bool", kind)
	}

	s := json.data[json.head:json.tail]

	if bytes.Equal(s, trueBytes) {
		return true, nil
	} else if bytes.Equal(s, falseBytes) {
		return false, nil
	}

	return false, fmt.Errorf("ParseBoolean: no valid boolean string")
}

// Kind returns current kind of json represented by json.data[head:tail]
func (json *JSON) Kind() Kind {
	now := json.offset
	defer func() {
		json.offset = now
	}()
	json.offset = json.head

	return json.Predict()
}

func (json *JSON) String() string {
	if json.head == 0 && json.tail == 0 {
		return string(json.data)
	}
	return string(json.data[json.head:json.tail])
}

// Err returns any error when parsing json syntax
func (json *JSON) Err() error {
	return json.err
}

// mustBe assert the JSON must be expected type,
// it will panic otherwise.
func (json *JSON) mustBe(expected Kind) {
	kind := json.Predict()
	if kind != expected {
		panic(&KindError{methodName(), kind})
	}
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

func isDigit(c byte) bool {
	switch c {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}
