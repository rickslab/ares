package util

import (
	"reflect"
	"runtime"
	"testing"
)

func Assert(condition bool, info string) {
	if !condition {
		panic(info)
	}
}

func AssertError(err error) {
	if err != nil {
		panic(err)
	}
}

func AssertEqualT(t *testing.T, x, y any) {
	if !reflect.DeepEqual(x, y) {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("\n%s:%d: %+v not equal to %+v", file, line, x, y)
	}
}

func AssertNotEqualT(t *testing.T, x, y any) {
	if reflect.DeepEqual(x, y) {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("\n%s:%d: %+v equal to %+v", file, line, x, y)
	}
}

func AssertNilT(t *testing.T, x any) {
	if !reflect.ValueOf(x).IsNil() {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("\n%s:%d: %+v is not nil", file, line, x)
	}
}

func AssertNotNilT(t *testing.T, x any) {
	if reflect.ValueOf(x).IsNil() {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("\n%s:%d: %+v is nil", file, line, x)
	}
}

func AssertErrorT(t *testing.T, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("\n%s:%d: %v", file, line, err)
	}
}
