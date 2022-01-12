package util

import (
	"testing"
	"time"
)

type test1 struct {
	A int64
	B string
	C time.Time `xcopy:"Unix"`
	D string    `xcopy:"-"`
}

type test2 struct {
	A int64
	B string
	C int64
}

func TestXcopyStruce(t *testing.T) {
	t1 := test1{
		A: 100,
		B: "Hello",
		C: time.Unix(10000, 0),
		D: "World",
	}
	var t2 *test2
	Xcopy(&t2, t1)

	AssertEqualT(t, t2.A, int64(100))
	AssertEqualT(t, t2.B, "Hello")
	AssertEqualT(t, t2.C, int64(10000))

	t2 = Xcopy(&test2{}, &t1).(*test2)

	AssertEqualT(t, t2.A, int64(100))
	AssertEqualT(t, t2.B, "Hello")
	AssertEqualT(t, t2.C, int64(10000))
}

func TestXcopySlice(t *testing.T) {
	t1 := []test1{
		{
			A: 100,
			B: "Hello",
			C: time.Unix(10000, 0),
			D: "World",
		},
		{
			A: 200,
			B: "Hello2",
			C: time.Unix(20000, 0),
			D: "World2",
		},
	}
	t2 := []test2{
		{
			A: 0,
			B: "Test",
			C: 100,
		},
	}
	Xcopy(&t2, t1)

	AssertEqualT(t, len(t2), 2)
	AssertEqualT(t, t2[0].A, int64(100))
	AssertEqualT(t, t2[0].B, "Hello")
	AssertEqualT(t, t2[0].C, int64(10000))
	AssertEqualT(t, t2[1].A, int64(200))
	AssertEqualT(t, t2[1].B, "Hello2")
	AssertEqualT(t, t2[1].C, int64(20000))

	t3 := []*test2{
		{
			A: 0,
			B: "Test",
			C: 100,
		},
	}
	Xcopy(&t3, &t1)

	AssertEqualT(t, len(t2), 2)
	AssertEqualT(t, t3[0].A, int64(100))
	AssertEqualT(t, t3[0].B, "Hello")
	AssertEqualT(t, t3[0].C, int64(10000))
	AssertEqualT(t, t3[1].A, int64(200))
	AssertEqualT(t, t3[1].B, "Hello2")
	AssertEqualT(t, t3[1].C, int64(20000))
}

func TestXcopySlicePtr(t *testing.T) {
	t1 := []*test1{
		{
			A: 100,
			B: "Hello",
			C: time.Unix(10000, 0),
			D: "World",
		},
		{
			A: 200,
			B: "Hello2",
			C: time.Unix(20000, 0),
			D: "World2",
		},
	}
	t2 := []*test2{
		{
			A: 0,
			B: "Test",
			C: 100,
		},
	}
	Xcopy(&t2, t1)

	AssertEqualT(t, len(t2), 2)
	AssertEqualT(t, t2[0].A, int64(100))
	AssertEqualT(t, t2[0].B, "Hello")
	AssertEqualT(t, t2[0].C, int64(10000))
	AssertEqualT(t, t2[1].A, int64(200))
	AssertEqualT(t, t2[1].B, "Hello2")
	AssertEqualT(t, t2[1].C, int64(20000))

	t3 := []*test2{
		{
			A: 0,
			B: "Test",
			C: 100,
		},
	}
	Xcopy(&t3, &t1)

	AssertEqualT(t, len(t2), 2)
	AssertEqualT(t, t3[0].A, int64(100))
	AssertEqualT(t, t3[0].B, "Hello")
	AssertEqualT(t, t3[0].C, int64(10000))
	AssertEqualT(t, t3[1].A, int64(200))
	AssertEqualT(t, t3[1].B, "Hello2")
	AssertEqualT(t, t3[1].C, int64(20000))
}

type test3 struct {
	A int64
}

func (t3 *test3) Xcopy(t2 *test2) {
	t2.A = t3.A
	t2.B = "Golang"
	t2.C = 1001
}

func TestXcopyMethod(t *testing.T) {
	t3 := &test3{
		A: 300,
	}

	t2 := &test2{}
	Xcopy(t2, t3)

	AssertEqualT(t, t2.A, int64(300))
	AssertEqualT(t, t2.B, "Golang")
	AssertEqualT(t, t2.C, int64(1001))
}
