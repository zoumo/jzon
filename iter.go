package jzon

// ObjectIter is an iterable on object type json
type ObjectIter struct {
	*JSON
	key string
	len int
}

// Reset resets the ObjectIter then you can use it again
func (iter *ObjectIter) Reset() {
	iter.offset = 0
	iter.key = ""
}

// Next finds the next key and value pair of object.
// If not return false
// example:
// for iter.Next() {
// 	  key := iter.Key()
// 	  value := iter.Value()
// 	  // do something
// }
func (iter *ObjectIter) Next() bool {
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
			iter.key = string(iter.data[iter.offset+1 : end-1])
			iter.offset = end
		case ':':
			iter.offset++
			end, _ := iter.unsafeValueEnd()
			iter.head = iter.offset
			iter.tail = end
			iter.offset = end
			break Loop
		case ',':
			iter.offset++
		case '}':
			iter.offset++
			return false
		}

	}
	return true
}

// Len returns the length of object using Next() api
// and cache it
func (iter *ObjectIter) Len() int {
	if iter.len > 0 {
		return iter.len
	}
	for iter.Next() {
		iter.len++
	}
	iter.Reset()
	return iter.len
}

// Key returns current key
func (iter *ObjectIter) Key() string {
	return iter.key
}

// Value returns current value
func (iter *ObjectIter) Value() *JSON {
	return iter.JSON
}

// ----------------------------------------------------------------------------

// ArrayIter is an iterable on array type json
type ArrayIter struct {
	*JSON
	index int
	len   int
}

// Reset resets the ArrayIter then you can use it again
func (iter *ArrayIter) Reset() {
	iter.offset = 0
	iter.index = -1
}

// Next finds the next value of array.
// If not return false
// example:
// for iter.Next() {
// 	  index := iter.Index()
// 	  value := iter.Value()
// 	  // do something
// }
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
			iter.index++
			iter.head = iter.offset
			iter.tail = end
			// iter.value = iter.JSON
			iter.offset = end

			break Loop
		}

	}
	return true
}

// Len returns the length of object using Next() api
// and cache it
func (iter *ArrayIter) Len() int {
	if iter.len > 0 {
		return iter.len
	}
	for iter.Next() {
		iter.len++
	}
	iter.Reset()
	return iter.len
}

// Index returns current index
func (iter *ArrayIter) Index() int {
	return iter.index
}

// Value returns current value
func (iter *ArrayIter) Value() *JSON {
	return iter.JSON
}

// ----------------------------------------------------------------------------

// Object returns an ObjectIter which is an iterable on the object after valified.
func (json *JSON) Object() (*ObjectIter, error) {
	json.mustBe(Object)
	// it is very important to valify object before using it
	end := json.validObjectEnd()
	if end == -1 {
		return nil, json.err
	}
	return &ObjectIter{
		JSON: &JSON{
			data: json.data[json.offset:end],
		},
	}, nil
}

// UnsafeObject returns an ObjectIter which is an iterable
// on the object without valified. It is faster than Object() function,
// but it is unsafe, you should make sure the object is valid by your self.
func (json *JSON) UnsafeObject() (*ObjectIter, error) {
	json.mustBe(Object)
	end := json.unsafeObjectEnd()
	if end == -1 {
		return nil, json.err
	}
	return &ObjectIter{
		JSON: &JSON{
			data: json.data[json.offset:end],
		},
	}, nil
}

// Array returns an ArrayIter which is an iterable on the array after valified.
func (json *JSON) Array() (*ArrayIter, error) {
	json.mustBe(Array)
	end := json.validArrayEnd()
	if end == -1 {
		return nil, json.err
	}
	return &ArrayIter{
		JSON: &JSON{
			data: json.data[json.offset+1 : end-1],
		},
		index: -1,
	}, nil
}

// UnsafeArray returns an ArrayIter which is an iterable
// on the array without valified.It is faster than Array() function,
// but it is unsafe, you should make sure the array is valid by your self.
func (json *JSON) UnsafeArray() (*ArrayIter, error) {
	json.mustBe(Array)
	end := json.unsafeArrayEnd()
	if end == -1 {
		return nil, json.err
	}
	return &ArrayIter{
		JSON: &JSON{
			data: json.data[json.offset+1 : end-1],
		},
	}, nil
}
