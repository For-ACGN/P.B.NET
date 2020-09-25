package vm

import (
	"fmt"
	"reflect"

	"project/external/anko/ast"
	"project/external/anko/env"
)

// invokeExpr evaluates one expression.
// nolint: gocyclo
//gocyclo:ignore
func (runInfo *runInfoStruct) invokeExpr() {
	switch expr := runInfo.expr.(type) {

	// OpExpr
	case *ast.OpExpr:
		runInfo.operator = expr.Op
		runInfo.invokeOperator()

	// IdentExpr
	case *ast.IdentExpr:
		runInfo.rv, runInfo.err = runInfo.env.GetValue(expr.Lit)
		if runInfo.err != nil {
			runInfo.err = newError(expr, runInfo.err)
		}

	// LiteralExpr
	case *ast.LiteralExpr:
		runInfo.rv = expr.Literal

	// ArrayExpr
	case *ast.ArrayExpr:
		if expr.TypeData == nil {
			slice := make([]interface{}, len(expr.Exprs))
			var i int
			for i, runInfo.expr = range expr.Exprs {
				runInfo.invokeExpr()
				if runInfo.err != nil {
					return
				}
				slice[i] = runInfo.rv.Interface()
			}
			runInfo.rv = reflect.ValueOf(slice)
			return
		}

		t := makeType(runInfo, expr.TypeData)
		if runInfo.err != nil {
			runInfo.rv = nilValue
			return
		}
		if t == nil {
			runInfo.err = newStringError(expr, "cannot make type nil")
			runInfo.rv = nilValue
			return
		}

		slice := reflect.MakeSlice(t, len(expr.Exprs), len(expr.Exprs))
		var i int
		valueType := t.Elem()
		for i, runInfo.expr = range expr.Exprs {
			runInfo.invokeExpr()
			if runInfo.err != nil {
				return
			}

			runInfo.rv, runInfo.err = convertReflectValueToType(runInfo.rv, valueType)
			if runInfo.err != nil {
				const format = "cannot use type %s as type %s as slice value"
				errStr := fmt.Sprintf(format, runInfo.rv.Type(), valueType)
				runInfo.err = newStringError(expr, errStr)
				runInfo.rv = nilValue
				return
			}

			slice.Index(i).Set(runInfo.rv)
		}
		runInfo.rv = slice

		// MapExpr
	case *ast.MapExpr:
		if expr.TypeData == nil {
			var i int
			var key reflect.Value
			m := make(map[interface{}]interface{}, len(expr.Keys))
			for i, runInfo.expr = range expr.Keys {
				runInfo.invokeExpr()
				if runInfo.err != nil {
					return
				}
				key = runInfo.rv

				runInfo.expr = expr.Values[i]
				runInfo.invokeExpr()
				if runInfo.err != nil {
					return
				}

				m[key.Interface()] = runInfo.rv.Interface()
			}
			runInfo.rv = reflect.ValueOf(m)
			return
		}

		t := makeType(runInfo, expr.TypeData)
		if runInfo.err != nil {
			runInfo.rv = nilValue
			return
		}
		if t == nil {
			runInfo.err = newStringError(expr, "cannot make type nil")
			runInfo.rv = nilValue
			return
		}

		runInfo.rv, runInfo.err = makeValue(t)
		if runInfo.err != nil {
			runInfo.rv = nilValue
			return
		}

		var i int
		var key reflect.Value
		m := runInfo.rv
		keyType := t.Key()
		valueType := t.Elem()
		for i, runInfo.expr = range expr.Keys {
			runInfo.invokeExpr()
			if runInfo.err != nil {
				return
			}
			key, runInfo.err = convertReflectValueToType(runInfo.rv, keyType)
			if runInfo.err != nil {
				const format = "cannot use type %s as type %s as map key"
				runInfo.err = newStringError(expr, fmt.Sprintf(format, key.Type(), keyType))
				runInfo.rv = nilValue
				return
			}

			runInfo.expr = expr.Values[i]
			runInfo.invokeExpr()
			if runInfo.err != nil {
				return
			}
			runInfo.rv, runInfo.err = convertReflectValueToType(runInfo.rv, valueType)
			if runInfo.err != nil {
				const format = "cannot use type %s as type %s as map value"
				runInfo.err = newStringError(expr, fmt.Sprintf(format, runInfo.rv.Type(), valueType))
				runInfo.rv = nilValue
				return
			}

			m.SetMapIndex(key, runInfo.rv)
		}
		runInfo.rv = m

	// dereferenceExpr
	case *ast.DerefExpr:
		runInfo.expr = expr.Expr
		runInfo.invokeExpr()
		if runInfo.err != nil {
			return
		}

		if runInfo.rv.Kind() != reflect.Ptr {
			runInfo.err = newStringError(expr.Expr, "cannot deference non-pointer")
			runInfo.rv = nilValue
			return
		}
		runInfo.rv = runInfo.rv.Elem()

	// AddrExpr
	case *ast.AddrExpr:
		runInfo.expr = expr.Expr
		runInfo.invokeExpr()
		if runInfo.err != nil {
			return
		}

		if runInfo.rv.CanAddr() {
			runInfo.rv = runInfo.rv.Addr()
		} else {
			i := runInfo.rv.Interface()
			runInfo.rv = reflect.ValueOf(&i)
		}

	// UnaryExpr
	case *ast.UnaryExpr:
		runInfo.expr = expr.Expr
		runInfo.invokeExpr()
		if runInfo.err != nil {
			return
		}

		switch expr.Operator {
		case "-":
			switch runInfo.rv.Kind() {
			case reflect.Int64:
				runInfo.rv = reflect.ValueOf(-runInfo.rv.Int())
			case reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int, reflect.Bool:
				runInfo.rv = reflect.ValueOf(-toInt64(runInfo.rv))
			case reflect.Float64:
				runInfo.rv = reflect.ValueOf(-runInfo.rv.Float())
			default:
				runInfo.rv = reflect.ValueOf(-toFloat64(runInfo.rv))
			}
		case "^":
			runInfo.rv = reflect.ValueOf(^toInt64(runInfo.rv))
		case "!":
			if toBool(runInfo.rv) {
				runInfo.rv = falseValue
			} else {
				runInfo.rv = trueValue
			}
		default:
			runInfo.err = newStringError(expr, "unknown operator")
			runInfo.rv = nilValue
		}

	// ParenExpr
	case *ast.ParenExpr:
		runInfo.expr = expr.SubExpr
		runInfo.invokeExpr()
		if runInfo.err != nil {
			return
		}

	// MemberExpr
	case *ast.MemberExpr:
		runInfo.expr = expr.Expr
		runInfo.invokeExpr()
		if runInfo.err != nil {
			return
		}

		if runInfo.rv.Kind() == reflect.Interface && !runInfo.rv.IsNil() {
			runInfo.rv = runInfo.rv.Elem()
		}

		if e, ok := runInfo.rv.Interface().(*env.Env); ok {
			runInfo.rv, runInfo.err = e.GetValue(expr.Name)
			if runInfo.err != nil {
				runInfo.err = newError(expr, runInfo.err)
				runInfo.rv = nilValue
			}
			return
		}

		value := runInfo.rv.MethodByName(expr.Name)
		if value.IsValid() {
			runInfo.rv = value
			return
		}

		if runInfo.rv.Kind() == reflect.Ptr {
			runInfo.rv = runInfo.rv.Elem()
		}

		switch runInfo.rv.Kind() {
		case reflect.Struct:
			field, found := runInfo.rv.Type().FieldByName(expr.Name)
			if found {
				runInfo.rv = runInfo.rv.FieldByIndex(field.Index)
				return
			}
			if runInfo.rv.CanAddr() {
				runInfo.rv = runInfo.rv.Addr()
				method, found := runInfo.rv.Type().MethodByName(expr.Name)
				if found {
					runInfo.rv = runInfo.rv.Method(method.Index)
					return
				}
			}
			runInfo.err = newStringError(expr, "no member named '"+expr.Name+"' for struct")
			runInfo.rv = nilValue
		case reflect.Map:
			runInfo.rv = getMapIndex(reflect.ValueOf(expr.Name), runInfo.rv)
		default:
			runInfo.err = newStringError(expr, "type "+runInfo.rv.Kind().String()+" does not support member operation")
			runInfo.rv = nilValue
		}

	default:
		runInfo.err = newStringError(expr, "unknown expression")
		runInfo.rv = nilValue
	}
}
