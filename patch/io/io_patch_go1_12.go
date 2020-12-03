// +build go1.10,!go1.12

package io

// StringWriter is the interface that wraps the WriteString method.
//
// From go1.12
type StringWriter interface {
	WriteString(s string) (n int, err error)
}
