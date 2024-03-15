package slices

import (
	"reflect"
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		s    []string
		v    string
		want bool
	}{
		{nil, "", false},
		{[]string{}, "", false},
		{[]string{"1", "2", "3"}, "2", true},
		{[]string{"1", "2", "2", "3"}, "2", true},
		{[]string{"1", "2", "3", "2"}, "2", true},
	}
	for _, tt := range tests {
		if got := Contains(tt.s, tt.v); got != tt.want {
			t.Errorf("index() = %v, want %v", got, tt.want)
		}
	}
}

func TestAppendIfNotPresent(t *testing.T) {
	tests := []struct {
		s    []string
		v    string
		want []string
	}{
		{[]string{"1", "2"}, "3", []string{"1", "2", "3"}},
		{[]string{"1", "2", "3"}, "3", []string{"1", "2", "3"}},
	}
	for _, tt := range tests {
		if got := AppendIfNotPresent(tt.s, tt.v); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("AppendIfNotPresent() = %v, want %v", got, tt.want)
		}
	}
}
