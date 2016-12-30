package jzon

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"unsafe"
)

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

type Parser struct {
	data   []byte
	offset int
	Kind   Kind
	err    error
	stack  stack
}

func ParseBytes(data []byte) *Parser {
	p := &Parser{
		data:   data,
		offset: 0,
		stack:  make(stack, 0),
	}
	return p
}

func ParseReader(r io.Reader) (*Parser, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	p := &Parser{
		data:   buf,
		offset: 0,
		stack:  make(stack, 0),
	}
	return p, nil
}

func (p *Parser) Reuse(data []byte) *Parser {
	p.data = data
	p.offset = 0
	p.stack = make(stack, 0)
	p.nextToken()
	return p
}

// nextToken returns the byte read at index i, move offset to i
// if a valid byte is found
func (p *Parser) nextToken() (byte, bool) {
	if p.offset >= len(p.data) {
		return 0, false
	}
	for i, c := range p.data[p.offset:] {
		switch c {
		case ' ', '\n', '\t', '\r':
			continue
		}
		p.offset += i
		return p.data[p.offset], true
	}
	p.offset = len(p.data)
	return 0, false
}

func (p *Parser) Predict() Kind {
	c, ok := p.nextToken()
	if !ok {
		return Invalid
	}
	switch c {
	case '{':
		return Object
	case '[':
		return Array
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

func (p *Parser) stringEnd() (int, bool) {

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

	validStringEnd := func(data []byte) (int, bool) {
		// https://tools.ietf.org/html/rfc7159#section-7
		n := len(data)
		escaped := false
		for i := 0; i < n; {
			switch data[i] {
			case '\\':
				// look one more byte
				if i+1 >= n {
					return -1, escaped
				}
				escaped = true
				switch data[i+1] {
				case 'u':
					// maybe \uXXXX
					// need 4 hexdemical digit
					if !validHexDigit(data[i+2:]) {
						// error string format
						return -1, escaped
					}
					i += 6
				case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
					// \", \\, \/, \b, \f, \n, \r, \t
					i += 2
				default:
					// error
					return -1, escaped
				}

			case '"':
				// match end
				return i + 1, escaped
			default:
				i++
			}
		}
		// can not find string end
		return -1, escaped
	}

	if p.data[p.offset] != '"' {
		return -1, false
	}

	end, es := validStringEnd(p.data[p.offset+1:])
	if end == -1 {
		return -1, es
	}

	return p.offset + 1 + end, es
}

func (p *Parser) literalValueEnd() int {
	// https://github.com/golang/go/commit/69cd91a5981c49eaaa59b33196bdb5586c18d289
	tokenEnd := func(data []byte) int {
		for i, c := range data {
			switch c {
			case ' ', '\n', '\r', '\t', ',', '}', ']':
				return i
			}
		}
		return len(data)
	}
	end := tokenEnd(p.data[p.offset:]) + p.offset
	switch end - p.offset {
	case 4:
		if bytes.Equal(p.data[p.offset:end], trueBytes) ||
			bytes.Equal(p.data[p.offset:end], nullBytes) {
			return end
		}
	case 5:
		if bytes.Equal(p.data[p.offset:end], falseBytes) {
			return end
		}
	}
	return -1
}

func (p *Parser) numberEnd() (int, *flag) {

	validNumberEnd := func(data []byte) (int, *flag) {
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
				return -1, flag
			}
			// start with 0, next byte is not digit
			i++
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			i++
		default:
			return -1, flag
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
					return -1, flag
				}
				flag.add(flagIsScientific)

				// +1
				j, isdigit, more := lookOneMore(i)
				if !more {
					// invalid format: 1.23e
					return -1, flag
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
					return -1, flag
				} else if !isdigit {
					// invalid format: 1.23ea
					return -1, flag
				}
				// 1.23e3
				i += 2
			case '.':
				if flag.contains(flagIsFloat) || flag.contains(flagIsScientific) {
					// 1. get more than one dot
					// 2. or already get one e or E, only digit is allowed after e in scientific notation
					return -1, flag
				}
				flag.add(flagIsFloat)
				_, isdigit, more := lookOneMore(i)
				if !more || !isdigit {
					// 1223.
					// 123.a
					return -1, flag
				}

				i += 2
			default:
				return -1, flag
			}
		}
		// meet string end
		return i, flag
	}
	end, f := validNumberEnd(p.data[p.offset:])

	if end == -1 {
		return -1, f
	}
	return p.offset + end, f
}

