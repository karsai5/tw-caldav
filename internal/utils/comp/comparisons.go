package comp

import "time"

func EqualPtrs[T comparable](a, b *T) bool {
	if a != nil && b != nil {
		return *a == *b
	}
	return false
}

func EqualTimePtrs(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}
