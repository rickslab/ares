package util

func ArrayContains[T int64 | int32 | int | string](arr []T, dst T) bool {
	for _, obj := range arr {
		if obj == dst {
			return true
		}
	}
	return false
}

func MapContainsKey[K string | int64, V any | bool](dict map[K]V, key K) bool {
	for k := range dict {
		if k == key {
			return true
		}
	}
	return false
}
