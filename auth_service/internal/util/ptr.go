package util

func GenericPtr[T any](v T) *T {
	return &v
}
