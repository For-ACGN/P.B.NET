package vm

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"project/external/anko/env"
	"project/external/anko/parser"
)

type (
	testStruct1 struct {
		aInterface interface{}
		aBool      bool
		aInt32     int32
		aInt64     int64
		aFloat32   float32
		aFloat64   float32
		aString    string
		aFunc      func()

		aPtrInterface      *interface{}
		aPtrBool           *bool
		aPtrInt32          *int32
		aPtrInt64          *int64
		aPtrFloat32        *float32
		aPtrFloat64        *float32
		aPtrString         *string
		aPtrSliceInterface *[]interface{}
		aPtrSliceBool      *[]bool
		aPtrSliceInt32     *[]int32
		aPtrSliceInt64     *[]int64
		aPtrSliceFloat32   *[]float32
		aPtrSliceFloat64   *[]float32
		aPtrSliceString    *[]string

		aSliceInterface    []interface{}
		aSliceBool         []bool
		aSliceInt32        []int32
		aSliceInt64        []int64
		aSliceFloat32      []float32
		aSliceFloat64      []float32
		aSliceString       []string
		aSlicePtrInterface []*interface{}
		aSlicePtrBool      []*bool
		aSlicePtrInt32     []*int32
		aSlicePtrInt64     []*int64
		aSlicePtrFloat32   []*float32
		aSlicePtrFloat64   []*float32
		aSlicePtrString    []*string

		aMapInterface    map[string]interface{}
		aMapBool         map[string]bool
		aMapInt32        map[string]int32
		aMapInt64        map[string]int64
		aMapFloat32      map[string]float32
		aMapFloat64      map[string]float32
		aMapString       map[string]string
		aMapPtrInterface map[string]*interface{}
		aMapPtrBool      map[string]*bool
		aMapPtrInt32     map[string]*int32
		aMapPtrInt64     map[string]*int64
		aMapPtrFloat32   map[string]*float32
		aMapPtrFloat64   map[string]*float32
		aMapPtrString    map[string]*string

		aChanInterface    chan interface{}
		aChanBool         chan bool
		aChanInt32        chan int32
		aChanInt64        chan int64
		aChanFloat32      chan float32
		aChanFloat64      chan float32
		aChanString       chan string
		aChanPtrInterface chan *interface{}
		aChanPtrBool      chan *bool
		aChanPtrInt32     chan *int32
		aChanPtrInt64     chan *int64
		aChanPtrFloat32   chan *float32
		aChanPtrFloat64   chan *float32
		aChanPtrString    chan *string

		aPtrStruct *testStruct1
	}
	testStruct2 struct {
		aStruct testStruct1
	}
)

var (
	testVarValue    = reflect.Value{}
	testVarBool     = true
	testVarBoolP    = &testVarBool
	testVarInt32    = int32(1)
	testVarInt32P   = &testVarInt32
	testVarInt64    = int64(1)
	testVarInt64P   = &testVarInt64
	testVarFloat32  = float32(1)
	testVarFloat32P = &testVarFloat32
	testVarFloat64  = float64(1)
	testVarFloat64P = &testVarFloat64
	testVarString   = "a"
	testVarStringP  = &testVarString
	testVarFunc     = func() int64 { return 1 }
	testVarFuncP    = &testVarFunc

	testVarValueBool    = reflect.ValueOf(true)
	testVarValueInt32   = reflect.ValueOf(int32(1))
	testVarValueInt64   = reflect.ValueOf(int64(1))
	testVarValueFloat32 = reflect.ValueOf(float32(1.1))
	testVarValueFloat64 = reflect.ValueOf(1.1)
	testVarValueString  = reflect.ValueOf("a")

	testSliceEmpty []interface{}
	testSlice      = []interface{}{nil, true, int64(1), 1.1, "a"}
	testMapEmpty   map[interface{}]interface{}
	testMap        = map[interface{}]interface{}{"a": nil, "b": true, "c": int64(1), "d": 1.1, "e": "e"}
)

// Test is utility struct to make tests easy.
type Test struct {
	Script         string
	ParseError     error
	ParseErrorFunc *func(*testing.T, error)
	EnvSetupFunc   *func(*testing.T, *env.Env)
	Types          map[string]interface{}
	Input          map[string]interface{}
	RunError       error
	RunErrorFunc   *func(*testing.T, error)
	RunOutput      interface{}
	Output         map[string]interface{}
}

// TestOptions is utility struct to pass options to the test.
type TestOptions struct {
	EnvSetupFunc *func(*testing.T, *env.Env)
	Timeout      time.Duration
}

func runTests(t *testing.T, tests []Test, testOptions *TestOptions, options *Options) {
	for _, test := range tests {
		runTest(t, test, testOptions, options)
	}
}

