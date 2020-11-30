// +build go1.10, !go1.11

package reflect

import (
	"math"
	"unsafe"
)

// IsZero reports whether v is the zero value for its type.
// It panics if the argument is invalid.
func (v Value) IsZero() bool {
	switch v.kind() {
	case Bool:
		return !v.Bool()
	case Int, Int8, Int16, Int32, Int64:
		return v.Int() == 0
	case Uint, Uint8, Uint16, Uint32, Uint64, Uintptr:
		return v.Uint() == 0
	case Float32, Float64:
		return math.Float64bits(v.Float()) == 0
	case Complex64, Complex128:
		c := v.Complex()
		return math.Float64bits(real(c)) == 0 && math.Float64bits(imag(c)) == 0
	case Array:
		for i := 0; i < v.Len(); i++ {
			if !v.Index(i).IsZero() {
				return false
			}
		}
		return true
	case Chan, Func, Interface, Map, Ptr, Slice, UnsafePointer:
		return v.IsNil()
	case String:
		return v.Len() == 0
	case Struct:
		for i := 0; i < v.NumField(); i++ {
			if !v.Field(i).IsZero() {
				return false
			}
		}
		return true
	default:
		// This should never happens, but will act as a safeguard for
		// later, as a default value doesn't makes sense here.
		panic(&ValueError{"reflect.Value.IsZero", v.Kind()})
	}
}

// A MapIter is an iterator for ranging over a map.
// See Value.MapRange.
type MapIter struct {
	m  Value
	it unsafe.Pointer
}

// Key returns the key of the iterator's current map entry.
func (it *MapIter) Key() Value {
	if it.it == nil {
		panic("MapIter.Key called before Next")
	}
	if mapiterkey(it.it) == nil {
		panic("MapIter.Key called on exhausted iterator")
	}

	t := (*mapType)(unsafe.Pointer(it.m.typ))
	ktype := t.key
	return copyVal(ktype, it.m.flag.ro()|flag(ktype.Kind()), mapiterkey(it.it))
}

// Value returns the value of the iterator's current map entry.
func (it *MapIter) Value() Value {
	if it.it == nil {
		panic("MapIter.Value called before Next")
	}
	if mapiterkey(it.it) == nil {
		panic("MapIter.Value called on exhausted iterator")
	}

	t := (*mapType)(unsafe.Pointer(it.m.typ))
	vtype := t.elem
	return copyVal(vtype, it.m.flag.ro()|flag(vtype.Kind()), mapiterelem(it.it))
}

// Next advances the map iterator and reports whether there is another
// entry. It returns false when the iterator is exhausted; subsequent
// calls to Key, Value, or Next will panic.
func (it *MapIter) Next() bool {
	if it.it == nil {
		it.it = mapiterinit(it.m.typ, it.m.pointer())
	} else {
		if mapiterkey(it.it) == nil {
			panic("MapIter.Next called on exhausted iterator")
		}
		mapiternext(it.it)
	}
	return mapiterkey(it.it) != nil
}

// MapRange returns a range iterator for a map.
// It panics if v's Kind is not Map.
//
// Call Next to advance the iterator, and Key/Value to access each entry.
// Next returns false when the iterator is exhausted.
// MapRange follows the same iteration semantics as a range statement.
//
// Example:
//
//	iter := reflect.ValueOf(m).MapRange()
// 	for iter.Next() {
//		k := iter.Key()
//		v := iter.Value()
//		...
//	}
//
func (v Value) MapRange() *MapIter {
	v.mustBe(Map)
	return &MapIter{m: v}
}

// copyVal returns a Value containing the map key or value at ptr,
// allocating a new variable as needed.
func copyVal(typ *rtype, fl flag, ptr unsafe.Pointer) Value {
	if ifaceIndir(typ) {
		// Copy result so future changes to the map
		// won't change the underlying value.
		c := unsafe_New(typ)
		typedmemmove(typ, c, ptr)
		return Value{typ, c, fl | flagIndir}
	}
	return Value{typ, *(*unsafe.Pointer)(ptr), fl}
}

//go:noescape
func mapiterelem(it unsafe.Pointer) (elem unsafe.Pointer)
