package packages

import (
	"math"
	"reflect"

	"project/external/anko/env"
)

func init() {
	env.Packages["math"] = map[string]reflect.Value{
		"Abs":             reflect.ValueOf(math.Abs),
		"Acos":            reflect.ValueOf(math.Acos),
		"Acosh":           reflect.ValueOf(math.Acosh),
		"Asin":            reflect.ValueOf(math.Asin),
		"Asinh":           reflect.ValueOf(math.Asinh),
		"Atan":            reflect.ValueOf(math.Atan),
		"Atan2":           reflect.ValueOf(math.Atan2),
		"Atanh":           reflect.ValueOf(math.Atanh),
		"Cbrt":            reflect.ValueOf(math.Cbrt),
		"Ceil":            reflect.ValueOf(math.Ceil),
		"Copysign":        reflect.ValueOf(math.Copysign),
		"Cos":             reflect.ValueOf(math.Cos),
		"Cosh":            reflect.ValueOf(math.Cosh),
		"Dim":             reflect.ValueOf(math.Dim),
		"Erf":             reflect.ValueOf(math.Erf),
		"Erfc":            reflect.ValueOf(math.Erfc),
		"Exp":             reflect.ValueOf(math.Exp),
		"Exp2":            reflect.ValueOf(math.Exp2),
		"Expm1":           reflect.ValueOf(math.Expm1),
		"Float32bits":     reflect.ValueOf(math.Float32bits),
		"Float32frombits": reflect.ValueOf(math.Float32frombits),
		"Float64bits":     reflect.ValueOf(math.Float64bits),
		"Float64frombits": reflect.ValueOf(math.Float64frombits),
		"Floor":           reflect.ValueOf(math.Floor),
		"Frexp":           reflect.ValueOf(math.Frexp),
		"Gamma":           reflect.ValueOf(math.Gamma),
		"Hypot":           reflect.ValueOf(math.Hypot),
		"Ilogb":           reflect.ValueOf(math.Ilogb),
		"Inf":             reflect.ValueOf(math.Inf),
		"IsInf":           reflect.ValueOf(math.IsInf),
		"IsNaN":           reflect.ValueOf(math.IsNaN),
		"J0":              reflect.ValueOf(math.J0),
		"J1":              reflect.ValueOf(math.J1),
		"Jn":              reflect.ValueOf(math.Jn),
		"Ldexp":           reflect.ValueOf(math.Ldexp),
		"Lgamma":          reflect.ValueOf(math.Lgamma),
		"Log":             reflect.ValueOf(math.Log),
		"Log10":           reflect.ValueOf(math.Log10),
		"Log1p":           reflect.ValueOf(math.Log1p),
		"Log2":            reflect.ValueOf(math.Log2),
		"Logb":            reflect.ValueOf(math.Logb),
		"Max":             reflect.ValueOf(math.Max),
		"Min":             reflect.ValueOf(math.Min),
		"Mod":             reflect.ValueOf(math.Mod),
		"Modf":            reflect.ValueOf(math.Modf),
		"NaN":             reflect.ValueOf(math.NaN),
		"Nextafter":       reflect.ValueOf(math.Nextafter),
		"Pow":             reflect.ValueOf(math.Pow),
		"Pow10":           reflect.ValueOf(math.Pow10),
		"Remainder":       reflect.ValueOf(math.Remainder),
		"Signbit":         reflect.ValueOf(math.Signbit),
		"Sin":             reflect.ValueOf(math.Sin),
		"Sincos":          reflect.ValueOf(math.Sincos),
		"Sinh":            reflect.ValueOf(math.Sinh),
		"Sqrt":            reflect.ValueOf(math.Sqrt),
		"Tan":             reflect.ValueOf(math.Tan),
		"Tanh":            reflect.ValueOf(math.Tanh),
		"Trunc":           reflect.ValueOf(math.Trunc),
		"Y0":              reflect.ValueOf(math.Y0),
		"Y1":              reflect.ValueOf(math.Y1),
		"Yn":              reflect.ValueOf(math.Yn),
	}
}