func (p *Parser) arrayEnd(verified bool) int {
	validArrayEnd := func() int {
		flag := new(flag).add(flagNeedStart)
		for {
			c, ok := p.nextToken()
			if !ok {
				return -1
			}
			switch c {
			case '[':
				if !flag.contains(flagNeedStart) {
					return -1
				}
				flag.remove(flagNeedStart).add(flagNeedValue, flagNeedEnd)
				p.offset++
			case ',':
				if !flag.contains(flagNeedComma) {
					return -1
				}
				flag.remove(flagNeedComma, flagNeedEnd).add(flagNeedValue)
				p.offset++
			case ']':
				if !flag.contains(flagNeedEnd) {
					return -1
				}
				p.offset++
				return p.offset
			default:
				if !flag.contains(flagNeedValue) {
					return -1
				}
				end, _ := p.valueEnd()
				if end == -1 {
					return -1
				}
				flag.remove(flagNeedValue).add(flagNeedComma, flagNeedEnd)
				p.offset = end
			}
		}
	}

	skipArrayEnd := func() int {
		level := 0
		i := p.offset
		n := len(p.data)

		for ; i < n; i++ {
			switch p.data[i] {
			case '{':
				level++
			case '}':
				level--
				if level == 0 {
					return i + 1
				}
			case '"':
				end, _ := p.stringEnd()
				if end == -1 {
					return -1
				}
				i = end - 1
			}

		}
		return -1
	}

	var end int
	now := p.offset
	if verified {
		end = validArrayEnd()
	} else {
		end = skipArrayEnd()
	}
	p.offset = now
	if end == -1 {
		return -1
	}
	return end

}

func (p *Parser) objectEnd() int {
	validObjectEnd := func() int {
		flag := new(flag).add(flagNeedStart)
		for {
			c, ok := p.nextToken()
			if !ok {
				return -1
			}
			switch c {
			case '{':
				if !flag.contains(flagNeedStart) {
					return -1
				}
				flag.remove(flagNeedStart).add(flagNeedKey, flagNeedEnd)
				p.offset++
			case '"':
				if !flag.contains(flagNeedKey) {
					return -1
				}
				end, _ := p.stringEnd()
				if end == -1 {
					return -1
				}
				flag.remove(flagNeedKey, flagNeedEnd).add(flagNeedColon) // clean
				p.offset = end                                           // move to end
			case ':':
				if !flag.contains(flagNeedColon) {
					return -1
				}
				p.offset++
				end, _ := p.valueEnd()
				if end == -1 {
					return -1
				}
				flag.remove(flagNeedColon).add(flagNeedComma, flagNeedEnd) // clean flagColon
				p.offset = end
			case ',':
				if !flag.contains(flagNeedComma) {
					return -1
				}
				flag.remove(flagNeedComma, flagNeedEnd).add(flagNeedKey)
				p.offset++
			case '}':
				if !flag.contains(flagNeedEnd) {
					return -1
				}
				p.offset++
				return p.offset
			}

		}
	}
	now := p.offset
	end := validObjectEnd()
	p.offset = now
	return end
}

func (p *Parser) readJSONValue() (*JSON, error) {
	switch p.Kind {
	case Invalid:
		return nil, fmt.Errorf("not a valid json value")
	case Object:
		return p.readObject()
	case Array:
		return p.readArray()
	case String:
		return p.readString()
	case Number:
		return p.readNumber()
	case Bool:
		return p.readBool()
	case Null:
		return p.readNull()
	default:
		return nil, fmt.Errorf("unknow byte")
	}
}

