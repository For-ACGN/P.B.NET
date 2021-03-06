// Package ranger implements value validations
//
// Copyright 2014 Roberto Teixeira <robteix@robteix.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ranger_test

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	. "gopkg.in/check.v1"

	"project/external/ranger"
)

func Test(t *testing.T) {
	TestingT(t)
}

type MySuite struct{}

var _ = Suite(&MySuite{})

type Simple struct {
	A int `range:"min=10"`
}

type I interface {
	Foo() string
}

type Impl struct {
	F string `range:"len=3"`
}

func (i *Impl) Foo() string {
	return i.F
}

type Impl2 struct {
	F string `range:"len=3"`
}

func (i Impl2) Foo() string {
	return i.F
}

type TestStruct struct {
	A   int    `range:"nonzero" json:"a"`
	B   string `range:"len=8,min=6,max=4"`
	Sub struct {
		A int `range:"nonzero" json:"sub_a"`
		B string
		C float64 `range:"nonzero,min=1" json:"c_is_a_float"`
		D *string `range:"nonzero"`
	}
	D *Simple `range:"nonzero"`
	E I       `range:"nonzero"`
}

func (ms *MySuite) TestValidate(c *C) {
	t := TestStruct{
		A: 0,
		B: "12345",
	}
	t.Sub.A = 1
	t.Sub.B = ""
	t.Sub.C = 0.0
	t.D = &Simple{10}
	t.E = &Impl{"hello"}

	err := ranger.Validate(t)
	c.Assert(err, NotNil)

	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A"], HasError, ranger.ErrZeroValue)
	c.Assert(errs["B"], HasError, ranger.ErrLen)
	c.Assert(errs["B"], HasError, ranger.ErrMin)
	c.Assert(errs["B"], HasError, ranger.ErrMax)
	c.Assert(errs["Sub.A"], HasLen, 0)
	c.Assert(errs["Sub.B"], HasLen, 0)
	c.Assert(errs["Sub.C"], HasLen, 2)
	c.Assert(errs["Sub.D"], HasError, ranger.ErrZeroValue)
	c.Assert(errs["E.F"], HasError, ranger.ErrLen)
}

func (ms *MySuite) TestValidSlice(c *C) {
	s := make([]int, 0, 10)
	err := ranger.Valid(s, "nonzero")
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, ranger.ErrZeroValue)

	for i := 0; i < 10; i++ {
		s = append(s, i)
	}

	err = ranger.Valid(s, "min=11,max=5,len=9,nonzero")
	c.Assert(err, NotNil)
	errs, ok = err.(ranger.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, ranger.ErrMin)
	c.Assert(errs, HasError, ranger.ErrMax)
	c.Assert(errs, HasError, ranger.ErrLen)
	c.Assert(errs, Not(HasError), ranger.ErrZeroValue)
}

func (ms *MySuite) TestValidMap(c *C) {
	m := make(map[string]string)
	err := ranger.Valid(m, "nonzero")
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, ranger.ErrZeroValue)

	err = ranger.Valid(m, "min=1")
	c.Assert(err, NotNil)
	errs, ok = err.(ranger.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, ranger.ErrMin)

	m = map[string]string{"A": "a", "B": "a"}
	err = ranger.Valid(m, "max=1")
	c.Assert(err, NotNil)
	errs, ok = err.(ranger.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, ranger.ErrMax)

	err = ranger.Valid(m, "min=2, max=5")
	c.Assert(err, IsNil)

	m = map[string]string{
		"1": "a",
		"2": "b",
		"3": "c",
		"4": "d",
		"5": "e",
	}
	err = ranger.Valid(m, "len=4,min=6,max=1,nonzero")
	c.Assert(err, NotNil)
	errs, ok = err.(ranger.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, ranger.ErrLen)
	c.Assert(errs, HasError, ranger.ErrMin)
	c.Assert(errs, HasError, ranger.ErrMax)
	c.Assert(errs, Not(HasError), ranger.ErrZeroValue)

}

func (ms *MySuite) TestValidFloat(c *C) {
	err := ranger.Valid(12.34, "nonzero")
	c.Assert(err, IsNil)

	err = ranger.Valid(0.0, "nonzero")
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, ranger.ErrZeroValue)
}

