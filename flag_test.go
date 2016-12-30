package jzon

import "testing"

func Test_flag_addFlag(t *testing.T) {
	type args struct {
		mask []flag
	}
	tests := []struct {
		name string
		f    *flag
		args args
		want *flag
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.add(tt.args.mask...); got != tt.want {
				t.Errorf("flag.addFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_flag_cleanFlag(t *testing.T) {
	type args struct {
		mask []flag
	}
	tests := []struct {
		name string
		f    *flag
		args args
		want *flag
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.remove(tt.args.mask...); got != tt.want {
				t.Errorf("flag.cleanFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}
