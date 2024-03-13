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
