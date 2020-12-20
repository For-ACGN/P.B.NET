package pkg

import (
	"fmt"
	"unsafe"
)

// T is a test structure.
type T struct {
	A int
	b int
}

func f9() {
	fmt.Println(unsafe.Offsetof(T{}.A)) // #nosec
	fmt.Println(unsafe.Offsetof(T{}.b)) // #nosec
}
