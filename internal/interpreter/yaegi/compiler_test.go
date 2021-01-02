package yaegi

import (
	"go/build"
	"go/format"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

func TestCompile(t *testing.T) {
	t.Run("dir", func(t *testing.T) {
		ctx := build.Default
		ctx.GOOS = "windows"
		ctx.GOARCH = "amd64"
		// set release tags
		releaseTags := make([]string, 0, 16)
		for i := 1; i <= 16; i++ {
			releaseTags = append(releaseTags, "go1."+strconv.Itoa(i))
		}
		ctx.BuildTags = releaseTags
		ctx.ReleaseTags = releaseTags

		code, err := CompileWithContext(&ctx, "testdata/pkg")
		require.NoError(t, err)
		const expected = `
package pkg

import (
	"fmt"
	fmt1 "fmt"
	"log"
	_ "log"
	_ "strings"
	"unsafe"
)

func f1() {
	fmt.Println("f1")
}

func f2() {
	fmt.Println("f2")
	log.Println("f2")
}

func f3() {
	_, _ = fmt1.Println("f3")
}

// no import

func f4() {
	println("f4")
}

func f8() {
	fmt.Println("f8")
}

// T is a test structure.
type T struct {
	A int
	b int
}

func f9() {
	fmt.Println(unsafe.Offsetof(T{}, "A")) // #nosec
	fmt.Println(unsafe.Offsetof(T{}, "Xb")) // #nosec
}
`
		require.Equal(t, expected[1:], code)
	})

	t.Run("file", func(t *testing.T) {
		code, err := Compile("testdata/pkg/a.go")
		require.NoError(t, err)
		const expected = `
package pkg

import (
	"fmt"
)

func f1() {
	fmt.Println("f1")
}
`
		require.Equal(t, expected[1:], code)
	})

	t.Run("invalid path", func(t *testing.T) {
		code, err := Compile("testdata/foo")
		require.Error(t, err)
		require.Zero(t, code)
	})
}

func TestCompileFiles(t *testing.T) {
	code, err := CompileFiles("testdata/pkg", []string{"a.go", "b.go", "foo.go"})
	require.NoError(t, err)
	const expected = `
package pkg

import (
	"fmt"
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
`
	require.Equal(t, expected[1:], code)
}

func TestMergeFiles(t *testing.T) {
	patch := func(string) ([]os.FileInfo, error) {
		return nil, monkey.Error
	}
	pg := monkey.Patch(ioutil.ReadDir, patch)
	defer pg.Unpatch()

	code, err := MergeFiles(defaultContext, "testdata/pkg", []string{"a.go", "b.go"})
	require.Error(t, err)
	require.Zero(t, code)
}

func TestMergeDir(t *testing.T) {
	t.Run("failed to import dir", func(t *testing.T) {
		code, err := MergeDir(defaultContext, "testdata/foo")
		require.Error(t, err)
		require.Zero(t, code)
	})

	t.Run("contain invalid go files", func(t *testing.T) {
		var ctx *build.Context
		patch := func(interface{}, string, build.ImportMode) (*build.Package, error) {
			return &build.Package{InvalidGoFiles: []string{"foo"}}, nil
		}
		pg := monkey.PatchInstanceMethod(ctx, "ImportDir", patch)
		defer pg.Unpatch()

		code, err := MergeDir(defaultContext, "testdata/pkg")
		require.Error(t, err)
		require.Zero(t, code)
	})

	t.Run("failed to read file", func(t *testing.T) {
		patch := func(string) ([]byte, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(ioutil.ReadFile, patch)
		defer pg.Unpatch()

		code, err := MergeDir(defaultContext, "testdata/pkg")
		require.Error(t, err)
		require.Zero(t, code)
	})

	t.Run("failed to format source", func(t *testing.T) {
		patch := func([]byte) ([]byte, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(format.Source, patch)
		defer pg.Unpatch()

		code, err := MergeDir(defaultContext, "testdata/pkg")
		require.Error(t, err)
		require.Zero(t, code)
	})
}

func TestProcessUnsafeOffsetof(t *testing.T) {
	const code = `
func f1() {
	unsafe.Offsetof(T{}.A)
}

func f2() {
	unsafe.Offsetof(T{}.A)
}
`
	output := ProcessUnsafeOffsetof(code[1:])
	const expected = `
func f1() {
	unsafe.Offsetof(T{}, "A")
}

func f2() {
	unsafe.Offsetof(T{}, "A")
}
`
	require.Equal(t, expected[1:], output)
}
