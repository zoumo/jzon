package jzon

import "testing"

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
