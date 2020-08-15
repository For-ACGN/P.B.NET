// Package thirdparty generate by script/code/anko/package.go, don't edit it.
package thirdparty

import (
	"reflect"

	"github.com/mattn/anko/env"
	"github.com/pelletier/go-toml"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/vmihailenco/msgpack/v5/msgpcode"
)

func init() {
	initGithubComPelletierGoTOML()
	initGithubComVmihailencoMsgpackV5()
	initGithubComVmihailencoMsgpackV5Msgpcode()
}

func initGithubComPelletierGoTOML() {
	env.Packages["github.com/pelletier/go-toml"] = map[string]reflect.Value{
		// define constants
		"OrderAlphabetical": reflect.ValueOf(toml.OrderAlphabetical),
		"OrderPreserve":     reflect.ValueOf(toml.OrderPreserve),

		// define variables

		// define functions
		"Load":               reflect.ValueOf(toml.Load),
		"LoadBytes":          reflect.ValueOf(toml.LoadBytes),
		"LoadFile":           reflect.ValueOf(toml.LoadFile),
		"LoadReader":         reflect.ValueOf(toml.LoadReader),
		"LocalDateOf":        reflect.ValueOf(toml.LocalDateOf),
		"LocalDateTimeOf":    reflect.ValueOf(toml.LocalDateTimeOf),
		"LocalTimeOf":        reflect.ValueOf(toml.LocalTimeOf),
		"Marshal":            reflect.ValueOf(toml.Marshal),
		"NewDecoder":         reflect.ValueOf(toml.NewDecoder),
		"NewEncoder":         reflect.ValueOf(toml.NewEncoder),
		"ParseLocalDate":     reflect.ValueOf(toml.ParseLocalDate),
		"ParseLocalDateTime": reflect.ValueOf(toml.ParseLocalDateTime),
		"ParseLocalTime":     reflect.ValueOf(toml.ParseLocalTime),
		"TreeFromMap":        reflect.ValueOf(toml.TreeFromMap),
		"Unmarshal":          reflect.ValueOf(toml.Unmarshal),
	}
	var (
		decoder       toml.Decoder
		encoder       toml.Encoder
		localDate     toml.LocalDate
		localDateTime toml.LocalDateTime
		localTime     toml.LocalTime
		marshaler     toml.Marshaler
		position      toml.Position
		setOptions    toml.SetOptions
		tree          toml.Tree
		unmarshaler   toml.Unmarshaler
	)
	env.PackageTypes["github.com/pelletier/go-toml"] = map[string]reflect.Type{
		"Decoder":       reflect.TypeOf(&decoder).Elem(),
		"Encoder":       reflect.TypeOf(&encoder).Elem(),
		"LocalDate":     reflect.TypeOf(&localDate).Elem(),
		"LocalDateTime": reflect.TypeOf(&localDateTime).Elem(),
		"LocalTime":     reflect.TypeOf(&localTime).Elem(),
		"Marshaler":     reflect.TypeOf(&marshaler).Elem(),
		"Position":      reflect.TypeOf(&position).Elem(),
		"SetOptions":    reflect.TypeOf(&setOptions).Elem(),
		"Tree":          reflect.TypeOf(&tree).Elem(),
		"Unmarshaler":   reflect.TypeOf(&unmarshaler).Elem(),
	}
}

func initGithubComVmihailencoMsgpackV5() {
	env.Packages["github.com/vmihailenco/msgpack/v5"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"GetDecoder":         reflect.ValueOf(msgpack.GetDecoder),
		"GetEncoder":         reflect.ValueOf(msgpack.GetEncoder),
		"Marshal":            reflect.ValueOf(msgpack.Marshal),
		"NewDecoder":         reflect.ValueOf(msgpack.NewDecoder),
		"NewEncoder":         reflect.ValueOf(msgpack.NewEncoder),
		"PutDecoder":         reflect.ValueOf(msgpack.PutDecoder),
		"PutEncoder":         reflect.ValueOf(msgpack.PutEncoder),
		"Register":           reflect.ValueOf(msgpack.Register),
		"RegisterExt":        reflect.ValueOf(msgpack.RegisterExt),
		"RegisterExtDecoder": reflect.ValueOf(msgpack.RegisterExtDecoder),
		"RegisterExtEncoder": reflect.ValueOf(msgpack.RegisterExtEncoder),
		"Unmarshal":          reflect.ValueOf(msgpack.Unmarshal),
	}
	var (
		customDecoder        msgpack.CustomDecoder
		customEncoder        msgpack.CustomEncoder
		decoder              msgpack.Decoder
		encoder              msgpack.Encoder
		marshaler            msgpack.Marshaler
		marshalerUnmarshaler msgpack.MarshalerUnmarshaler
		rawMessage           msgpack.RawMessage
		unmarshaler          msgpack.Unmarshaler
	)
	env.PackageTypes["github.com/vmihailenco/msgpack/v5"] = map[string]reflect.Type{
		"CustomDecoder":        reflect.TypeOf(&customDecoder).Elem(),
		"CustomEncoder":        reflect.TypeOf(&customEncoder).Elem(),
		"Decoder":              reflect.TypeOf(&decoder).Elem(),
		"Encoder":              reflect.TypeOf(&encoder).Elem(),
		"Marshaler":            reflect.TypeOf(&marshaler).Elem(),
		"MarshalerUnmarshaler": reflect.TypeOf(&marshalerUnmarshaler).Elem(),
		"RawMessage":           reflect.TypeOf(&rawMessage).Elem(),
		"Unmarshaler":          reflect.TypeOf(&unmarshaler).Elem(),
	}
}

