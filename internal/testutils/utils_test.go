// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU GPL License version 3.

package testutils

import "testing"

// NB: fields have to be public to be accessible through reflection (Interface method)
type testStruct struct {
	Field1 string
	Field2 testInnerStruct
	Arr    []int
	Arr2   []testInnerStruct
}

type testInnerStruct struct {
	FieldA bool
	FieldB uint
}

func TestEvalExpr(t *testing.T) {
	s := testStruct{
		Field1: "field1value",
		Field2: testInnerStruct{false, 42},
		Arr:    []int{42, 1337},
		Arr2: []testInnerStruct{
			{true, 1},
			{false, 2},
		},
	}
	testEvalExprHelper(t, s, "x.Field1", "field1value")
	testEvalExprHelper(t, s, "x.Field2.FieldB", uint(42))
	testEvalExprHelper(t, s, "x.Arr[1]", 1337)
	testEvalExprHelper(t, s, "x.Arr2[0].FieldA", true)
}

func testEvalExprHelper(t *testing.T, base interface{}, expr string, expected interface{}) {
	t.Helper()
	v, err := evalExpr(expr, base)
	if err != nil {
		t.Fatal(err)
	}
	AssertEq(t, v, expected)
}
