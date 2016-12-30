package jzon

import "testing"

func TestParser_numberEnd(t *testing.T) {
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
			p := &Parser{
				data:   tt.fields.data,
				offset: 0,
			}
			if got, _ := p.numberEnd(); got != tt.want {
				t.Errorf("Parser.numberEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}
