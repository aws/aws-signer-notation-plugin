// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
// http://aws.amazon.com/apache2.0
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

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
