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

func Except[TInput any](slice []TInput, fn func(item TInput) bool) []TInput {
	var result []TInput
	for _, item := range slice {
		if fn(item) {
			continue
		}
		result = append(result, item)
	}

	return result
}