func (ms *MySuite) TestValidInt(c *C) {
	i := 123
	err := ranger.Valid(i, "nonzero")
	c.Assert(err, IsNil)

	err = ranger.Valid(i, "min=1")
	c.Assert(err, IsNil)

	err = ranger.Valid(i, "min=124, max=122")
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, ranger.ErrMin)
	c.Assert(errs, HasError, ranger.ErrMax)

	err = ranger.Valid(i, "max=10")
	c.Assert(err, NotNil)
	errs, ok = err.(ranger.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, ranger.ErrMax)
}

func (ms *MySuite) TestValidString(c *C) {
	s := "test1234"
	err := ranger.Valid(s, "len=8")
	c.Assert(err, IsNil)

	err = ranger.Valid(s, "len=0")
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasError, ranger.ErrLen)

	err = ranger.Valid(s, "regexp=^[tes]{4}.*")
	c.Assert(err, IsNil)

	err = ranger.Valid(s, "regexp=^.*[0-9]{5}$")
	c.Assert(err, NotNil)

	err = ranger.Valid("", "nonzero,len=3,max=1")
	c.Assert(err, NotNil)
	errs, ok = err.(ranger.ErrorArray)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 2)
	c.Assert(errs, HasError, ranger.ErrZeroValue)
	c.Assert(errs, HasError, ranger.ErrLen)
	c.Assert(errs, Not(HasError), ranger.ErrMax)
}

func (ms *MySuite) TestValidateStructVar(c *C) {
	// just verifies that a the given val is a struct
	err := ranger.SetValidationFunc("struct", func(val interface{}, _ string) error {
		v := reflect.ValueOf(val)
		if v.Kind() == reflect.Struct {
			return nil
		}
		return ranger.ErrUnsupported
	})
	c.Assert(err, IsNil)

	type test struct {
		A int
	}
	err = ranger.Valid(test{}, "struct")
	c.Assert(err, IsNil)

	type test2 struct {
		B int
	}
	type test1 struct {
		A test2 `range:"struct"`
	}

	err = ranger.Validate(test1{})
	c.Assert(err, IsNil)

	type test4 struct {
		B int `range:"foo"`
	}
	type test3 struct {
		A test4
	}
	err = ranger.Validate(test3{})
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A.B"], HasError, ranger.ErrUnknownTag)
}

func (ms *MySuite) TestValidatePointerVar(c *C) {
	// just verifies that a the given val is a struct
	err := ranger.SetValidationFunc("struct", func(val interface{}, _ string) error {
		v := reflect.ValueOf(val)
		if v.Kind() == reflect.Struct {
			return nil
		}
		return ranger.ErrUnsupported
	})
	c.Assert(err, IsNil)

	err = ranger.SetValidationFunc("nil", func(val interface{}, _ string) error {
		v := reflect.ValueOf(val)
		if v.IsNil() {
			return nil
		}
		return ranger.ErrUnsupported
	})
	c.Assert(err, IsNil)

	type test struct {
		A int
	}
	err = ranger.Valid(&test{}, "struct")
	c.Assert(err, IsNil)

	type test2 struct {
		B int
	}
	type test1 struct {
		A *test2 `range:"struct"`
	}

	err = ranger.Validate(&test1{&test2{}})
	c.Assert(err, IsNil)

	type test4 struct {
		B int `range:"foo"`
	}
	type test3 struct {
		A test4
	}
	err = ranger.Validate(&test3{})
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A.B"], HasError, ranger.ErrUnknownTag)

	err = ranger.Valid((*test)(nil), "nil")
	c.Assert(err, IsNil)

	type test5 struct {
		A *test2 `range:"nil"`
	}
	err = ranger.Validate(&test5{})
	c.Assert(err, IsNil)

	type test6 struct {
		A *test2 `range:"nonzero"`
	}
	err = ranger.Validate(&test6{})
	errs, ok = err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A"], HasError, ranger.ErrZeroValue)

	err = ranger.Validate(&test6{&test2{}})
	c.Assert(err, IsNil)

	type test7 struct {
		A *string `range:"min=6"`
		B *int    `range:"len=7"`
		C *int    `range:"min=12"`
		D *int    `range:"nonzero"`
		E *int    `range:"nonzero"`
		F *int    `range:"nonnil"`
		G *int    `range:"nonnil"`
	}
	s := "aaa"
	b := 8
	e := 0
	err = ranger.Validate(&test7{&s, &b, nil, nil, &e, &e, nil})
	errs, ok = err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A"], HasError, ranger.ErrMin)
	c.Assert(errs["B"], HasError, ranger.ErrLen)
	c.Assert(errs["C"], IsNil)
	c.Assert(errs["D"], HasError, ranger.ErrZeroValue)
	c.Assert(errs["E"], HasError, ranger.ErrZeroValue)
	c.Assert(errs["F"], IsNil)
	c.Assert(errs["G"], HasError, ranger.ErrZeroValue)
}

