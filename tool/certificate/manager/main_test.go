package main

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/term"
	"project/internal/testsuite"

	"project/internal/patch/monkey"
)

var testPath = "testdata/certpool.bin"

func TestMain(m *testing.M) {
	err := os.RemoveAll("testdata")
	testsuite.TestMainCheckError(err)

	patch := func(code int) {
		if code != 0 {
			panic("error occurred in test")
		}
	}
	monkey.Patch(os.Exit, patch)

	code := m.Run()

	err = os.RemoveAll("testdata")
	testsuite.TestMainCheckError(err)

	os.Exit(code)
}

func TestInitialize(t *testing.T) {
	patch := func(int) ([]byte, error) {
		return []byte("test"), nil
	}
	pg := monkey.Patch(term.ReadPassword, patch)
	defer pg.Unpatch()

	initialize(testPath)

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
		manage(testPath)
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

	initialize(testPath)
	resetPassword(testPath)
	manage(testPath)
}
