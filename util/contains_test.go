package util

import "testing"

func TestArrayContains(t *testing.T) {
	arr := []int64{1, 3, 2, 4}
	ok := ArrayContains(arr, 2)
	AssertEqualT(t, ok, true)

	dict := map[string]any{
		"Id":   123,
		"Name": "Rick",
	}
	ok = MapContainsKey(dict, "Id")
	AssertEqualT(t, ok, true)
}
