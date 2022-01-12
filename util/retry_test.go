package util_test

import (
	"errors"
	"testing"
	"time"

	"github.com/rickslab/ares/util"
)

var (
	errFailed = errors.New("failed")
)

func TestRetry(t *testing.T) {
	i := 0
	err := util.Retry(func() error {
		i++
		t.Logf("i=%d\n", i)
		if i == 5 {
			return nil
		}
		return errFailed
	}, 10, time.Millisecond, 0)
	util.AssertErrorT(t, err)

	i = 0
	err = util.Retry(func() error {
		i++
		t.Logf("i=%d\n", i)
		if i == 5 {
			return nil
		}
		return errFailed
	}, 4, time.Millisecond, 0)
	util.AssertEqualT(t, err, errFailed)
}
