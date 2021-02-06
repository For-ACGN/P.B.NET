package manager

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/term"
	"project/internal/cert/certpool"
	"project/internal/testsuite"

	"project/internal/patch/monkey"
)

func TestMain(m *testing.M) {
	testClean()
	filePath = "testdata/certpool.bin"
	testMode = true

	// hook os.Exit()
	patch := func(code int) {
		if code != 0 {
			panic("error occurred in test")
		}
	}
	monkey.Patch(os.Exit, patch)

	code := m.Run()

	testClean()
	os.Exit(code)
}

func testClean() {
	err := os.RemoveAll("testdata")
	testsuite.TestMainCheckError(err)
}

func TestInitialize(t *testing.T) {
	patch := func(int) ([]byte, error) {
		return []byte("test"), nil
	}
	pg := monkey.Patch(term.ReadPassword, patch)
	defer pg.Unpatch()

	initMgr = true
	defer func() { initMgr = false }()

	fmt.Println("================================================")
	fmt.Println(initMgr)
	main()
	fmt.Println("================================================")

	initMgr = false

	// simulate user input
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer func() {
		err = r.Close()
		require.NoError(t, err)
		err = w.Close()
		require.NoError(t, err)
	}()
	stdin := os.Stdin
	defer func() { os.Stdin = stdin }()
	os.Stdin = r

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		main()
	}()

	for _, cmd := range []string{
		"help", "exit",
	} {
		_, err = w.WriteString(cmd + "\n")
		require.NoError(t, err)
	}

	wg.Wait()
}

func TestResetPassword(t *testing.T) {
	var n int
	patch := func(int) ([]byte, error) {
		n++
		if n < 4 {
			return []byte("test"), nil
		}
		return []byte("test123"), nil
	}
	pg := monkey.Patch(term.ReadPassword, patch)
	defer pg.Unpatch()

	fmt.Println("================================================")
	initialize()
	fmt.Println("================================================")
	resetPassword()
	fmt.Println("================================================")

	// simulate user input
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer func() {
		err = r.Close()
		require.NoError(t, err)
		err = w.Close()
		require.NoError(t, err)
	}()
	stdin := os.Stdin
	defer func() { os.Stdin = stdin }()
	os.Stdin = r

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		manage()
	}()

	for _, cmd := range []string{
		"help", "exit",
	} {
		_, err = w.WriteString(cmd + "\n")
		require.NoError(t, err)
	}

	wg.Wait()
}

func testManager(t *testing.T, fn func(w *os.File)) {
	patch := func(int) ([]byte, error) {
		return []byte("test"), nil
	}
	pg := monkey.Patch(term.ReadPassword, patch)
	defer pg.Unpatch()

	fmt.Println("================================================")
	main()
	fmt.Println("================================================")

	// simulate user input
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer func() {
		err = r.Close()
		require.NoError(t, err)
		err = w.Close()
		require.NoError(t, err)
	}()
	stdin := os.Stdin
	defer func() { os.Stdin = stdin }()
	os.Stdin = r

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		manage()
	}()

	fn(w)

	_, err = w.WriteString("exit\n")
	require.NoError(t, err)

	wg.Wait()
}

func testGetCertPool(old *certpool.Pool) *certpool.Pool {
	for {
		v := testCertPool.Load()
		if v == nil {
			continue
		}
		pool := v.(*certpool.Pool)
		if pool != old {
			return pool
		}
	}
}

func TestReloadAndSave(t *testing.T) {
	testManager(t, func(w *os.File) {
		pool := testGetCertPool(nil)

		n0 := len(pool.GetPublicRootCACerts())

		for _, cmd := range []string{
			"public", "root-ca",
			"delete 0",
			"save", "reload",
		} {
			_, err := w.WriteString(cmd + "\n")
			require.NoError(t, err)
		}

		pool = testGetCertPool(pool)

		n1 := len(pool.GetPublicRootCACerts())
		require.True(t, n0-n1 == 1)
	})
}
