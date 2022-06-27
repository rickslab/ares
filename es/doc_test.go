package es

import (
	"testing"

	"github.com/rickslab/ares/util"
)

func TestGetDoc(t *testing.T) {
	a := struct {
		Id   int64
		Name string `es:"name"`
	}{
		Id:   1,
		Name: "Rick",
	}

	doc, err := GetDoc(a)
	util.AssertErrorT(t, err)
	util.AssertEqualT(t, doc["name"], "Rick")

	b := map[string]interface{}{
		"id":   123,
		"name": "Rick",
	}

	doc, err = GetDoc(b)
	util.AssertErrorT(t, err)
	util.AssertEqualT(t, doc["id"], 123)
	util.AssertEqualT(t, doc["name"], "Rick")
}
