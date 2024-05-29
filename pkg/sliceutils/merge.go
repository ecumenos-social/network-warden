package sliceutils

func Merge[T interface{}](slices ...[]T) []T {
	out := make([]T, 0, len(slices))
	for _, sl := range slices {
		out = append(out, sl...)
	}

	return out
}