// nolint: gocyclo
//gocyclo:ignore
func runTest(t *testing.T, test Test, testOptions *TestOptions, options *Options) {
	timeout := 60 * time.Second

	// parser.EnableErrorVerbose()
	// parser.EnableDebug(8)

	stmt, err := parser.ParseSrc(test.Script)
	if test.ParseErrorFunc != nil {
		(*test.ParseErrorFunc)(t, err)
	} else if err != nil && test.ParseError != nil {
		if err.Error() != test.ParseError.Error() {
			const format = "ParseSrc error - received: %v - expected: %v - script: %v"
			t.Errorf(format, err, test.ParseError, test.Script)
			return
		}
	} else if err != test.ParseError {
		const format = "ParseSrc error - received: %v - expected: %v - script: %v"
		t.Errorf(format, err, test.ParseError, test.Script)
		return
	}
	// Note: Still want to run the code even after a parse error to see what happens

	envTest := env.NewEnv()
	if testOptions != nil {
		if testOptions.EnvSetupFunc != nil {
			(*testOptions.EnvSetupFunc)(t, envTest)
		}
		if testOptions.Timeout != 0 {
			timeout = testOptions.Timeout
		}
	}
	if test.EnvSetupFunc != nil {
		(*test.EnvSetupFunc)(t, envTest)
	}

	for typeName, typeValue := range test.Types {
		err = envTest.DefineType(typeName, typeValue)
		if err != nil {
			t.Errorf("DefineType error: %v - typeName: %v - script: %v", err, typeName, test.Script)
			return
		}
	}

	for inputName, inputValue := range test.Input {
		err = envTest.Define(inputName, inputValue)
		if err != nil {
			t.Errorf("Define error: %v - inputName: %v - script: %v", err, inputName, test.Script)
			return
		}
	}

	var value interface{}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	value, err = RunContext(ctx, envTest, options, stmt)
	cancel()
	if test.RunErrorFunc != nil {
		(*test.RunErrorFunc)(t, err)
	} else if err != nil && test.RunError != nil {
		if err.Error() != test.RunError.Error() {
			t.Errorf("Run error - received: %v - expected: %v - script: %v", err, test.RunError, test.Script)
			return
		}
	} else if err != test.RunError {
		t.Errorf("Run error - received: %v - expected: %v - script: %v", err, test.RunError, test.Script)
		return
	}

	if !valueEqual(value, test.RunOutput) {
		t.Errorf("Run output - received: %#v - expected: %#v - script: %v", value, test.RunOutput, test.Script)
		t.Errorf("received type: %T - expected: %T", value, test.RunOutput)
		return
	}

	for outputName, outputValue := range test.Output {
		value, err = envTest.Get(outputName)
		if err != nil {
			t.Errorf("Get error: %v - outputName: %v - script: %v", err, outputName, test.Script)
			return
		}

		if !valueEqual(value, outputValue) {
			const format = "outputName %v - received: %#v - expected: %#v - script: %v"
			t.Errorf(format, outputName, value, outputValue, test.Script)
			t.Errorf("received type: %T - expected: %T", value, outputValue)
			continue
		}
	}
}

// valueEqual return true if v1 and v2 is same value. If passed function, does
// extra checks otherwise just doing reflect.DeepEqual.
func valueEqual(v1 interface{}, v2 interface{}) bool {
	v1RV := reflect.ValueOf(v1)
	switch v1RV.Kind() {
	case reflect.Func:
		// This is best effort to check if functions match, but it could be wrong
		v2RV := reflect.ValueOf(v2)
		if !v1RV.IsValid() || !v2RV.IsValid() {
			if v1RV.IsValid() != !v2RV.IsValid() {
				return false
			}
			return true
		} else if v1RV.Kind() != v2RV.Kind() {
			return false
		} else if v1RV.Type() != v2RV.Type() {
			return false
		} else if v1RV.Pointer() != v2RV.Pointer() {
			// From reflect: If v's Kind is Func, the returned pointer is an underlying
			// code pointer, but not necessarily enough to identify a single function uniquely.
			return false
		}
		return true
	}
	switch value1 := v1.(type) {
	case error:
		switch value2 := v2.(type) {
		case error:
			return value1.Error() == value2.Error()
		}
	}

	return reflect.DeepEqual(v1, v2)
}

