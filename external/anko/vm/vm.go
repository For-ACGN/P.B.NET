package vm

import (
	"context"
	"fmt"
	"reflect"

	"project/external/anko/ast"
	"project/external/anko/env"
	"project/external/anko/parser"
	"project/internal/xpanic"
)

var (
	stringType         = reflect.TypeOf("a")
	byteType           = reflect.TypeOf(byte('a'))
	runeType           = reflect.TypeOf('a')
	interfaceType      = reflect.ValueOf([]interface{}{int64(1)}).Index(0).Type()
	interfaceSliceType = reflect.TypeOf([]interface{}{})
	reflectValueType   = reflect.TypeOf(reflect.Value{})
	errorType          = reflect.ValueOf([]error{nil}).Index(0).Type()
	vmErrorType        = reflect.TypeOf(&Error{})
	contextType        = reflect.TypeOf((*context.Context)(nil)).Elem()

	nilValue                  = reflect.New(reflect.TypeOf((*interface{})(nil)).Elem()).Elem()
	zeroValue                 = reflect.Value{}
	trueValue                 = reflect.ValueOf(true)
	falseValue                = reflect.ValueOf(false)
	reflectValueNilValue      = reflect.ValueOf(nilValue)
	reflectValueErrorNilValue = reflect.ValueOf(reflect.New(errorType).Elem())
)

// Options provides options to run VM with.
type Options struct {
	Debug      bool // run in Debug mode
	StackTrace bool // print stack trace if appear panic
}

// Execute parses script and executes in the specified environment.
func Execute(env *env.Env, opts *Options, script string) (interface{}, error) {
	stmt, err := parser.ParseSrc(script)
	if err != nil {
		return nilValue, err
	}
	return RunContext(context.Background(), env, opts, stmt)
}

// ExecuteContext parses script and executes in the specified environment with context.
func ExecuteContext(ctx context.Context, env *env.Env, opts *Options, script string) (interface{}, error) {
	stmt, err := parser.ParseSrc(script)
	if err != nil {
		return nilValue, err
	}
	return RunContext(ctx, env, opts, stmt)
}

// Run executes statement in the specified environment.
func Run(env *env.Env, options *Options, stmt ast.Stmt) (interface{}, error) {
	return RunContext(context.Background(), env, options, stmt)
}

// RunContext executes statement in the specified environment with context.
func RunContext(ctx context.Context, env *env.Env, opts *Options, stmt ast.Stmt) (interface{}, error) {
	ri := runInfo{ctx: ctx, env: env, opts: opts, stmt: stmt, rv: nilValue}
	if ri.opts == nil {
		ri.opts = &Options{}
	}
	ri.runSingleStmt()
	if ri.err == ErrReturn {
		ri.err = nil
	}
	return ri.rv.Interface(), ri.err
}

// recoverFunc generic recover function.
func recoverFunc(ri *runInfo) {
	r := recover()
	if r == nil {
		return
	}
	switch value := r.(type) {
	case *Error:
		ri.err = value
	case error:
		ri.err = value
	default:
		ri.err = fmt.Errorf("%v", r)
	}
	if ri.opts.StackTrace {
		b := xpanic.PrintPanic(r, "capture panic in anko vm", 3)
		ri.err = fmt.Errorf("err: %s\n%s", ri.err, b)
	}
}

func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		// from reflect IsNil:
		// Note that IsNil is not always equivalent to a regular comparison with nil in Go.
		// For example, if v was created by calling ValueOf with an uninitialized interface variable i,
		// i==nil will be true but v.IsNil will panic as v will be the zero Value.
		return v.IsNil()
	default:
		return false
	}
}

