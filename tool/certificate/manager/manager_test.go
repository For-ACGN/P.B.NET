package manager

import (
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/cert"
	"project/internal/cert/certpool"
	"project/internal/testsuite"
)

const (
	testFilePath   = "testdata/key/certpool.bin"
	testExportCert = "testdata/export/cert.pem"
	testExportKey  = "testdata/export/key.pem"
)

var testPassword = []byte("test")

func testCleanTestData(t *testing.T) {
	err := os.RemoveAll("testdata/key")
	require.NoError(t, err)
	err = os.RemoveAll("testdata/export")
	require.NoError(t, err)
}

func testNewManager(r io.Reader) *Manager {
	mgr := New(r, testFilePath)
	mgr.testMode = true
	return mgr
}

func TestManager_Initialize(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	mgr := testNewManager(nil)
	err := mgr.Initialize(testPassword)
	require.NoError(t, err)

	testsuite.IsDestroyed(t, mgr)
}

func TestManager_ResetPassword(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

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

	// generate test certificates parallel
	opts := cert.Options{Algorithm: "ed25519"}
	pairs := make(chan *cert.Pair, 6)
	for i := 0; i < 6; i++ {
		go func() {
			pair, err := cert.GenerateCA(&opts)
			require.NoError(t, err)
			pairs <- pair
		}()
	}

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

	// add test certificates
	err = mgr.pool.AddPublicRootCACert((<-pairs).ASN1())
	require.NoError(t, err)
	err = mgr.pool.AddPublicClientCACert((<-pairs).ASN1())
	require.NoError(t, err)
	err = mgr.pool.AddPublicClientPair((<-pairs).Encode())
	require.NoError(t, err)
	err = mgr.pool.AddPrivateRootCAPair((<-pairs).Encode())
	require.NoError(t, err)
	err = mgr.pool.AddPrivateClientCAPair((<-pairs).Encode())
	require.NoError(t, err)
	err = mgr.pool.AddPrivateClientPair((<-pairs).Encode())
	require.NoError(t, err)

	fn(mgr, w)

	wg.Wait()

	fmt.Println("================================================")

	testsuite.IsDestroyed(t, mgr)
}

func testGetCertPool(mgr *Manager, old *certpool.Pool) *certpool.Pool {
	for {
		v := mgr.testPool.Load()
		if v == nil {
			continue
		}
		pool := v.(*certpool.Pool)
		if pool != old {
			return pool
		}
	}
}

func TestManager_SaveAndReload(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		certs0 := pool1.GetPublicRootCACerts()

		for _, cmd := range []string{
			"public", "root-ca",
			"delete 0",
			"save", "reload",
			"exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}

		pool2 := testGetCertPool(mgr, pool1)
		certs1 := pool2.GetPublicRootCACerts()

		require.True(t, len(certs0)-len(certs1) == 1)
		for i := 0; i < len(certs1); i++ {
			require.Equal(t, certs0[i+1].Raw, certs1[i].Raw)
		}
	})
}

const (
	testExceptNone = iota
	testExceptPublicRootCA
	testExceptPublicClientCA
	testExceptPublicClient
	testExceptPrivateRootCA
	testExceptPrivateClientCA
	testExceptPrivateClient
)

func testCompareCertPool(t *testing.T, cp1, cp2 *certpool.Pool, except int) {
	if except != testExceptPublicRootCA {
		certs1 := cp1.GetPublicRootCACerts()
		certs2 := cp2.GetPublicRootCACerts()
		require.Equal(t, certs1, certs2)
	}
	if except != testExceptPublicClientCA {
		certs1 := cp1.GetPublicClientCACerts()
		certs2 := cp2.GetPublicClientCACerts()
		require.Equal(t, certs1, certs2)
	}
	if except != testExceptPublicClient {
		pairs1 := cp1.GetPublicClientPairs()
		pairs2 := cp2.GetPublicClientPairs()
		require.Equal(t, pairs1, pairs2)
	}
	if except != testExceptPrivateRootCA {
		pairs1 := cp1.GetPrivateRootCAPairs()
		pairs2 := cp2.GetPrivateRootCAPairs()
		require.Equal(t, pairs1, pairs2)
	}
	if except != testExceptPrivateClientCA {
		pairs1 := cp1.GetPrivateClientCAPairs()
		pairs2 := cp2.GetPrivateClientCAPairs()
		require.Equal(t, pairs1, pairs2)
	}
	if except != testExceptPrivateClient {
		pairs1 := cp1.GetPrivateClientPairs()
		pairs2 := cp2.GetPrivateClientPairs()
		require.Equal(t, pairs1, pairs2)
	}
}