func TestNumbers(t *testing.T) {
	tests := []Test{
		{Script: ``},
		{Script: `;`},
		{Script: `
`},
		{Script: `
1
`, RunOutput: int64(1)},

		{Script: `1..1`, ParseError: fmt.Errorf("invalid number: 1..1")},
		{Script: `1e.1`, ParseError: fmt.Errorf("invalid number: 1e.1")},
		{Script: `1ee1`, ParseError: fmt.Errorf("syntax error")},
		{Script: `1e+e1`, ParseError: fmt.Errorf("syntax error")},
		{Script: `0x1g`, ParseError: fmt.Errorf("syntax error")},
		{Script: `9223372036854775808`, ParseError: fmt.Errorf("invalid number: 9223372036854775808")},
		{Script: `-9223372036854775809`, ParseError: fmt.Errorf("invalid number: -9223372036854775809")},

		{Script: `1`, RunOutput: int64(1)},
		{Script: `-1`, RunOutput: int64(-1)},
		{Script: `9223372036854775807`, RunOutput: int64(9223372036854775807)},
		{Script: `-9223372036854775808`, RunOutput: int64(-9223372036854775808)},
		{Script: `-9223372036854775807-1`, RunOutput: int64(-9223372036854775808)},
		{Script: `-9223372036854775807 -1`, RunOutput: int64(-9223372036854775808)},
		{Script: `-9223372036854775807 - 1`, RunOutput: int64(-9223372036854775808)},
		{Script: `1.1`, RunOutput: 1.1},
		{Script: `-1.1`, RunOutput: -1.1},

		{Script: `1e1`, RunOutput: float64(10)},
		{Script: `1.5e1`, RunOutput: float64(15)},
		{Script: `1e-1`, RunOutput: 0.1},

		{Script: `-1e1`, RunOutput: float64(-10)},
		{Script: `-1.5e1`, RunOutput: float64(-15)},
		{Script: `-1e-1`, RunOutput: -0.1},

		{Script: `0x1`, RunOutput: int64(1)},
		{Script: `0xa`, RunOutput: int64(10)},
		{Script: `0xb`, RunOutput: int64(11)},
		{Script: `0xc`, RunOutput: int64(12)},
		{Script: `0xe`, RunOutput: int64(14)},
		{Script: `0xf`, RunOutput: int64(15)},
		{Script: `0Xf`, RunOutput: int64(15)},
		{Script: `0XF`, RunOutput: int64(15)},
		{Script: `0x7FFFFFFFFFFFFFFF`, RunOutput: int64(9223372036854775807)},

		{Script: `-0x1`, RunOutput: int64(-1)},
		{Script: `-0xc`, RunOutput: int64(-12)},
		{Script: `-0xe`, RunOutput: int64(-14)},
		{Script: `-0xf`, RunOutput: int64(-15)},
		{Script: `-0Xf`, RunOutput: int64(-15)},
		{Script: `-0x7FFFFFFFFFFFFFFF`, RunOutput: int64(-9223372036854775807)},
	}
	runTests(t, tests, nil, &Options{Debug: true})
}