func (p *Parser) readNumber() (*JSON, error) {
	end, _ := p.numberEnd()
	if end == -1 {
		return nil, fmt.Errorf("invalid number")
	}
	obj := &JSON{
		data: p.data[p.offset:end],
		Kind: Number,
	}
	p.offset = end
	return obj, nil
}

func (p *Parser) readString() (*JSON, error) {
	end, _ := p.stringEnd()
	if end == -1 {
		return nil, fmt.Errorf("invalid string")
	}
	obj := &JSON{
		data: p.data[p.offset+1 : end-1],
		Kind: String,
	}
	p.offset = end
	return obj, nil
}

func (p *Parser) readBool() (*JSON, error) {
	end := p.literalValueEnd()
	if end == -1 {
		return nil, fmt.Errorf("invalid null")
	}
	obj := &JSON{
		data: p.data[p.offset:end],
		Kind: Bool,
	}
	p.offset = end
	return obj, nil
}

func (p *Parser) readNull() (*JSON, error) {
	end := p.literalValueEnd()
	if end == -1 {
		return nil, fmt.Errorf("invalid null")
	}
	obj := &JSON{
		data: p.data[p.offset:end],
		Kind: Null,
	}
	p.offset = end
	return obj, nil
}

const ()

type EndLoop struct{}

func (e EndLoop) Error() string {
	return "end loop"
}

func isEndLoop(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(EndLoop); ok {
		return true
	}
	return false
}

type SchemaError struct {
	offset int
	data   []byte
}

func (e SchemaError) Error() string {
	start := e.offset - 30
	if start < 0 {
		start = 0
	}
	end := e.offset + 30
	if end > len(e.data) {
		end = len(e.data)
	}

	return fmt.Sprintf("json schema error at position[%d], context: %s", e.offset, string(e.data[start:end]))
}

func (p *Parser) valueEnd() (int, Kind) {
	switch kind := p.Predict(); kind {
	case Invalid:
		return -1, kind
	case Object:

	case Array:

	case Number:
		end, _ := p.numberEnd()
		return end, kind
	case String:
		end, _ := p.stringEnd()
		return end, kind
	case Bool:
		return p.literalValueEnd(), kind
	case Null:
		return p.literalValueEnd(), kind
	}
	return -1, Invalid
}

func (p *Parser) Next() func() (string, []byte, error) {
	flag := new(flag).add(flagNeedStart)
	var start int
	var key string
	var value []byte
	// var err error
	return func() (string, []byte, error) {
		for {
			c, ok := p.nextToken()
			if !ok {
				return "", nil, fmt.Errorf("invalid object 2")
			}
			switch c {
			case '{':
				if !flag.contains(flagNeedStart) {
					return "key", nil, fmt.Errorf("invalid object start")
				}
				flag.remove(flagNeedStart).add(flagNeedKey, flagNeedEnd)
				start = p.offset
				p.offset++
			case '"':
				if !flag.contains(flagNeedKey) {
					return "", nil, fmt.Errorf("invalid object key")
				}
				// readKey
				end, _ := p.stringEnd()
				if end == -1 {
					return "", nil, fmt.Errorf("invalid object key")
				}
				key = string(p.data[p.offset:end])
				flag.remove(flagNeedKey, flagNeedEnd).add(flagNeedColon) // clean
			case ':':
				if !flag.contains(flagNeedColon) {
					return "", nil, fmt.Errorf("invalid object colon")
				}
				p.offset++
				now := p.offset

				value = p.data[now:p.offset]
				flag.remove(flagNeedColon).add(flagNeedComma, flagNeedEnd) // clean flagColon
				return key, value, nil
			case ',':
				if !flag.contains(flagNeedComma) {
					return "", nil, fmt.Errorf("invalid object Comma")
				}
				p.offset++
				flag.remove(flagNeedComma, flagNeedEnd).add(flagNeedKey)
			case '}':
				if !flag.contains(flagNeedEnd) {
					return "", nil, fmt.Errorf("invalid object end")
				}
				p.offset++
				return "", nil, EndLoop{}
			}

		}
	}

}

