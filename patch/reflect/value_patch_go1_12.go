// +build go1.10,!go1.12

package reflect

import (
	"unsafe"
)

// A MapIter is an iterator for ranging over a map.
// See Value.MapRange.
//
// From go1.12
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
//
// From go1.12
func (v Value) MapRange() *MapIter {
	v.mustBe(Map)
	return &MapIter{m: v}
}

// copyVal returns a Value containing the map key or value at ptr,
// allocating a new variable as needed.
//
// From go1.12
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
//
// From go1.12
func mapiterelem(it unsafe.Pointer) (elem unsafe.Pointer)
