package crypto

import (
	"testing"

	"github.com/rickslab/ares/util"
)

func TestAesCBCDecrypt(t *testing.T) {
	cases := []string{"", "Test", "Hello", "汪顺夺得男子200米混合泳金牌"}

	for _, c := range cases {
		key := util.RandomString(32)

		data, err := AesCBCEncrypt([]byte(c), []byte(key))
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		rawData, err := AesCBCDecrypt(data, []byte(key))
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if string(rawData) != c {
			t.Fail()
		}
	}
}
