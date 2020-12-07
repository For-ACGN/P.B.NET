package xpanic

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

func TestPrintStack(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		testFuncA()
	})

	t.Run("skip > max depth", func(t *testing.T) {
		buf := new(bytes.Buffer)
		PrintStack(buf, maxDepth+1)

		fmt.Println("-----begin-----")
		fmt.Println(buf)
		fmt.Println("-----end-----")
	})

	t.Run("panic", func(t *testing.T) {
		patch := func(uintptr) *runtime.Func {
			panic(monkey.Panic)
		}
		pg := monkey.Patch(runtime.FuncForPC, patch)
		defer pg.Unpatch()

		testLogPanic()
	})

	t.Run("unknown frame", func(t *testing.T) {
		patch := func(uintptr) *runtime.Func {
			return nil
		}
		pg := monkey.Patch(runtime.FuncForPC, patch)
		defer pg.Unpatch()

		defer func() {
			r := recover()
			fmt.Println("-----begin-----")
			fmt.Println(Error(r, "TestUnknown title"))
			fmt.Println("-----end-----")
		}()

		testPanic()
	})
}

func testFuncA() {
	testFuncB()
}

func testFuncB() {
	testFuncC()
}

func testFuncC() {
	// appear some error
	testLogPanic()
}

func testPanic() {
	var foo []int
	foo[0] = 0
}

func testLogPanic() {
	buf := new(bytes.Buffer)
	PrintStack(buf, 0)

	fmt.Println("-----begin-----")
	fmt.Println(buf)
	fmt.Println("-----end-----")
}

func TestError(t *testing.T) {
	defer func() {
		r := recover()
		fmt.Println("-----begin-----")
		fmt.Println(Error(r, "TestError title"))
		fmt.Println("-----end-----")
	}()

	testPanic()
}

func TestErrorf(t *testing.T) {
	defer func() {
		r := recover()
		fmt.Println("-----begin-----")
		fmt.Println(Errorf(r, "TestError title %s", "errorf"))
		fmt.Println("-----end-----")
	}()

	testPanic()
}

func TestLog(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				buf := Log(r, "testLog")
				require.NotNil(t, buf)
			}
		}()

		testPanic()
	}()
	wg.Wait()
}

func TestLogf(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				buf := Logf(r, "testLog %s", "logf")
				require.NotNil(t, buf)
			}
		}()

		testPanic()
	}()
	wg.Wait()
}