func isNum(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// equal returns true when lhsV and rhsV is same value.
// nolint: gocyclo
//gocyclo:ignore
func equal(lhsV, rhsV reflect.Value) bool {
	lhsIsNil, rhsIsNil := isNil(lhsV), isNil(rhsV)
	if lhsIsNil && rhsIsNil {
		return true
	}
	if (!lhsIsNil && rhsIsNil) || (lhsIsNil && !rhsIsNil) {
		return false
	}
	if lhsV.Kind() == reflect.Interface || lhsV.Kind() == reflect.Ptr {
		lhsV = lhsV.Elem()
	}
	if rhsV.Kind() == reflect.Interface || rhsV.Kind() == reflect.Ptr {
		rhsV = rhsV.Elem()
	}

	// Compare a string and a number.
	// This will attempt to convert the string to a number,
	// while leaving the other side alone. Code further
	// down takes care of converting integers and floats as needed.
	if isNum(lhsV) && rhsV.Kind() == reflect.String {
		rhsF, err := tryToFloat64(rhsV)
		if err != nil {
			// Couldn't convert RHS to a float, they can't be compared.
			return false
		}
		rhsV = reflect.ValueOf(rhsF)
	} else if lhsV.Kind() == reflect.String && isNum(rhsV) {
		// If the LHS is a string formatted as an int, try that before trying float
		lhsI, err := tryToInt64(lhsV)
		if err != nil {
			// if LHS is a float, e.g. "1.2", we need to set lhsV to a float64
			lhsF, err := tryToFloat64(lhsV)
			if err != nil {
				return false
			}
			lhsV = reflect.ValueOf(lhsF)
		} else {
			lhsV = reflect.ValueOf(lhsI)
		}
	}

	if isNum(lhsV) && isNum(rhsV) {
		return fmt.Sprintf("%v", lhsV) == fmt.Sprintf("%v", rhsV)
	}

	// Try to compare booleans to strings and numbers
	if lhsV.Kind() == reflect.Bool || rhsV.Kind() == reflect.Bool {
		lhsB, err := tryToBool(lhsV)
		if err != nil {
			return false
		}
		rhsB, err := tryToBool(rhsV)
		if err != nil {
			return false
		}
		return lhsB == rhsB
	}

	return reflect.DeepEqual(lhsV.Interface(), rhsV.Interface())
}

func getMapIndex(key reflect.Value, aMap reflect.Value) reflect.Value {
	if aMap.IsNil() {
		return nilValue
	}

	var err error
	key, err = convertReflectValueToType(key, aMap.Type().Key())
	if err != nil {
		return nilValue
	}

	// From reflect MapIndex:
	// It returns the zero Value if key is not found in the map or if v represents a nil map.
	value := aMap.MapIndex(key)
	if !value.IsValid() {
		return nilValue
	}

	if aMap.Type().Elem() == interfaceType && !value.IsNil() {
		value = reflect.ValueOf(value.Interface())
	}

	return value
}

// appendSlice appends rhs to lhs, function assumes lhsV and rhsV are slice or array.
// nolint: gocyclo
//gocyclo:ignore
func appendSlice(expr ast.Expr, lhsV reflect.Value, rhsV reflect.Value) (reflect.Value, error) {
	lhsT := lhsV.Type().Elem()
	rhsT := rhsV.Type().Elem()

	if lhsT == rhsT {
		return reflect.AppendSlice(lhsV, rhsV), nil
	}

	if rhsT.ConvertibleTo(lhsT) {
		for i := 0; i < rhsV.Len(); i++ {
			lhsV = reflect.Append(lhsV, rhsV.Index(i).Convert(lhsT))
		}
		return lhsV, nil
	}

	leftHasSubArray := lhsT.Kind() == reflect.Slice || lhsT.Kind() == reflect.Array
	rightHasSubArray := rhsT.Kind() == reflect.Slice || rhsT.Kind() == reflect.Array

	if leftHasSubArray != rightHasSubArray && lhsT != interfaceType && rhsT != interfaceType {
		return nilValue, newStringError(expr, "invalid type conversion")
	}

	if !leftHasSubArray && !rightHasSubArray {
		for i := 0; i < rhsV.Len(); i++ {
			value := rhsV.Index(i)
			if rhsT == interfaceType {
				value = value.Elem()
			}
			if lhsT == value.Type() {
				lhsV = reflect.Append(lhsV, value)
			} else if value.Type().ConvertibleTo(lhsT) {
				lhsV = reflect.Append(lhsV, value.Convert(lhsT))
			} else {
				return nilValue, newStringError(expr, "invalid type conversion")
			}
		}
		return lhsV, nil
	}

	if (leftHasSubArray || lhsT == interfaceType) && (rightHasSubArray || rhsT == interfaceType) {
		for i := 0; i < rhsV.Len(); i++ {
			value := rhsV.Index(i)
			if rhsT == interfaceType {
				value = value.Elem()
				if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
					return nilValue, newStringError(expr, "invalid type conversion")
				}
			}
			newSlice, err := appendSlice(expr, reflect.MakeSlice(lhsT, 0, value.Len()), value)
			if err != nil {
				return nilValue, err
			}
			lhsV = reflect.Append(lhsV, newSlice)
		}
		return lhsV, nil
	}

	return nilValue, newStringError(expr, "invalid type conversion")
}