func (p *Parser) readObject() (*JSON, error) {
	flag := new(flag).add(flagNeedStart)
	var start int
	for {
		c, ok := p.nextToken()
		if !ok {
			return nil, fmt.Errorf("invalid object 2")
		}
		switch c {
		case '{':
			if !flag.contains(flagNeedStart) {
				return nil, fmt.Errorf("invalid object start")
			}
			p.stack.Push(&JSON{
				Kind: Object,
				ptr:  unsafe.Pointer(&map[string]*JSON{}),
			})
			flag.remove(flagNeedStart).add(flagNeedKey, flagNeedEnd)
			start = p.offset
			p.offset++
		case '"':
			if !flag.contains(flagNeedKey) {
				return nil, fmt.Errorf("invalid object key")
			}
			// readKey
			j, err := p.readString()
			if err != nil {
				return nil, err
			}
			p.stack.Push(j)
			flag.remove(flagNeedKey, flagNeedEnd).add(flagNeedColon) // clean
		case ':':
			if !flag.contains(flagNeedColon) {
				return nil, fmt.Errorf("invalid object colon")
			}
			p.offset++
			j, err := p.readJSONValue()
			if err != nil {
				return nil, err
			}

			// k/v
			k := p.stack.Pop()
			o := p.stack.Peek()
			(*(*map[string]*JSON)(o.ptr))[k.String()] = j

			flag.remove(flagNeedColon).add(flagNeedComma, flagNeedEnd) // clean flagColon
		case ',':
			if !flag.contains(flagNeedComma) {
				return nil, fmt.Errorf("invalid object Comma")
			}
			p.offset++
			flag.remove(flagNeedComma, flagNeedEnd).add(flagNeedKey)
		case '}':
			if !flag.contains(flagNeedEnd) {
				return nil, fmt.Errorf("invalid object end")
			}
			p.offset++
			j := p.stack.Pop()
			if j.Kind != Object {
				return nil, fmt.Errorf("need object")
			}
			j.data = p.data[start:p.offset]
			return j, nil
		}

	}

}

func (p *Parser) readArray() (*JSON, error) {
	flag := new(flag).add(flagNeedStart)
	var start int
	for {
		c, ok := p.nextToken()
		if !ok {
			return nil, fmt.Errorf("invalid object 2")
		}
		switch c {
		case '[':
			if !flag.contains(flagNeedStart) {
				return nil, fmt.Errorf("invalid object start")
			}
			p.stack.Push(&JSON{
				Kind: Array,
				ptr:  unsafe.Pointer(&[]*JSON{}),
			})
			flag.remove(flagNeedStart).add(flagNeedValue, flagNeedEnd)
			start = p.offset
			p.offset++
		case ',':
			if !flag.contains(flagNeedComma) {
				return nil, fmt.Errorf("invalid array comma")
			}
			p.offset++
			flag.remove(flagNeedComma, flagNeedEnd).add(flagNeedValue)
		case ']':
			if !flag.contains(flagNeedEnd) {
				return nil, fmt.Errorf("invalid array end")
			}
			p.offset++
			j := p.stack.Pop()
			if j.Kind != Array {
				return nil, fmt.Errorf("need array")
			}
			j.data = p.data[start:p.offset]
			return j, nil
		default:
			if !flag.contains(flagNeedValue) {
				return nil, fmt.Errorf("invalid array value")
			}
			j, err := p.readJSONValue()
			if err != nil {
				return nil, err
			}
			a := p.stack.Peek()
			t := append((*(*[]*JSON)(a.ptr)), j)
			a.ptr = unsafe.Pointer(&t)
			flag.remove(flagNeedValue).add(flagNeedComma, flagNeedEnd)
		}
	}
}

func (p *Parser) Parse() (*JSON, error) {
	p.Predict()
	j, err := p.readJSONValue()
	return j, err
}
