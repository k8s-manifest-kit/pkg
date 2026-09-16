package util

// ToAnySlice converts a typed slice to a slice of any values.
func ToAnySlice[T any](values []T) []any {
	result := make([]any, len(values))
	for i, value := range values {
		result[i] = value
	}

	return result
}