// nolint: gocyclo
//gocyclo:ignore
func makeType(ri *runInfo, typeStruct *ast.TypeStruct) reflect.Type {
	switch typeStruct.Kind {
	case ast.TypeDefault:
		return getTypeFromEnv(ri, typeStruct)
	case ast.TypePtr:
		var t reflect.Type
		if typeStruct.SubType != nil {
			t = makeType(ri, typeStruct.SubType)
		} else {
			t = getTypeFromEnv(ri, typeStruct)
		}
		if ri.err != nil {
			return nil
		}
		if t == nil {
			return nil
		}
		return reflect.PtrTo(t)
	case ast.TypeSlice:
		var t reflect.Type
		if typeStruct.SubType != nil {
			t = makeType(ri, typeStruct.SubType)
		} else {
			t = getTypeFromEnv(ri, typeStruct)
		}
		if ri.err != nil {
			return nil
		}
		if t == nil {
			return nil
		}
		for i := 1; i < typeStruct.Dimensions; i++ {
			t = reflect.SliceOf(t)
		}
		return reflect.SliceOf(t)
	case ast.TypeMap:
		key := makeType(ri, typeStruct.Key)
		if ri.err != nil {
			return nil
		}
		if key == nil {
			return nil
		}
		t := makeType(ri, typeStruct.SubType)
		if ri.err != nil {
			return nil
		}
		if t == nil {
			return nil
		}
		if !ri.opts.Debug {
			// captures panic
			defer recoverFunc(ri)
		}
		t = reflect.MapOf(key, t)
		return t
	case ast.TypeChan:
		var t reflect.Type
		if typeStruct.SubType != nil {
			t = makeType(ri, typeStruct.SubType)
		} else {
			t = getTypeFromEnv(ri, typeStruct)
		}
		if ri.err != nil {
			return nil
		}
		if t == nil {
			return nil
		}
		return reflect.ChanOf(reflect.BothDir, t)
	case ast.TypeStructType:
		var t reflect.Type
		fields := make([]reflect.StructField, 0, len(typeStruct.StructNames))
		for i := 0; i < len(typeStruct.StructNames); i++ {
			t = makeType(ri, typeStruct.StructTypes[i])
			if ri.err != nil {
				return nil
			}
			if t == nil {
				return nil
			}
			fields = append(fields, reflect.StructField{Name: typeStruct.StructNames[i], Type: t})
		}
		if !ri.opts.Debug {
			// captures panic
			defer recoverFunc(ri)
		}
		t = reflect.StructOf(fields)
		return t
	default:
		ri.err = fmt.Errorf("unknown kind")
		return nil
	}
}

func getTypeFromEnv(ri *runInfo, typeStruct *ast.TypeStruct) reflect.Type {
	var e *env.Env
	e, ri.err = ri.env.GetEnvFromPath(typeStruct.Env)
	if ri.err != nil {
		return nil
	}

	var t reflect.Type
	t, ri.err = e.Type(typeStruct.Name)
	return t
}

func makeValue(t reflect.Type) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.Chan:
		return reflect.MakeChan(t, 0), nil
	case reflect.Func:
		return reflect.MakeFunc(t, nil), nil
	case reflect.Map:
		// note creating slice as work around to create map
		// just doing MakeMap can give incorrect type for defined types
		value := reflect.MakeSlice(reflect.SliceOf(t), 0, 1)
		value = reflect.Append(value, reflect.MakeMap(reflect.MapOf(t.Key(), t.Elem())))
		return value.Index(0), nil
	case reflect.Ptr:
		ptrV := reflect.New(t.Elem())
		v, err := makeValue(t.Elem())
		if err != nil {
			return nilValue, err
		}
		ptrV.Elem().Set(v)
		return ptrV, nil
	case reflect.Slice:
		return reflect.MakeSlice(t, 0, 0), nil
	}
	return reflect.New(t).Elem(), nil
}

// precedenceOfKinds returns the greater of two kinds, string > float > int.
func precedenceOfKinds(kind1 reflect.Kind, kind2 reflect.Kind) reflect.Kind {
	if kind1 == kind2 {
		return kind1
	}
	switch kind1 {
	case reflect.String:
		return kind1
	case reflect.Float64, reflect.Float32:
		switch kind2 {
		case reflect.String:
			return kind2
		}
		return kind1
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch kind2 {
		case reflect.String, reflect.Float64, reflect.Float32:
			return kind2
		}
	}
	return kind1
}