func initGithubComVmihailencoMsgpackV5Msgpcode() {
	env.Packages["github.com/vmihailenco/msgpack/v5"] = map[string]reflect.Value{
		// define constants

		// define variables
		"Array16":         reflect.ValueOf(msgpcode.Array16),
		"Array32":         reflect.ValueOf(msgpcode.Array32),
		"Bin16":           reflect.ValueOf(msgpcode.Bin16),
		"Bin32":           reflect.ValueOf(msgpcode.Bin32),
		"Bin8":            reflect.ValueOf(msgpcode.Bin8),
		"Double":          reflect.ValueOf(msgpcode.Double),
		"Ext16":           reflect.ValueOf(msgpcode.Ext16),
		"Ext32":           reflect.ValueOf(msgpcode.Ext32),
		"Ext8":            reflect.ValueOf(msgpcode.Ext8),
		"False":           reflect.ValueOf(msgpcode.False),
		"FixExt1":         reflect.ValueOf(msgpcode.FixExt1),
		"FixExt16":        reflect.ValueOf(msgpcode.FixExt16),
		"FixExt2":         reflect.ValueOf(msgpcode.FixExt2),
		"FixExt4":         reflect.ValueOf(msgpcode.FixExt4),
		"FixExt8":         reflect.ValueOf(msgpcode.FixExt8),
		"FixedArrayHigh":  reflect.ValueOf(msgpcode.FixedArrayHigh),
		"FixedArrayLow":   reflect.ValueOf(msgpcode.FixedArrayLow),
		"FixedArrayMask":  reflect.ValueOf(msgpcode.FixedArrayMask),
		"FixedMapHigh":    reflect.ValueOf(msgpcode.FixedMapHigh),
		"FixedMapLow":     reflect.ValueOf(msgpcode.FixedMapLow),
		"FixedMapMask":    reflect.ValueOf(msgpcode.FixedMapMask),
		"FixedStrHigh":    reflect.ValueOf(msgpcode.FixedStrHigh),
		"FixedStrLow":     reflect.ValueOf(msgpcode.FixedStrLow),
		"FixedStrMask":    reflect.ValueOf(msgpcode.FixedStrMask),
		"Float":           reflect.ValueOf(msgpcode.Float),
		"Int16":           reflect.ValueOf(msgpcode.Int16),
		"Int32":           reflect.ValueOf(msgpcode.Int32),
		"Int64":           reflect.ValueOf(msgpcode.Int64),
		"Int8":            reflect.ValueOf(msgpcode.Int8),
		"Map16":           reflect.ValueOf(msgpcode.Map16),
		"Map32":           reflect.ValueOf(msgpcode.Map32),
		"NegFixedNumLow":  reflect.ValueOf(msgpcode.NegFixedNumLow),
		"Nil":             reflect.ValueOf(msgpcode.Nil),
		"PosFixedNumHigh": reflect.ValueOf(msgpcode.PosFixedNumHigh),
		"Str16":           reflect.ValueOf(msgpcode.Str16),
		"Str32":           reflect.ValueOf(msgpcode.Str32),
		"Str8":            reflect.ValueOf(msgpcode.Str8),
		"True":            reflect.ValueOf(msgpcode.True),
		"Uint16":          reflect.ValueOf(msgpcode.Uint16),
		"Uint32":          reflect.ValueOf(msgpcode.Uint32),
		"Uint64":          reflect.ValueOf(msgpcode.Uint64),
		"Uint8":           reflect.ValueOf(msgpcode.Uint8),

		// define functions
		"IsBin":         reflect.ValueOf(msgpcode.IsBin),
		"IsExt":         reflect.ValueOf(msgpcode.IsExt),
		"IsFixedArray":  reflect.ValueOf(msgpcode.IsFixedArray),
		"IsFixedExt":    reflect.ValueOf(msgpcode.IsFixedExt),
		"IsFixedMap":    reflect.ValueOf(msgpcode.IsFixedMap),
		"IsFixedNum":    reflect.ValueOf(msgpcode.IsFixedNum),
		"IsFixedString": reflect.ValueOf(msgpcode.IsFixedString),
		"IsString":      reflect.ValueOf(msgpcode.IsString),
	}
	var ()
	env.PackageTypes["github.com/vmihailenco/msgpack/v5"] = map[string]reflect.Type{}
}
