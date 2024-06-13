package utils

func Shift[T any](slice []T, count int) (rest []T, shifted []T) {
	if len(slice) == 0 {
		return []T{}, []T{}
	}
	if count <= 0 || count > len(slice) {
		return []T{}, slice
	}

	shifted = slice[:count]

	rest = slice[count:]

	return rest, shifted
}

func Select[TInput any, TOutput any](slice []TInput, fn func(item TInput) TOutput) []TOutput {
	var result []TOutput
	for _, item := range slice {
		result = append(result, fn(item))
	}

	return result
}
