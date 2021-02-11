package manager

import (
	"fmt"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

const testFilePath = "testdata/key/certpool.bin"

var testPassword = []byte("test")

func testClean() {
	err := os.RemoveAll("testdata/key")
	testsuite.TestMainCheckError(err)
}

func testNewManager(r io.Reader) *Manager {
	mgr := New(r, testFilePath)
	mgr.testMode = true
	return mgr
}

func TestInitialize(t *testing.T) {
	testClean()
	defer testClean()

	mgr := testNewManager(nil)
	err := mgr.Initialize(testPassword)
	require.NoError(t, err)

	testsuite.IsDestroyed(t, mgr)
}

func TestResetPassword(t *testing.T) {
	testClean()
	defer testClean()

	newPassword := []byte("test123")

	// simulate user input
	r, w := io.Pipe()
	defer func() {
		err := r.Close()
		require.NoError(t, err)
		err = w.Close()
		require.NoError(t, err)
	}()
	mgr := testNewManager(r)

	fmt.Println("================================================")
	err := mgr.Initialize(testPassword)
	require.NoError(t, err)
	fmt.Println("================================================")
	err = mgr.ResetPassword(testPassword, newPassword)
	require.NoError(t, err)
	fmt.Println("================================================")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := mgr.Manage(newPassword)
		require.NoError(t, err)
	}()

	for _, cmd := range []string{
		"help", "exit",
	} {
		_, err = w.Write([]byte(cmd + "\n"))
		require.NoError(t, err)
	}

	wg.Wait()

	fmt.Println("================================================")

	testsuite.IsDestroyed(t, mgr)
}

func testManager(t *testing.T, fn func(mgr *Manager, w io.Writer)) {
	testClean()
	defer testClean()

	// simulate user input
	r, w := io.Pipe()
	defer func() {
		err := r.Close()
		require.NoError(t, err)
		err = w.Close()
		require.NoError(t, err)
	}()
	mgr := testNewManager(r)

	fmt.Println("================================================")
	err := mgr.Initialize(testPassword)
	require.NoError(t, err)
	fmt.Println("================================================")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := mgr.Manage(testPassword)
		require.NoError(t, err)
	}()

	// make sure readCommandLoop is running
	_, err = w.Write([]byte("help\n"))
	require.NoError(t, err)

	fn(mgr, w)

	_, err = w.Write([]byte("exit\n"))
	require.NoError(t, err)

	wg.Wait()

	fmt.Println("================================================")

	testsuite.IsDestroyed(t, mgr)
}

func TestReloadAndSave(t *testing.T) {
	testManager(t, func(mgr *Manager, w io.Writer) {
		n0 := len(mgr.pool.GetPublicRootCACerts())

		for _, cmd := range []string{
			"public", "root-ca",
			"delete 0",
			"save", "reload",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}

		n1 := len(mgr.pool.GetPublicRootCACerts())
		require.True(t, n0-n1 == 1)
	})
}
