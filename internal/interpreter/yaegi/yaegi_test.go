package yaegi

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"

	"project/internal/interpreter/yaegi/compiler"
	"project/internal/interpreter/yaegi/goroot/unsafe"
)

func TestChannel(t *testing.T) {
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)

	const src = `
package main

import (
	"fmt"
)

func main() {
	cha := make(chan int, 1)
	chb := make(chan int, 1)
	cha <- 1
	select {
	case <-cha:
		fmt.Println("1")
	case chb <- 1:
		fmt.Println("2")
	}
}
`
	_, err := i.Eval(src)
	require.NoError(t, err)
}

func TestUnsafe(t *testing.T) {
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)
	i.Use(unsafe.Symbols)

	const src = `
package main

import (
	"fmt"
	"unsafe"
)

// can use like const but not set to array.
const a = unsafe.Sizeof("a") + unsafe.Sizeof("a")

type at struct{
	aa []int
	// AA [a]int  this is not allowed
	bb string
	BB string
}

func main(){
	fmt.Println(a)

	fmt.Println(unsafe.Sizeof(at{}))
	fmt.Println(unsafe.Sizeof(at{}.aa))
	fmt.Println(unsafe.Sizeof(at{}.bb))
	fmt.Println(unsafe.Sizeof(at{}.BB))

	fmt.Println(unsafe.Alignof(at{}))
	fmt.Println(unsafe.Alignof(at{}.aa))
	fmt.Println(unsafe.Alignof(at{}.bb))
	fmt.Println(unsafe.Alignof(at{}.BB))

	fmt.Println(unsafe.Offsetof(at{}.aa))
	fmt.Println(unsafe.Offsetof(at{}.bb))
	fmt.Println(unsafe.Offsetof(at{}.BB))
}
`
	_, err := i.Eval(compiler.ProcessUnsafeOffsetof(src))
	require.NoError(t, err)
}
