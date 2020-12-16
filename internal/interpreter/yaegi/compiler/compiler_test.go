package compiler

import (
	"go/build"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	ctx := build.Default
	ctx.GOOS = "windows"
	ctx.GOARCH = "amd64"
	// set release tags
	releaseTags := make([]string, 0, 16)
	for i := 1; i <= 16; i++ {
		releaseTags = append(releaseTags, "go1."+strconv.Itoa(i))
	}
	ctx.BuildTags = releaseTags

	code, err := Compile(&ctx, "testdata/pkg")
	require.NoError(t, err)
	const expected = `
package pkg

import (
	"fmt"
	fmt1 "fmt"
	"log"
	_ "log"
	_ "strings"
)

func f1() {
	fmt.Println("f1")
}

func f2() {
	fmt.Println("f2")
	log.Println("f2")
}

func f3() {
	fmt1.Println("f3")
}

// no import

func f4() {
	println("f4")
}

func f8() {
	fmt.Println("f8")
}
`
	require.Equal(t, expected[1:], code)
}