func (ms *MySuite) TestValidateOmittedStructVar(c *C) {
	type test2 struct {
		B int `range:"min=1"`
	}
	type test1 struct {
		A test2 `range:"-"`
	}

	t := test1{}
	err := ranger.Validate(t)
	c.Assert(err, IsNil)

	errs := ranger.Valid(test2{}, "-")
	c.Assert(errs, IsNil)
}

func (ms *MySuite) TestUnknownTag(c *C) {
	type test struct {
		A int `range:"foo"`
	}
	t := test{}
	err := ranger.Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 1)
	c.Assert(errs["A"], HasError, ranger.ErrUnknownTag)
}

func (ms *MySuite) TestValidateStructWithSlice(c *C) {
	type test2 struct {
		Num    int    `range:"max=2"`
		String string `range:"nonzero"`
	}

	type test struct {
		Slices []test2 `range:"len=1"`
	}

	t := test{
		Slices: []test2{{
			Num:    6,
			String: "foo",
		}},
	}
	err := ranger.Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["Slices[0].Num"], HasError, ranger.ErrMax)
	c.Assert(errs["Slices[0].String"], IsNil) // sanity check
}

func (ms *MySuite) TestValidateStructWithNestedSlice(c *C) {
	type test2 struct {
		Num int `range:"max=2"`
	}

	type test struct {
		Slices [][]test2
	}

	t := test{
		Slices: [][]test2{{{Num: 6}}},
	}
	err := ranger.Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["Slices[0][0].Num"], HasError, ranger.ErrMax)
}

func (ms *MySuite) TestValidateStructWithMap(c *C) {
	type test2 struct {
		Num int `range:"max=2"`
	}

	type test struct {
		Map          map[string]test2
		StructKeyMap map[test2]test2
	}

	t := test{
		Map: map[string]test2{
			"hello": {Num: 6},
		},
		StructKeyMap: map[test2]test2{
			{Num: 3}: {Num: 1},
		},
	}
	err := ranger.Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)

	c.Assert(errs["Map[hello](value).Num"], HasError, ranger.ErrMax)
	c.Assert(errs["StructKeyMap[{Num:3}](key).Num"], HasError, ranger.ErrMax)
}

func (ms *MySuite) TestUnsupported(c *C) {
	type test struct {
		A int     `range:"regexp=a.*b"`
		B float64 `range:"regexp=.*"`
	}
	t := test{}
	err := ranger.Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 2)
	c.Assert(errs["A"], HasError, ranger.ErrUnsupported)
	c.Assert(errs["B"], HasError, ranger.ErrUnsupported)
}

func (ms *MySuite) TestBadParameter(c *C) {
	type test struct {
		A string `range:"min="`
		B string `range:"len=="`
		C string `range:"max=foo"`
	}
	t := test{}
	err := ranger.Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 3)
	c.Assert(errs["A"], HasError, ranger.ErrBadParameter)
	c.Assert(errs["B"], HasError, ranger.ErrBadParameter)
	c.Assert(errs["C"], HasError, ranger.ErrBadParameter)
}

