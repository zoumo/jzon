package jzon

import "testing"

// func TestJSON_Object(t *testing.T) {
// 	type fields struct {
// 		data   []byte
// 		offset int
// 		head   int
// 		tail   int
// 		err    error
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		want   *ObjectIter
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			json := &JSON{
// 				data:   tt.fields.data,
// 				offset: tt.fields.offset,
// 				head:   tt.fields.head,
// 				tail:   tt.fields.tail,
// 				err:    tt.fields.err,
// 			}
// 			if got := json.Object(); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("JSON.Object() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestJSON_Array(t *testing.T) {
// 	type fields struct {
// 		data   []byte
// 		offset int
// 		head   int
// 		tail   int
// 		err    error
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		want   *ArrayIter
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			json := &JSON{
// 				data:   tt.fields.data,
// 				offset: tt.fields.offset,
// 				head:   tt.fields.head,
// 				tail:   tt.fields.tail,
// 				err:    tt.fields.err,
// 			}
// 			if got := json.Array(); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("JSON.Array() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestObjectIter_Reset(t *testing.T) {
// 	type fields struct {
// 		JSON  *JSON
// 		Key   string
// 		Value *JSON
// 		len   int
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			iter := &ObjectIter{
// 				JSON:  tt.fields.JSON,
// 				Key:   tt.fields.Key,
// 				Value: tt.fields.Value,
// 				len:   tt.fields.len,
// 			}
// 			iter.Reset()
// 		})
// 	}
// }

var (
	array  = `[1,2,3,4,5]`
	object = `{"k1": "v1", "k2\tk2": true}`
)

func TestObjectIter_Next(t *testing.T) {
	type fields struct {
		JSON *JSON
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"1", fields{FromString(object)}, "k2\tk2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iter, _ := tt.fields.JSON.Object()
			for iter.Next() {
			}
			if got := iter.Key(); got != tt.want {
				t.Errorf("ObjectIter.Next() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObjectIter_Len(t *testing.T) {
	type fields struct {
		JSON *JSON
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{"1", fields{FromString(object)}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iter, _ := tt.fields.JSON.Object()
			if got := iter.Len(); got != tt.want {
				t.Errorf("ObjectIter.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArrayIter_Next(t *testing.T) {
	type fields struct {
		JSON *JSON
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{"1", fields{FromString(array)}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iter, _ := tt.fields.JSON.Array()
			for iter.Next() {
			}
			if got := iter.Index(); got != tt.want {
				t.Errorf("ArrayIter.Next() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArrayIter_Len(t *testing.T) {
	type fields struct {
		JSON *JSON
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{"1", fields{FromString(array)}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iter, _ := tt.fields.JSON.Array()
			if got := iter.Len(); got != tt.want {
				t.Errorf("ArrayIter.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}
