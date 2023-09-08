package es

import (
	"testing"

	"github.com/rickslab/ares/util"
)

func TestGetObject(t *testing.T) {
	a := struct {
		Id   int64
		Name string `es:"name"`
	}{
		Id:   1,
		Name: "Rick",
	}

	obj, err := GetObject(a)
	util.AssertErrorT(t, err)
	util.AssertEqualT(t, obj["name"], "Rick")

	b := map[string]any{
		"id":   123,
		"name": "Rick",
	}

	obj, err = GetObject(b)
	util.AssertErrorT(t, err)
	util.AssertEqualT(t, obj["id"], 123)
	util.AssertEqualT(t, obj["name"], "Rick")
}