func (ms *MySuite) TestCopy(c *C) {
	v := ranger.NewValidator()
	// WithTag calls copy, so we just copy the ranger with the same tag
	v2 := v.WithTag("validate")
	// now we add a custom func only to the second one, it shouldn't get added
	// to the first
	err := v2.SetValidationFunc("custom", func(_ interface{}, _ string) error { return nil })
	c.Assert(err, IsNil)

	type test struct {
		A string `range:"custom"`
	}
	err = v2.Validate(test{})
	c.Assert(err, IsNil)

	err = v.Validate(test{})
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 1)
	c.Assert(errs["A"], HasError, ranger.ErrUnknownTag)
}

func (ms *MySuite) TestTagEscape(c *C) {
	type test struct {
		A string `range:"min=0,regexp=^a{3\\,10}"`
	}
	t := test{"aaaa"}
	err := ranger.Validate(t)
	c.Assert(err, IsNil)

	t2 := test{"aa"}
	err = ranger.Validate(t2)
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A"], HasError, ranger.ErrRegexp)
}

func (ms *MySuite) TestEmbeddedFields(c *C) {
	type baseTest struct {
		A string `range:"min=1"`
	}
	type test struct {
		baseTest
		B string `range:"min=1"`
	}

	err := ranger.Validate(test{})
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 2)
	c.Assert(errs["baseTest.A"], HasError, ranger.ErrMin)
	c.Assert(errs["B"], HasError, ranger.ErrMin)

	type test2 struct {
		baseTest `range:"-"`
	}
	err = ranger.Validate(test2{})
	c.Assert(err, IsNil)
}

func (ms *MySuite) TestEmbeddedPointerFields(c *C) {
	type baseTest struct {
		A string `range:"min=1"`
	}
	type test struct {
		*baseTest
		B string `range:"min=1"`
	}

	err := ranger.Validate(test{baseTest: &baseTest{}})
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 2)
	c.Assert(errs["baseTest.A"], HasError, ranger.ErrMin)
	c.Assert(errs["B"], HasError, ranger.ErrMin)
}

func (ms *MySuite) TestEmbeddedNilPointerFields(c *C) {
	type baseTest struct {
		A string `range:"min=1"`
	}
	type test struct {
		*baseTest
	}

	err := ranger.Validate(test{})
	c.Assert(err, IsNil)
}

func (ms *MySuite) TestPrivateFields(c *C) {
	type test struct {
		b string `range:"min=2"`
	}
	t := test{
		b: "1",
	}
	err := ranger.Validate(t)
	c.Assert(err, IsNil)
}

func (ms *MySuite) TestEmbeddedUnexported(c *C) {
	type baseTest struct {
		A string `range:"min=1"`
	}
	type test struct {
		baseTest `range:"nonnil"`
	}

	err := ranger.Validate(test{})
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 2)
	c.Assert(errs["baseTest"], HasError, ranger.ErrCannotValidate)
	c.Assert(errs["baseTest.A"], HasError, ranger.ErrMin)
}

func (ms *MySuite) TestValidateStructWithByteSliceSlice(c *C) {
	type test struct {
		Slices [][]byte `range:"len=1"`
	}

	t := test{
		Slices: [][]byte{[]byte(``)},
	}
	err := ranger.Validate(t)
	c.Assert(err, IsNil)
}

func (ms *MySuite) TestEmbeddedInterface(c *C) {
	type test struct {
		I
	}

	err := ranger.Validate(test{Impl2{"hello"}})
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 1)
	c.Assert(errs["I.F"], HasError, ranger.ErrLen)

	err = ranger.Validate(test{&Impl{"hello"}})
	c.Assert(err, NotNil)
	errs, ok = err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 1)
	c.Assert(errs["I.F"], HasError, ranger.ErrLen)

	err = ranger.Validate(test{})
	c.Assert(err, IsNil)

	type test2 struct {
		I `range:"nonnil"`
	}
	err = ranger.Validate(test2{})
	c.Assert(err, NotNil)
	errs, ok = err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs, HasLen, 1)
	c.Assert(errs["I"], HasError, ranger.ErrZeroValue)
}

