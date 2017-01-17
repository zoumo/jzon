package jzon

import "testing"

var (
	jsonStr = `{
    "string": "\"string\"",
    "true": true,
    "false": false,
    "number1": -0,
    "number2": -0.123e+01,
    "object": {
        "list": [],
        "k2": [1,2,34],
        "o": {},
        "o2": {
            "k1": "string"
        }
    },
    "list": [
        {
            "name": "n1",
            "code": 0,
            "values": [
                1,2,3,4
            ] 
        },
        {
            "name": "n2",
            "code": 1,
            "values": [
                1,2,3,4
            ] 
        }
    ]
}`
)

func TestJSON_validStringEnd(t *testing.T) {
	type fields struct {
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{"1", fields{[]byte(`"test"`)}, 6},
		{"2", fields{[]byte(`"te\"st"`)}, 8},
		{"3", fields{[]byte(`"te\\st"`)}, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JSON{
				data: tt.fields.data,
			}
			if got := j.validStringEnd(); got != tt.want {
				t.Errorf("JSON.validStringEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSON_validLiteralValueEnd(t *testing.T) {
	type fields struct {
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{"1", fields{[]byte("true")}, 4},
		{"2", fields{[]byte("false")}, 5},
		{"3", fields{[]byte("null")}, 4},
		{"4", fields{[]byte("nulll")}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JSON{
				data: tt.fields.data,
			}
			if got := j.validLiteralValueEnd(); got != tt.want {
				t.Errorf("JSON.validLiteralValueEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSON_validNumberEnd(t *testing.T) {
	type fields struct {
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// 0, -0, -00, -0., 0e 0.e, 0e, 0e0
		// 1, -1, -10, -1., 1e0, 1e-1, 1e, 1e0, 1e+1
		// 1.1, -1.1, 11..
		{"1", fields{[]byte("0")}, 1},
		{"2", fields{[]byte("-0")}, 2},
		{"3", fields{[]byte("+0")}, -1},
		{"4", fields{[]byte("-0.")}, -1},
		{"5", fields{[]byte("-01")}, -1},
		{"6", fields{[]byte("0e")}, -1},
		{"7", fields{[]byte("0.x")}, -1},
		{"8", fields{[]byte("0e0")}, 3},
		{"9", fields{[]byte("0e+0")}, 4},
		{"10", fields{[]byte("10")}, 2},
		{"11", fields{[]byte("1.")}, -1},
		{"12", fields{[]byte("1e")}, -1},
		{"13", fields{[]byte("1.e")}, -1},
		{"14", fields{[]byte("1e.")}, -1},
		{"15", fields{[]byte("1.1")}, 3},
		{"16", fields{[]byte("1e0")}, 3},
		{"17", fields{[]byte("1.1e0")}, 5},
		{"18", fields{[]byte("1.1.1")}, -1},
		{"19", fields{[]byte("1e1e1")}, -1},
		{"20", fields{[]byte("-0.123e+01")}, 10},
		{"21", fields{[]byte("-123e+0.1")}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JSON{
				data: tt.fields.data,
			}
			if got := j.validNumberEnd(); got != tt.want {
				t.Errorf("Parser.numberEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSON_unsafeNumberEnd(t *testing.T) {
	type fields struct {
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// 0, -0, -00, -0., 0e 0.e, 0e, 0e0
		// 1, -1, -10, -1., 1e0, 1e-1, 1e, 1e0, 1e+1
		// 1.1, -1.1, 11..
		{"1", fields{[]byte("0")}, 1},
		{"2", fields{[]byte("-0")}, 2},
		{"3", fields{[]byte("0.x")}, -1},
		{"4", fields{[]byte("0e0")}, 3},
		{"5", fields{[]byte("0e+0")}, 4},
		{"6", fields{[]byte("10")}, 2},
		{"7", fields{[]byte("1.1")}, 3},
		{"8", fields{[]byte("1e0")}, 3},
		{"9", fields{[]byte("1.1e0")}, 5},
		{"10", fields{[]byte("-0.123e+01")}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JSON{
				data: tt.fields.data,
			}
			if got := j.unsafeNumberEnd(); got != tt.want {
				t.Errorf("Parser.numberEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSON_unsafeArrayEnd(t *testing.T) {
	type fields struct {
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{"1", fields{[]byte(`["1", "2"]`)}, 10},
		{"2", fields{[]byte(`[1, 2]`)}, 6},
		{"3", fields{[]byte(`[{}, []]`)}, 8},
		{"4", fields{[]byte(`[[1,2], [3,4]]`)}, 14},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JSON{
				data: tt.fields.data,
			}
			if got := j.unsafeArrayEnd(); got != tt.want {
				t.Errorf("JSON.unsafeArrayEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSON_validArrayEnd(t *testing.T) {
	type fields struct {
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{"1", fields{[]byte(`["1", "2"]`)}, 10},
		{"2", fields{[]byte(`[1, 2]`)}, 6},
		{"3", fields{[]byte(`[{}, []]`)}, 8},
		{"4", fields{[]byte(`[[1,2], [3,4]]`)}, 14},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JSON{
				data: tt.fields.data,
			}
			if got := j.validArrayEnd(); got != tt.want {
				t.Errorf("JSON.validArrayEnd() = %v, want %v, err: %v", got, tt.want, j.err)
			}
		})
	}
}

func TestJSON_unsafeObjectEnd(t *testing.T) {
	type fields struct {
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{"1", fields{[]byte(`{"key": true}`)}, 13},
		{"2", fields{[]byte(`{"1":{}}`)}, 8},
		{"3", fields{[]byte(`{"1":[]}`)}, 8},
		{"4", fields{[]byte(jsonStr)}, len(jsonStr)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JSON{
				data: tt.fields.data,
			}
			if got := j.unsafeObjectEnd(); got != tt.want {
				t.Errorf("JSON.unsafeObjectEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSON_validObjectEnd(t *testing.T) {
	type fields struct {
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{"1", fields{[]byte(`{"key": true}`)}, 13},
		{"2", fields{[]byte(`{"1":{}}`)}, 8},
		{"3", fields{[]byte(`{"1":[]}`)}, 8},
		{"4", fields{[]byte(jsonStr)}, len(jsonStr)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JSON{
				data: tt.fields.data,
			}
			if got := j.validObjectEnd(); got != tt.want {
				t.Errorf("JSON.objectEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestJSON_ObjectIndex(t *testing.T) {
// 	type fields struct {
// 		data   []byte
// 		offset int
// 		head   int
// 		tail   int
// 		err    error
// 	}
// 	type args struct {
// 		key string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
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
// 			if err := json.ObjectIndex(tt.args.key); (err != nil) != tt.wantErr {
// 				t.Errorf("JSON.ObjectIndex() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestJSON_Index(t *testing.T) {
// 	type fields struct {
// 		data   []byte
// 		offset int
// 		head   int
// 		tail   int
// 		err    error
// 	}
// 	type args struct {
// 		index int
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
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
// 			if err := json.Index(tt.args.index); (err != nil) != tt.wantErr {
// 				t.Errorf("JSON.Index() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func TestJSON_Path(t *testing.T) {
	type fields struct {
		data []byte
	}
	type args struct {
		keys []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"1", fields{[]byte(jsonStr)}, args{[]interface{}{"list", 0, "name"}}, false},
		{"2", fields{[]byte(jsonStr)}, args{[]interface{}{"list", 2, "name"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			json := &JSON{
				data: tt.fields.data,
			}
			if err := json.Path(tt.args.keys...); (err != nil) != tt.wantErr {
				t.Errorf("JSON.Path() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJSON_ParseInt64(t *testing.T) {
	type fields struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		{"1", fields{[]byte(`123`)}, 123, false},
		{"2", fields{[]byte(`1.23e2`)}, 0, true},
		{"3", fields{[]byte(`123e`)}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			json := &JSON{
				data: tt.fields.data,
			}
			got, err := json.ParseInt64()
			if (err != nil) != tt.wantErr {
				t.Errorf("JSON.ParseInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("JSON.ParseInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSON_ParseFloat(t *testing.T) {
	type fields struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    float64
		wantErr bool
	}{
		{"1", fields{[]byte(`123`)}, 123.0, false},
		{"2", fields{[]byte(`1.23e1`)}, 12.3, false},
		{"3", fields{[]byte(`123.1`)}, 123.1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			json := &JSON{
				data: tt.fields.data,
			}
			got, err := json.ParseFloat()
			if (err != nil) != tt.wantErr {
				t.Errorf("JSON.ParseFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("JSON.ParseFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSON_ParseBoolean(t *testing.T) {
	type fields struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{"1", fields{[]byte(`true`)}, true, false},
		{"2", fields{[]byte(`false`)}, false, false},
		{"3", fields{[]byte(`truess`)}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			json := &JSON{
				data: tt.fields.data,
			}
			got, err := json.ParseBoolean()
			if (err != nil) != tt.wantErr {
				t.Errorf("JSON.ParseBoolean() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("JSON.ParseBoolean() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSON_ParseString(t *testing.T) {
	type fields struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{"1", fields{[]byte(`"test"`)}, "test", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			json := &JSON{
				data: tt.fields.data,
			}
			got, err := json.ParseString()
			if (err != nil) != tt.wantErr {
				t.Errorf("JSON.ParseString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("JSON.ParseString() = %v, want %v", got, tt.want)
			}
		})
	}
}
