package jzon

// ObjectIter is an iterable on object type JSON
type ObjectIter struct {
	*JSON
	key       string
	len       int
	keysCache []string
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
			s, _ := unquote(iter.data[iter.offset:end])
			iter.key = string(s)
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

// Keys returns all keys and store in cache
func (iter *ObjectIter) Keys() []string {
	if iter.keysCache == nil {
		iter.keysCache = make([]string, 0)
	}
	if len(iter.keysCache) == 0 {
		iter.len = 0
		for iter.Next() {
			iter.keysCache = append(iter.keysCache, iter.key)
			iter.len++
		}
		iter.Reset()
	}
	return iter.keysCache
}

// HasKey checks whether iter cantains the given key
func (iter *ObjectIter) HasKey(k string) bool {
	for _, key := range iter.Keys() {
		if key == k {
			return true
		}
	}
	return false
}

// ----------------------------------------------------------------------------

// ArrayIter is an iterable on array type JSON
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

func (json *JSON) objectIter(valify bool) (*ObjectIter, error) {
	now := json.offset
	defer func() {
		json.offset = now
	}()
	json.offset = json.head

	json.mustBe(Object)

	if json.tail <= 0 {
		if valify {
			// it is very important to valify object before using it
			json.tail = json.validObjectEnd()
		} else {
			json.tail = json.unsafeObjectEnd()
		}
		if json.tail == -1 {
			return nil, json.err
		}
	}

	return &ObjectIter{
		JSON: &JSON{
			data: json.data[json.head:json.tail],
		},
	}, nil
}

// Object returns an ObjectIter which is an iterable on the object after valified.
func (json *JSON) Object() (*ObjectIter, error) {
	return json.objectIter(true)
}

// UnsafeObject returns an ObjectIter which is an iterable
// on the object without valified. It is 2.x faster than Object() function,
// but it is unsafe, you should make sure the object is valid by your self.
func (json *JSON) UnsafeObject() (*ObjectIter, error) {
	return json.objectIter(false)
}

func (json *JSON) arrayIter(valify bool) (*ArrayIter, error) {
	now := json.offset
	defer func() {
		json.offset = now
	}()
	json.offset = json.head

	json.mustBe(Array)
	if json.tail <= 0 {
		if valify {
			json.tail = json.validArrayEnd()
		} else {
			json.tail = json.unsafeArrayEnd()
		}
		if json.tail == -1 {
			return nil, json.err
		}
	}
	return &ArrayIter{
		JSON: &JSON{
			data: json.data[json.head+1 : json.tail-1],
		},
		index: -1,
	}, nil
}

// Array returns an ArrayIter which is an iterable on the array after valified.
func (json *JSON) Array() (*ArrayIter, error) {
	return json.arrayIter(true)
}

// UnsafeArray returns an ArrayIter which is an iterable
// on the array without valified. It is 2.x faster than Array() function,
// but it is unsafe, you should make sure the array is valid by your self.
func (json *JSON) UnsafeArray() (*ArrayIter, error) {
	return json.arrayIter(false)
}
