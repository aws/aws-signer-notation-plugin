//  Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License"). You may
//  not use this file except in compliance with the License. A copy of the
//  License is located at
//
// 	http://aws.amazon.com/apache2.0
//
//  or in the "license" file accompanying this file. This file is distributed
//  on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language governing
//  permissions and limitations under the License.

// Package slices provides utility methods for slice.
package slices

// AppendIfNotPresent appends t to s if t is not already present.
func AppendIfNotPresent[T comparable](s []T, t T) []T {
	if !Contains(s, t) {
		return append(s, t)
	}
	return s
}

// Contains reports whether v is present in s.
func Contains[E comparable](s []E, v E) bool {
	for _, vs := range s {
		if v == vs {
			return true
		}
	}
	return false
}
