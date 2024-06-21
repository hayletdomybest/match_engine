package utils

import "reflect"

func MapContains[TKey any](m any, k TKey) bool {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		panic("should use map as params 0")
	}

	key := reflect.ValueOf(k)
	return v.MapIndex(key).IsValid()
}

func SliceToMap[T any, TKey comparable](slice []T, fn func(T) TKey) map[TKey]T {
	m := make(map[TKey]T)

	for _, v := range slice {
		key := fn(v)
		m[key] = v
	}

	return m
}

func MapToSlice[T any, TKey comparable, TOut any](m map[TKey]T, fn func(TKey, T) TOut) []TOut {
	var res []TOut
	for k, v := range m {
		res = append(res, fn(k, v))
	}
	return res
}