func TestStrings(t *testing.T) {
	tests := []Test{
		{Script: `a`, Input: map[string]interface{}{"a": 'a'}, RunOutput: 'a', Output: map[string]interface{}{"a": 'a'}},
		{Script: `a.b`, Input: map[string]interface{}{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support member operation"), Output: map[string]interface{}{"a": 'a'}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support index operation"), RunOutput: nil, Output: map[string]interface{}{"a": 'a'}},
		{Script: `a[0:1]`, Input: map[string]interface{}{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support slice operation"), RunOutput: nil, Output: map[string]interface{}{"a": 'a'}},

		{Script: `a.b = "a"`, Input: map[string]interface{}{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support member operation"), RunOutput: nil, Output: map[string]interface{}{"a": 'a'}},
		{Script: `a[0] = "a"`, Input: map[string]interface{}{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support index operation"), RunOutput: nil, Output: map[string]interface{}{"a": 'a'}},
		{Script: `a[0:1] = "a"`, Input: map[string]interface{}{"a": 'a'}, RunError: fmt.Errorf("type int32 does not support slice operation"), RunOutput: nil, Output: map[string]interface{}{"a": 'a'}},

		{Script: `a.b = "a"`, Input: map[string]interface{}{"a": "test"}, RunError: fmt.Errorf("type string does not support member operation"), Output: map[string]interface{}{"a": "test"}},
		{Script: `a[0:1] = "a"`, Input: map[string]interface{}{"a": "test"}, RunError: fmt.Errorf("type string does not support slice operation for assignment"), Output: map[string]interface{}{"a": "test"}},

		{Script: `a`, Input: map[string]interface{}{"a": "test"}, RunOutput: "test", Output: map[string]interface{}{"a": "test"}},
		{Script: `a["a"]`, Input: map[string]interface{}{"a": "test"}, RunError: fmt.Errorf("index must be a number"), Output: map[string]interface{}{"a": "test"}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": ""}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": ""}},
		{Script: `a[-1]`, Input: map[string]interface{}{"a": "test"}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": "test"}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": "test"}, RunOutput: "t", Output: map[string]interface{}{"a": "test"}},
		{Script: `a[1]`, Input: map[string]interface{}{"a": "test"}, RunOutput: "e", Output: map[string]interface{}{"a": "test"}},
		{Script: `a[3]`, Input: map[string]interface{}{"a": "test"}, RunOutput: "t", Output: map[string]interface{}{"a": "test"}},
		{Script: `a[4]`, Input: map[string]interface{}{"a": "test"}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": "test"}},

		{Script: `a`, Input: map[string]interface{}{"a": `"a"`}, RunOutput: `"a"`, Output: map[string]interface{}{"a": `"a"`}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": `"a"`}, RunOutput: `"`, Output: map[string]interface{}{"a": `"a"`}},
		{Script: `a[1]`, Input: map[string]interface{}{"a": `"a"`}, RunOutput: "a", Output: map[string]interface{}{"a": `"a"`}},

		{Script: `a = "\"a\""`, RunOutput: `"a"`, Output: map[string]interface{}{"a": `"a"`}},
		{Script: `a = "\"a\""; a`, RunOutput: `"a"`, Output: map[string]interface{}{"a": `"a"`}},
		{Script: `a = "\"a\""; a[0]`, RunOutput: `"`, Output: map[string]interface{}{"a": `"a"`}},
		{Script: `a = "\"a\""; a[1]`, RunOutput: "a", Output: map[string]interface{}{"a": `"a"`}},

		{Script: `a`, Input: map[string]interface{}{"a": "a\\b"}, RunOutput: "a\\b", Output: map[string]interface{}{"a": "a\\b"}},
		{Script: `a`, Input: map[string]interface{}{"a": "a\\\\b"}, RunOutput: "a\\\\b", Output: map[string]interface{}{"a": "a\\\\b"}},
		{Script: `a = "a\b"`, RunOutput: "a\b", Output: map[string]interface{}{"a": "a\b"}},
		{Script: `a = "a\\b"`, RunOutput: "a\\b", Output: map[string]interface{}{"a": "a\\b"}},

		{Script: `a[:]`, Input: map[string]interface{}{"a": "test data"}, ParseError: fmt.Errorf("syntax error"), Output: map[string]interface{}{"a": "test data"}},

		{Script: `a[0:]`, Input: map[string]interface{}{"a": ""}, RunOutput: "", Output: map[string]interface{}{"a": ""}},
		{Script: `a[1:]`, Input: map[string]interface{}{"a": ""}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": ""}},
		{Script: `a[:0]`, Input: map[string]interface{}{"a": ""}, RunOutput: "", Output: map[string]interface{}{"a": ""}},
		{Script: `a[:1]`, Input: map[string]interface{}{"a": ""}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": ""}},
		{Script: `a[0:0]`, Input: map[string]interface{}{"a": ""}, RunOutput: "", Output: map[string]interface{}{"a": ""}},

		{Script: `a[1:0]`, Input: map[string]interface{}{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[-1:2]`, Input: map[string]interface{}{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[1:-2]`, Input: map[string]interface{}{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[-1:]`, Input: map[string]interface{}{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[:-2]`, Input: map[string]interface{}{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": "test data"}},

		{Script: `a[0:0]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[0:1]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "t", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[0:2]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "te", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[0:3]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "tes", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[0:7]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "test da", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[0:8]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "test dat", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[0:9]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "test data", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[0:10]`, Input: map[string]interface{}{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": "test data"}},

		{Script: `a[1:1]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[1:2]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "e", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[1:3]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "es", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[1:7]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "est da", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[1:8]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "est dat", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[1:9]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "est data", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[1:10]`, Input: map[string]interface{}{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": "test data"}},

		{Script: `a[0:9]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "test data", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[1:9]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "est data", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[2:9]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "st data", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[3:9]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "t data", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[7:9]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "ta", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[8:9]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "a", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[9:9]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "", Output: map[string]interface{}{"a": "test data"}},

		{Script: `a[:0]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[:1]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "t", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[:2]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "te", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[:3]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "tes", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[:7]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "test da", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[:8]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "test dat", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[:9]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "test data", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[:10]`, Input: map[string]interface{}{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": "test data"}},

		{Script: `a[0:]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "test data", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[1:]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "est data", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[2:]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "st data", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[3:]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "t data", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[7:]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "ta", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[8:]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "a", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[9:]`, Input: map[string]interface{}{"a": "test data"}, RunOutput: "", Output: map[string]interface{}{"a": "test data"}},
		{Script: `a[10:]`, Input: map[string]interface{}{"a": "test data"}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": "test data"}},

		// index assignment - len 0
		{Script: `a = ""; a[0] = "x"`, RunOutput: "x", Output: map[string]interface{}{"a": "x"}},
		{Script: `a = ""; a[1] = "x"`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": ""}},

		// index assignment - len 1
		{Script: `a = "a"; a[0] = "x"`, RunOutput: "x", Output: map[string]interface{}{"a": "x"}},
		{Script: `a = "a"; a[1] = "x"`, RunOutput: "x", Output: map[string]interface{}{"a": "ax"}},
		{Script: `a = "a"; a[2] = "x"`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": "a"}},

		// index assignment - len 2
		{Script: `a = "ab"; a[0] = "x"`, RunOutput: "x", Output: map[string]interface{}{"a": "xb"}},
		{Script: `a = "ab"; a[1] = "x"`, RunOutput: "x", Output: map[string]interface{}{"a": "ax"}},
		{Script: `a = "ab"; a[2] = "x"`, RunOutput: "x", Output: map[string]interface{}{"a": "abx"}},
		{Script: `a = "ab"; a[3] = "x"`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": "ab"}},

		// index assignment - len 3
		{Script: `a = "abc"; a[0] = "x"`, RunOutput: "x", Output: map[string]interface{}{"a": "xbc"}},
		{Script: `a = "abc"; a[1] = "x"`, RunOutput: "x", Output: map[string]interface{}{"a": "axc"}},
		{Script: `a = "abc"; a[2] = "x"`, RunOutput: "x", Output: map[string]interface{}{"a": "abx"}},
		{Script: `a = "abc"; a[3] = "x"`, RunOutput: "x", Output: map[string]interface{}{"a": "abcx"}},
		{Script: `a = "abc"; a[4] = "x"`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": "abc"}},

		// index assignment - vm types
		{Script: `a = "abc"; a[1] = nil`, RunOutput: nil, Output: map[string]interface{}{"a": "ac"}},
		{Script: `a = "abc"; a[1] = true`, RunError: fmt.Errorf("type bool cannot be assigned to type string"), Output: map[string]interface{}{"a": "abc"}},
		{Script: `a = "abc"; a[1] = 120`, RunOutput: int64(120), Output: map[string]interface{}{"a": "axc"}},
		{Script: `a = "abc"; a[1] = 2.2`, RunError: fmt.Errorf("type float64 cannot be assigned to type string"), Output: map[string]interface{}{"a": "abc"}},
		{Script: `a = "abc"; a[1] = ["a"]`, RunError: fmt.Errorf("type []interface {} cannot be assigned to type string"), Output: map[string]interface{}{"a": "abc"}},

		// index assignment - Go types
		{Script: `a = "abc"; a[1] = b`, Input: map[string]interface{}{"b": reflect.Value{}}, RunError: fmt.Errorf("type reflect.Value cannot be assigned to type string"), Output: map[string]interface{}{"a": "abc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]interface{}{"b": nil}, RunOutput: nil, Output: map[string]interface{}{"a": "ac"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]interface{}{"b": true}, RunError: fmt.Errorf("type bool cannot be assigned to type string"), Output: map[string]interface{}{"a": "abc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]interface{}{"b": int32(120)}, RunOutput: int32(120), Output: map[string]interface{}{"a": "axc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]interface{}{"b": int64(120)}, RunOutput: int64(120), Output: map[string]interface{}{"a": "axc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]interface{}{"b": float32(1.1)}, RunError: fmt.Errorf("type float32 cannot be assigned to type string"), Output: map[string]interface{}{"a": "abc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]interface{}{"b": 2.2}, RunError: fmt.Errorf("type float64 cannot be assigned to type string"), Output: map[string]interface{}{"a": "abc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]interface{}{"b": "x"}, RunOutput: "x", Output: map[string]interface{}{"a": "axc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]interface{}{"b": 'x'}, RunOutput: 'x', Output: map[string]interface{}{"a": "axc"}},
		{Script: `a = "abc"; a[1] = b`, Input: map[string]interface{}{"b": struct{}{}}, RunError: fmt.Errorf("type struct {} cannot be assigned to type string"), Output: map[string]interface{}{"a": "abc"}},
	}
	runTests(t, tests, nil, &Options{Debug: true})
}

func TestVar(t *testing.T) {
	testInput1 := map[string]interface{}{"b": func() {}}
	tests := []Test{
		// simple one variable
		{Script: `1 = 2`, RunError: fmt.Errorf("invalid operation")},
		{Script: `var 1 = 2`, ParseError: fmt.Errorf("syntax error")},
		{Script: `a = 1++`, RunError: fmt.Errorf("invalid operation")},
		{Script: `var a = 1++`, RunError: fmt.Errorf("invalid operation")},
		{Script: `a := 1`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a := 1`, ParseError: fmt.Errorf("syntax error")},
		{Script: `y = z`, RunError: fmt.Errorf("undefined symbol 'z'")},

		{Script: `a = nil`, RunOutput: nil, Output: map[string]interface{}{"a": nil}},
		{Script: `a = true`, RunOutput: true, Output: map[string]interface{}{"a": true}},
		{Script: `a = 1`, RunOutput: int64(1), Output: map[string]interface{}{"a": int64(1)}},
		{Script: `a = 1.1`, RunOutput: 1.1, Output: map[string]interface{}{"a": 1.1}},
		{Script: `a = "a"`, RunOutput: "a", Output: map[string]interface{}{"a": "a"}},

		{Script: `var a = nil`, RunOutput: nil, Output: map[string]interface{}{"a": nil}},
		{Script: `var a = true`, RunOutput: true, Output: map[string]interface{}{"a": true}},
		{Script: `var a = 1`, RunOutput: int64(1), Output: map[string]interface{}{"a": int64(1)}},
		{Script: `var a = 1.1`, RunOutput: 1.1, Output: map[string]interface{}{"a": 1.1}},
		{Script: `var a = "a"`, RunOutput: "a", Output: map[string]interface{}{"a": "a"}},

		{Script: `a = b`, Input: map[string]interface{}{"b": reflect.Value{}}, RunOutput: reflect.Value{}, Output: map[string]interface{}{"a": reflect.Value{}, "b": reflect.Value{}}},
		{Script: `a = b`, Input: map[string]interface{}{"b": nil}, RunOutput: nil, Output: map[string]interface{}{"a": nil, "b": nil}},
		{Script: `a = b`, Input: map[string]interface{}{"b": true}, RunOutput: true, Output: map[string]interface{}{"a": true, "b": true}},
		{Script: `a = b`, Input: map[string]interface{}{"b": int32(1)}, RunOutput: int32(1), Output: map[string]interface{}{"a": int32(1), "b": int32(1)}},
		{Script: `a = b`, Input: map[string]interface{}{"b": int64(1)}, RunOutput: int64(1), Output: map[string]interface{}{"a": int64(1), "b": int64(1)}},
		{Script: `a = b`, Input: map[string]interface{}{"b": float32(1.1)}, RunOutput: float32(1.1), Output: map[string]interface{}{"a": float32(1.1), "b": float32(1.1)}},
		{Script: `a = b`, Input: map[string]interface{}{"b": 1.1}, RunOutput: 1.1, Output: map[string]interface{}{"a": 1.1, "b": 1.1}},
		{Script: `a = b`, Input: map[string]interface{}{"b": "a"}, RunOutput: "a", Output: map[string]interface{}{"a": "a", "b": "a"}},
		{Script: `a = b`, Input: map[string]interface{}{"b": 'a'}, RunOutput: 'a', Output: map[string]interface{}{"a": 'a', "b": 'a'}},
		{Script: `a = b`, Input: map[string]interface{}{"b": struct{}{}}, RunOutput: struct{}{}, Output: map[string]interface{}{"a": struct{}{}, "b": struct{}{}}},

		{Script: `var a = b`, Input: map[string]interface{}{"b": reflect.Value{}}, RunOutput: reflect.Value{}, Output: map[string]interface{}{"a": reflect.Value{}, "b": reflect.Value{}}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": nil}, RunOutput: nil, Output: map[string]interface{}{"a": nil, "b": nil}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": true}, RunOutput: true, Output: map[string]interface{}{"a": true, "b": true}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": int32(1)}, RunOutput: int32(1), Output: map[string]interface{}{"a": int32(1), "b": int32(1)}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": int64(1)}, RunOutput: int64(1), Output: map[string]interface{}{"a": int64(1), "b": int64(1)}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": float32(1.1)}, RunOutput: float32(1.1), Output: map[string]interface{}{"a": float32(1.1), "b": float32(1.1)}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": 1.1}, RunOutput: 1.1, Output: map[string]interface{}{"a": 1.1, "b": 1.1}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": "a"}, RunOutput: "a", Output: map[string]interface{}{"a": "a", "b": "a"}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": 'a'}, RunOutput: 'a', Output: map[string]interface{}{"a": 'a', "b": 'a'}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": struct{}{}}, RunOutput: struct{}{}, Output: map[string]interface{}{"a": struct{}{}, "b": struct{}{}}},

		// simple one variable overwrite
		{Script: `a = true; a = nil`, RunOutput: nil, Output: map[string]interface{}{"a": nil}},
		{Script: `a = nil; a = true`, RunOutput: true, Output: map[string]interface{}{"a": true}},
		{Script: `a = 1; a = 2`, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(2)}},
		{Script: `a = 1.1; a = 2.2`, RunOutput: 2.2, Output: map[string]interface{}{"a": 2.2}},
		{Script: `a = "a"; a = "b"`, RunOutput: "b", Output: map[string]interface{}{"a": "b"}},

		{Script: `var a = true; var a = nil`, RunOutput: nil, Output: map[string]interface{}{"a": nil}},
		{Script: `var a = nil; var a = true`, RunOutput: true, Output: map[string]interface{}{"a": true}},
		{Script: `var a = 1; var a = 2`, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(2)}},
		{Script: `var a = 1.1; var a = 2.2`, RunOutput: 2.2, Output: map[string]interface{}{"a": 2.2}},
		{Script: `var a = "a"; var a = "b"`, RunOutput: "b", Output: map[string]interface{}{"a": "b"}},

		{Script: `a = nil`, Input: map[string]interface{}{"a": true}, RunOutput: nil, Output: map[string]interface{}{"a": nil}},
		{Script: `a = true`, Input: map[string]interface{}{"a": nil}, RunOutput: true, Output: map[string]interface{}{"a": true}},
		{Script: `a = 2`, Input: map[string]interface{}{"a": int32(1)}, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(2)}},
		{Script: `a = 2`, Input: map[string]interface{}{"a": int64(1)}, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(2)}},
		{Script: `a = 2.2`, Input: map[string]interface{}{"a": float32(1.1)}, RunOutput: 2.2, Output: map[string]interface{}{"a": 2.2}},
		{Script: `a = 2.2`, Input: map[string]interface{}{"a": 1.1}, RunOutput: 2.2, Output: map[string]interface{}{"a": 2.2}},
		{Script: `a = "b"`, Input: map[string]interface{}{"a": "a"}, RunOutput: "b", Output: map[string]interface{}{"a": "b"}},

		{Script: `var a = nil`, Input: map[string]interface{}{"a": true}, RunOutput: nil, Output: map[string]interface{}{"a": nil}},
		{Script: `var a = true`, Input: map[string]interface{}{"a": nil}, RunOutput: true, Output: map[string]interface{}{"a": true}},
		{Script: `var a = 2`, Input: map[string]interface{}{"a": int32(1)}, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(2)}},
		{Script: `var a = 2`, Input: map[string]interface{}{"a": int64(1)}, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(2)}},
		{Script: `var a = 2.2`, Input: map[string]interface{}{"a": float32(1.1)}, RunOutput: 2.2, Output: map[string]interface{}{"a": 2.2}},
		{Script: `var a = 2.2`, Input: map[string]interface{}{"a": 1.1}, RunOutput: 2.2, Output: map[string]interface{}{"a": 2.2}},
		{Script: `var a = "b"`, Input: map[string]interface{}{"a": "a"}, RunOutput: "b", Output: map[string]interface{}{"a": "b"}},

		// Go variable copy
		{Script: `a = b`, Input: testInput1, RunOutput: testInput1["b"], Output: map[string]interface{}{"a": testInput1["b"], "b": testInput1["b"]}},
		{Script: `var a = b`, Input: testInput1, RunOutput: testInput1["b"], Output: map[string]interface{}{"a": testInput1["b"], "b": testInput1["b"]}},

		{Script: `a = b`, Input: map[string]interface{}{"b": testVarValue}, RunOutput: testVarValue, Output: map[string]interface{}{"a": testVarValue, "b": testVarValue}},
		{Script: `a = b`, Input: map[string]interface{}{"b": testVarBoolP}, RunOutput: testVarBoolP, Output: map[string]interface{}{"a": testVarBoolP, "b": testVarBoolP}},
		{Script: `a = b`, Input: map[string]interface{}{"b": testVarInt32P}, RunOutput: testVarInt32P, Output: map[string]interface{}{"a": testVarInt32P, "b": testVarInt32P}},
		{Script: `a = b`, Input: map[string]interface{}{"b": testVarInt64P}, RunOutput: testVarInt64P, Output: map[string]interface{}{"a": testVarInt64P, "b": testVarInt64P}},
		{Script: `a = b`, Input: map[string]interface{}{"b": testVarFloat32P}, RunOutput: testVarFloat32P, Output: map[string]interface{}{"a": testVarFloat32P, "b": testVarFloat32P}},
		{Script: `a = b`, Input: map[string]interface{}{"b": testVarFloat64P}, RunOutput: testVarFloat64P, Output: map[string]interface{}{"a": testVarFloat64P, "b": testVarFloat64P}},
		{Script: `a = b`, Input: map[string]interface{}{"b": testVarStringP}, RunOutput: testVarStringP, Output: map[string]interface{}{"a": testVarStringP, "b": testVarStringP}},
		{Script: `a = b`, Input: map[string]interface{}{"b": testVarFuncP}, RunOutput: testVarFuncP, Output: map[string]interface{}{"a": testVarFuncP, "b": testVarFuncP}},

		{Script: `var a = b`, Input: map[string]interface{}{"b": testVarValue}, RunOutput: testVarValue, Output: map[string]interface{}{"a": testVarValue, "b": testVarValue}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": testVarBoolP}, RunOutput: testVarBoolP, Output: map[string]interface{}{"a": testVarBoolP, "b": testVarBoolP}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": testVarInt32P}, RunOutput: testVarInt32P, Output: map[string]interface{}{"a": testVarInt32P, "b": testVarInt32P}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": testVarInt64P}, RunOutput: testVarInt64P, Output: map[string]interface{}{"a": testVarInt64P, "b": testVarInt64P}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": testVarFloat32P}, RunOutput: testVarFloat32P, Output: map[string]interface{}{"a": testVarFloat32P, "b": testVarFloat32P}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": testVarFloat64P}, RunOutput: testVarFloat64P, Output: map[string]interface{}{"a": testVarFloat64P, "b": testVarFloat64P}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": testVarStringP}, RunOutput: testVarStringP, Output: map[string]interface{}{"a": testVarStringP, "b": testVarStringP}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": testVarFuncP}, RunOutput: testVarFuncP, Output: map[string]interface{}{"a": testVarFuncP, "b": testVarFuncP}},

		{Script: `a = b`, Input: map[string]interface{}{"b": testVarValueBool}, RunOutput: testVarValueBool, Output: map[string]interface{}{"a": testVarValueBool, "b": testVarValueBool}},
		{Script: `a = b`, Input: map[string]interface{}{"b": testVarValueInt32}, RunOutput: testVarValueInt32, Output: map[string]interface{}{"a": testVarValueInt32, "b": testVarValueInt32}},
		{Script: `a = b`, Input: map[string]interface{}{"b": testVarValueInt64}, RunOutput: testVarValueInt64, Output: map[string]interface{}{"a": testVarValueInt64, "b": testVarValueInt64}},
		{Script: `a = b`, Input: map[string]interface{}{"b": testVarValueFloat32}, RunOutput: testVarValueFloat32, Output: map[string]interface{}{"a": testVarValueFloat32, "b": testVarValueFloat32}},
		{Script: `a = b`, Input: map[string]interface{}{"b": testVarValueFloat64}, RunOutput: testVarValueFloat64, Output: map[string]interface{}{"a": testVarValueFloat64, "b": testVarValueFloat64}},
		{Script: `a = b`, Input: map[string]interface{}{"b": testVarValueString}, RunOutput: testVarValueString, Output: map[string]interface{}{"a": testVarValueString, "b": testVarValueString}},

		{Script: `var a = b`, Input: map[string]interface{}{"b": testVarValueBool}, RunOutput: testVarValueBool, Output: map[string]interface{}{"a": testVarValueBool, "b": testVarValueBool}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": testVarValueInt32}, RunOutput: testVarValueInt32, Output: map[string]interface{}{"a": testVarValueInt32, "b": testVarValueInt32}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": testVarValueInt64}, RunOutput: testVarValueInt64, Output: map[string]interface{}{"a": testVarValueInt64, "b": testVarValueInt64}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": testVarValueFloat32}, RunOutput: testVarValueFloat32, Output: map[string]interface{}{"a": testVarValueFloat32, "b": testVarValueFloat32}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": testVarValueFloat64}, RunOutput: testVarValueFloat64, Output: map[string]interface{}{"a": testVarValueFloat64, "b": testVarValueFloat64}},
		{Script: `var a = b`, Input: map[string]interface{}{"b": testVarValueString}, RunOutput: testVarValueString, Output: map[string]interface{}{"a": testVarValueString, "b": testVarValueString}},

		// one variable spacing
		{Script: `
a  =  1
`, RunOutput: int64(1)},
		{Script: `
a  =  1;
`, RunOutput: int64(1)},

		// one variable many values
		{Script: `, b = 1, 2`, ParseError: fmt.Errorf("syntax error: unexpected ','"), RunOutput: int64(2), Output: map[string]interface{}{"b": int64(1)}},
		{Script: `var , b = 1, 2`, ParseError: fmt.Errorf("syntax error: unexpected ','"), RunOutput: int64(2), Output: map[string]interface{}{"b": int64(1)}},
		{Script: `a,  = 1, 2`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a,  = 1, 2`, ParseError: fmt.Errorf("syntax error")},

		// TO FIX: should not error?
		{Script: `a = 1, 2`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a = 1, 2`, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(1)}},
		// TO FIX: should not error?
		{Script: `a = 1, 2, 3`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a = 1, 2, 3`, RunOutput: int64(3), Output: map[string]interface{}{"a": int64(1)}},

		// two variables many values
		{Script: `a, b  = 1,`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a, b  = 1,`, ParseError: fmt.Errorf("syntax error")},
		{Script: `a, b  = ,1`, ParseError: fmt.Errorf("syntax error: unexpected ','"), RunOutput: int64(1)},
		{Script: `var a, b  = ,1`, ParseError: fmt.Errorf("syntax error: unexpected ','"), RunOutput: int64(1)},
		{Script: `a, b  = 1,,`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a, b  = 1,,`, ParseError: fmt.Errorf("syntax error")},
		{Script: `a, b  = ,1,`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a, b  = ,1,`, ParseError: fmt.Errorf("syntax error")},
		{Script: `a, b  = ,,1`, ParseError: fmt.Errorf("syntax error")},
		{Script: `var a, b  = ,,1`, ParseError: fmt.Errorf("syntax error")},

		{Script: `a.c, b = 1, 2`, RunError: fmt.Errorf("undefined symbol 'a'")},
		{Script: `a, b.c = 1, 2`, RunError: fmt.Errorf("undefined symbol 'b'")},

		{Script: `a, b = 1`, RunOutput: int64(1), Output: map[string]interface{}{"a": int64(1)}},
		{Script: `var a, b = 1`, RunOutput: int64(1), Output: map[string]interface{}{"a": int64(1)}},
		{Script: `a, b = 1, 2`, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(1), "b": int64(2)}},
		{Script: `var a, b = 1, 2`, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(1), "b": int64(2)}},
		{Script: `a, b = 1, 2, 3`, RunOutput: int64(3), Output: map[string]interface{}{"a": int64(1), "b": int64(2)}},
		{Script: `var a, b = 1, 2, 3`, RunOutput: int64(3), Output: map[string]interface{}{"a": int64(1), "b": int64(2)}},

		// three variables many values
		{Script: `a, b, c = 1`, RunOutput: int64(1), Output: map[string]interface{}{"a": int64(1)}},
		{Script: `var a, b, c = 1`, RunOutput: int64(1), Output: map[string]interface{}{"a": int64(1)}},
		{Script: `a, b, c = 1, 2`, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(1), "b": int64(2)}},
		{Script: `var a, b, c = 1, 2`, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(1), "b": int64(2)}},
		{Script: `a, b, c = 1, 2, 3`, RunOutput: int64(3), Output: map[string]interface{}{"a": int64(1), "b": int64(2), "c": int64(3)}},
		{Script: `var a, b, c = 1, 2, 3`, RunOutput: int64(3), Output: map[string]interface{}{"a": int64(1), "b": int64(2), "c": int64(3)}},
		{Script: `a, b, c = 1, 2, 3, 4`, RunOutput: int64(4), Output: map[string]interface{}{"a": int64(1), "b": int64(2), "c": int64(3)}},
		{Script: `var a, b, c = 1, 2, 3, 4`, RunOutput: int64(4), Output: map[string]interface{}{"a": int64(1), "b": int64(2), "c": int64(3)}},

		// scope
		{Script: `func(){ a = 1 }(); a`, RunError: fmt.Errorf("undefined symbol 'a'")},
		{Script: `func(){ var a = 1 }(); a`, RunError: fmt.Errorf("undefined symbol 'a'")},

		{Script: `a = 1; func(){ a = 2 }()`, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(2)}},
		{Script: `var a = 1; func(){ a = 2 }()`, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(2)}},
		{Script: `a = 1; func(){ var a = 2 }()`, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(1)}},
		{Script: `var a = 1; func(){ var a = 2 }()`, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(1)}},

		// function return
		{Script: `a, 1++ = func(){ return 1, 2 }()`, RunError: fmt.Errorf("invalid operation"), Output: map[string]interface{}{"a": int64(1)}},

		{Script: `a = func(){ return 1 }()`, RunOutput: int64(1), Output: map[string]interface{}{"a": int64(1)}},
		{Script: `var a = func(){ return 1 }()`, RunOutput: int64(1), Output: map[string]interface{}{"a": int64(1)}},
		{Script: `a, b = func(){ return 1, 2 }()`, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(1), "b": int64(2)}},
		{Script: `var a, b = func(){ return 1, 2 }()`, RunOutput: int64(2), Output: map[string]interface{}{"a": int64(1), "b": int64(2)}},
	}
	runTests(t, tests, nil, &Options{Debug: true})
}