func (ms *MySuite) TestErrors(c *C) {
	err := ranger.ErrorMap{
		"foo": ranger.ErrorArray{
			fmt.Errorf("bar"),
		},
		"baz": ranger.ErrorArray{
			fmt.Errorf("qux"),
		},
	}
	sep := ", "
	expected := "foo: bar, baz: qux"

	expectedParts := strings.Split(expected, sep)
	sort.Strings(expectedParts)

	errString := err.Error()
	errStringParts := strings.Split(errString, sep)
	sort.Strings(errStringParts)

	c.Assert(expectedParts, DeepEquals, errStringParts)
}

func (ms *MySuite) TestJSONPrint(c *C) {
	t := TestStruct{
		A: 0,
	}
	err := ranger.WithPrintJSON(true).Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A"], IsNil)
	c.Assert(errs["a"], HasError, ranger.ErrZeroValue)
}

func (ms *MySuite) TestJSONPrintOff(c *C) {
	t := TestStruct{
		A: 0,
	}
	err := ranger.WithPrintJSON(false).Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["A"], HasError, ranger.ErrZeroValue)
	c.Assert(errs["a"], IsNil)
}

func (ms *MySuite) TestJSONPrintNoTag(c *C) {
	t := TestStruct{
		B: "te",
	}
	err := ranger.WithPrintJSON(true).Validate(t)
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["B"], HasError, ranger.ErrLen)
}

func (ms *MySuite) TestValidateSlice(c *C) {
	type test2 struct {
		Num    int    `range:"max=2"`
		String string `range:"nonzero"`
	}

	err := ranger.Validate([]test2{
		{
			Num:    6,
			String: "foo",
		},
		{
			Num:    1,
			String: "foo",
		},
	})
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["[0].Num"], HasError, ranger.ErrMax)
	c.Assert(errs["[0].String"], IsNil) // sanity check
	c.Assert(errs["[1].Num"], IsNil)    // sanity check
	c.Assert(errs["[1].String"], IsNil) // sanity check
}

func (ms *MySuite) TestValidateMap(c *C) {
	type test2 struct {
		Num    int    `range:"max=2"`
		String string `range:"nonzero"`
	}

	err := ranger.Validate(map[string]test2{
		"first": {
			Num:    6,
			String: "foo",
		},
		"second": {
			Num:    1,
			String: "foo",
		},
	})
	c.Assert(err, NotNil)
	errs, ok := err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["[first](value).Num"], HasError, ranger.ErrMax)
	c.Assert(errs["[first](value).String"], IsNil)  // sanity check
	c.Assert(errs["[second](value).Num"], IsNil)    // sanity check
	c.Assert(errs["[second](value).String"], IsNil) // sanity check

	err = ranger.Validate(map[test2]string{
		{
			Num:    6,
			String: "foo",
		}: "first",
		{
			Num:    1,
			String: "foo",
		}: "second",
	})
	c.Assert(err, NotNil)
	errs, ok = err.(ranger.ErrorMap)
	c.Assert(ok, Equals, true)
	c.Assert(errs["[{Num:6 String:foo}](key).Num"], HasError, ranger.ErrMax)
	c.Assert(errs["[{Num:6 String:foo}](key).String"], IsNil) // sanity check
	c.Assert(errs["[{Num:1 String:foo}](key).Num"], IsNil)    // sanity check
	c.Assert(errs["[{Num:1 String:foo}](key).String"], IsNil) // sanity check
}

type hasErrorChecker struct {
	*CheckerInfo
}

func (c *hasErrorChecker) Check(params []interface{}, _ []string) (bool, string) {
	var (
		ok    bool
		slice []error
		value error
	)
	slice, ok = params[0].(ranger.ErrorArray)
	if !ok {
		return false, "First parameter is not an ErrorArray"
	}
	value, ok = params[1].(error)
	if !ok {
		return false, "Second parameter is not an error"
	}

	for _, v := range slice {
		if v == value {
			return true, ""
		}
	}
	return false, ""
}

func (c *hasErrorChecker) Info() *CheckerInfo {
	return c.CheckerInfo
}

var HasError = &hasErrorChecker{
	&CheckerInfo{
		Name:   "HasError",
		Params: []string{"HasError", "expected to contain"},
	},
}

// padding functions
func TestSet(t *testing.T) {
	ranger.SetTag("validate")
	ranger.SetTag("range")
	ranger.SetPrintJSON(false)
}