func TestManager_PublicRootCA(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool

		for _, cmd := range []string{
			"public", "root-ca",

			"print 0",
			"print", "print id",
			"print -1", "print 9999",

			"list", "save", "reload",
			"help", " ", "invalid-cmd",
			"return", "root-ca", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}

		pool2 := testGetCertPool(mgr, pool1)

		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
}

func TestManager_PublicRootCA_Add(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		certs0 := pool1.GetPublicRootCACerts()

		for _, cmd := range []string{
			"public", "root-ca",

			"add testdata/cert.pem",
			"add ", "add foo.pem",
			"add testdata/broken.pem",
			"add testdata/cert.pem",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}

		pool2 := testGetCertPool(mgr, pool1)
		certs1 := pool2.GetPublicRootCACerts()

		testCompareCertPool(t, pool1, pool2, testExceptPublicRootCA)
		require.True(t, len(certs1)-len(certs0) == 1)
		for i := 0; i < len(certs0); i++ {
			require.Equal(t, certs0[i].Raw, certs1[i].Raw)
		}
	})
}

func TestManager_PublicRootCA_Delete(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		certs0 := pool1.GetPublicRootCACerts()

		for _, cmd := range []string{
			"public", "root-ca",

			"delete 0",
			"delete", "delete id",
			"delete 9999",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}

		pool2 := testGetCertPool(mgr, pool1)
		certs1 := pool2.GetPublicRootCACerts()

		testCompareCertPool(t, pool1, pool2, testExceptPublicRootCA)
		require.True(t, len(certs0)-len(certs1) == 1)
		for i := 0; i < len(certs1); i++ {
			require.Equal(t, certs0[i+1].Raw, certs1[i].Raw)
		}
	})
}

func TestManager_PublicRootCA_Export(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool

		for _, cmd := range []string{
			"public", "root-ca",

			"export 0 " + testExportCert,
			"export", "export id path",
			"export 9999 path",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}

		raw := pool1.GetPublicRootCACerts()[0].Raw
		data, err := os.ReadFile(testExportCert)
		require.NoError(t, err)
		block, _ := pem.Decode(data)
		require.NotNil(t, block)
		require.Equal(t, raw, block.Bytes)

		pool2 := testGetCertPool(mgr, pool1)

		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
}

func TestManager_PublicClientCA(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool

		for _, cmd := range []string{
			"public", "client-ca",

			"print 0",
			"print", "print id",
			"print -1", "print 9999",

			"list", "save", "reload",
			"help", " ", "invalid-cmd",
			"return", "client-ca", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}

		pool2 := testGetCertPool(mgr, pool1)

		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
}

func TestManager_PublicClientCA_Add(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		certs0 := pool1.GetPublicClientCACerts()

		for _, cmd := range []string{
			"public", "client-ca",

			"add testdata/cert.pem",
			"add ", "add foo.pem",
			"add testdata/broken.pem",
			"add testdata/cert.pem",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}

		pool2 := testGetCertPool(mgr, pool1)
		certs1 := pool2.GetPublicClientCACerts()

		testCompareCertPool(t, pool1, pool2, testExceptPublicClientCA)
		require.True(t, len(certs1)-len(certs0) == 1)
		for i := 0; i < len(certs0); i++ {
			require.Equal(t, certs0[i].Raw, certs1[i].Raw)
		}
	})
}

func TestManager_PublicClientCA_Delete(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		certs0 := pool1.GetPublicClientCACerts()

		for _, cmd := range []string{
			"public", "client-ca",

			"delete 0",
			"delete", "delete id",
			"delete 9999",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}

		pool2 := testGetCertPool(mgr, pool1)
		certs1 := pool2.GetPublicClientCACerts()

		testCompareCertPool(t, pool1, pool2, testExceptPublicClientCA)
		require.True(t, len(certs0)-len(certs1) == 1)
		for i := 0; i < len(certs1); i++ {
			require.Equal(t, certs0[i+1].Raw, certs1[i].Raw)
		}
	})
}

func TestManager_PublicClientCA_Export(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool

		for _, cmd := range []string{
			"public", "client-ca",

			"export 0 " + testExportCert,
			"export", "export id path",
			"export 9999 path",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}

		raw := pool1.GetPublicClientCACerts()[0].Raw
		data, err := os.ReadFile(testExportCert)
		require.NoError(t, err)
		block, _ := pem.Decode(data)
		require.NotNil(t, block)
		require.Equal(t, raw, block.Bytes)

		pool2 := testGetCertPool(mgr, pool1)

		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
}
