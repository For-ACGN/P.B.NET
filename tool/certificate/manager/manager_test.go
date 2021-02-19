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
	"project/internal/system"
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
	t.Run("common", func(t *testing.T) {
		testCleanTestData(t)
		defer testCleanTestData(t)

		mgr := testNewManager(nil)
		err := mgr.Initialize(testPassword)
		require.NoError(t, err)

		testsuite.IsDestroyed(t, mgr)
	})

	t.Run("file already exist", func(t *testing.T) {
		testCleanTestData(t)
		defer testCleanTestData(t)

		err := system.WriteFile(testFilePath, []byte("test"))
		require.NoError(t, err)

		mgr := testNewManager(nil)
		err = mgr.Initialize(testPassword)
		require.Error(t, err)

		testsuite.IsDestroyed(t, mgr)
	})
}

func TestManager_ResetPassword(t *testing.T) {
	newPassword := []byte("test123")

	t.Run("common", func(t *testing.T) {
		testCleanTestData(t)
		defer testCleanTestData(t)

		// simulate user input
		r, w := io.Pipe()
		defer func() {
			err := r.Close()
			require.NoError(t, err)
			err = w.Close()
			require.NoError(t, err)
		}()

		// initialize certificate manager and reset password
		fmt.Println("================================================")
		mgr := testNewManager(r)
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
	})

	t.Run("file is not exist", func(t *testing.T) {
		testCleanTestData(t)
		defer testCleanTestData(t)

		mgr := testNewManager(nil)
		err := mgr.ResetPassword(testPassword, newPassword)
		require.Error(t, err)

		testsuite.IsDestroyed(t, mgr)
	})

	t.Run("invalid password", func(t *testing.T) {
		testCleanTestData(t)
		defer testCleanTestData(t)

		mgr := testNewManager(nil)
		err := mgr.Initialize(testPassword)
		require.NoError(t, err)

		err = mgr.ResetPassword(newPassword, testPassword)
		require.Error(t, err)
	})
}

func TestManager_Manage(t *testing.T) {
	t.Run("file not exist", func(t *testing.T) {
		testCleanTestData(t)
		defer testCleanTestData(t)

		mgr := testNewManager(nil)
		err := mgr.Manage(testPassword)
		require.Error(t, err)
	})

	t.Run("invalid password", func(t *testing.T) {
		testCleanTestData(t)
		defer testCleanTestData(t)

		mgr := testNewManager(nil)
		err := mgr.Initialize(testPassword)
		require.NoError(t, err)

		err = mgr.Manage(nil)
		require.Error(t, err)
	})
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

	// initialize certificate manager
	fmt.Println("================================================")
	mgr := testNewManager(r)
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

func TestManager_Main(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool

		for _, cmd := range []string{
			"public", "help", "return",
			"private", "help", "return",

			"save", "reload",
			"help", "", "cmd1 cmd2", "invalid-cmd",
			"clear", "reset", "cls",
			"exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixManager, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		testCompareCertPool(t, pool1, pool2, testExceptPublicRootCA)
	})
}

func TestManager_Public(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool

		for _, cmd := range []string{
			"public",

			"root-ca", "help", "return",
			"client-ca", "help", "return",
			"client", "help", "return",

			"save", "reload",
			"help", "", "cmd1 cmd2", "invalid-cmd",
			"clear", "reset", "cls",
			"return", "public", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPublic, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
}

func TestManager_Private(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool

		for _, cmd := range []string{
			"private",

			"root-ca", "help", "return",
			"client-ca", "help", "return",
			"client", "help", "return",

			"save", "reload",
			"help", "", "cmd1 cmd2", "invalid-cmd",
			"clear", "reset", "cls",
			"return", "private", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPrivate, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
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
			"clear", "reset", "cls",
			"return", "root-ca", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPublicRootCA, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
}

func TestManager_PublicRootCA_Add(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		certs1 := pool1.GetPublicRootCACerts()

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
		require.Equal(t, prefixPublicRootCA, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		certs2 := pool2.GetPublicRootCACerts()
		testCompareCertPool(t, pool1, pool2, testExceptPublicRootCA)
		require.True(t, len(certs2)-len(certs1) == 1)
		for i := 0; i < len(certs1); i++ {
			require.Equal(t, certs1[i].Raw, certs2[i].Raw)
		}
	})
}

func TestManager_PublicRootCA_Delete(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		certs1 := pool1.GetPublicRootCACerts()

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
		require.Equal(t, prefixPublicRootCA, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		certs2 := pool2.GetPublicRootCACerts()
		testCompareCertPool(t, pool1, pool2, testExceptPublicRootCA)
		require.True(t, len(certs1)-len(certs2) == 1)
		for i := 0; i < len(certs2); i++ {
			require.Equal(t, certs1[i+1].Raw, certs2[i].Raw)
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
			"export 0 testdata",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPublicRootCA, mgr.prefix)

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
			"clear", "reset", "cls",
			"return", "client-ca", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPublicClientCA, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
}

func TestManager_PublicClientCA_Add(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		certs1 := pool1.GetPublicClientCACerts()

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
		require.Equal(t, prefixPublicClientCA, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		certs2 := pool2.GetPublicClientCACerts()
		testCompareCertPool(t, pool1, pool2, testExceptPublicClientCA)
		require.True(t, len(certs2)-len(certs1) == 1)
		for i := 0; i < len(certs1); i++ {
			require.Equal(t, certs1[i].Raw, certs2[i].Raw)
		}
	})
}

func TestManager_PublicClientCA_Delete(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		certs1 := pool1.GetPublicClientCACerts()

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
		require.Equal(t, prefixPublicClientCA, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		certs2 := pool2.GetPublicClientCACerts()
		testCompareCertPool(t, pool1, pool2, testExceptPublicClientCA)
		require.True(t, len(certs1)-len(certs2) == 1)
		for i := 0; i < len(certs2); i++ {
			require.Equal(t, certs1[i+1].Raw, certs2[i].Raw)
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
			"export 0 testdata",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPublicClientCA, mgr.prefix)

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

func TestManager_PublicClient(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool

		for _, cmd := range []string{
			"public", "client",

			"print 0",
			"print", "print id",
			"print -1", "print 9999",

			"list", "save", "reload",
			"help", " ", "invalid-cmd",
			"clear", "reset", "cls",
			"return", "client", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPublicClient, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
}

func TestManager_PublicClient_Add(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		pairs1 := pool1.GetPublicClientPairs()

		for _, cmd := range []string{
			"public", "client",

			"add testdata/cert.pem testdata/key.pem",
			"add testdata/certs.pem testdata/keys.pem",
			"add ",
			"add testdata/foo.pem testdata/foo.pem",
			"add testdata/broken.pem testdata/foo.pem",
			"add testdata/cert.pem testdata/foo.pem",
			"add testdata/cert.pem testdata/broken.pem",
			"add testdata/certs.pem testdata/key.pem",
			"add testdata/cert.pem testdata/key.pem",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPublicClient, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		pairs2 := pool2.GetPublicClientPairs()
		testCompareCertPool(t, pool1, pool2, testExceptPublicClient)
		require.True(t, len(pairs2)-len(pairs1) == 3)
		for i := 0; i < len(pairs1); i++ {
			require.Equal(t, pairs1[i], pairs2[i])
		}
	})
}

func TestManager_PublicClient_Delete(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		pairs1 := pool1.GetPublicClientPairs()

		for _, cmd := range []string{
			"public", "client",

			"delete 0",
			"delete", "delete id",
			"delete 9999",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPublicClient, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		pairs2 := pool2.GetPublicClientPairs()
		testCompareCertPool(t, pool1, pool2, testExceptPublicClient)
		require.True(t, len(pairs1)-len(pairs2) == 1)
		for i := 0; i < len(pairs2); i++ {
			require.Equal(t, pairs1[i+1], pairs2[i])
		}
	})
}

func TestManager_PublicClient_Export(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool

		for _, cmd := range []string{
			"public", "client",

			"export 0 " + testExportCert + " " + testExportKey,
			"export", "export id path1 path2",
			"export 9999 path1 path2",
			"export 0 testdata testdata",
			"export 0 " + testExportCert + " testdata",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPublicClient, mgr.prefix)

		crt, key := pool1.GetPublicClientPairs()[0].Encode()
		data, err := os.ReadFile(testExportCert)
		require.NoError(t, err)
		block, _ := pem.Decode(data)
		require.NotNil(t, block)
		require.Equal(t, crt, block.Bytes)
		data, err = os.ReadFile(testExportKey)
		require.NoError(t, err)
		block, _ = pem.Decode(data)
		require.NotNil(t, block)
		require.Equal(t, key, block.Bytes)

		pool2 := testGetCertPool(mgr, pool1)
		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
}

func TestManager_PrivateRootCA(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool

		for _, cmd := range []string{
			"private", "root-ca",

			"print 0",
			"print", "print id",
			"print -1", "print 9999",

			"list", "save", "reload",
			"help", " ", "invalid-cmd",
			"clear", "reset", "cls",
			"return", "root-ca", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPrivateRootCA, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
}

func TestManager_PrivateRootCA_Add(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		pairs1 := pool1.GetPrivateRootCAPairs()

		for _, cmd := range []string{
			"private", "root-ca",

			"add testdata/cert.pem testdata/key.pem",
			"add testdata/certs.pem testdata/keys.pem",
			"add ",
			"add testdata/foo.pem testdata/foo.pem",
			"add testdata/broken.pem testdata/foo.pem",
			"add testdata/cert.pem testdata/foo.pem",
			"add testdata/cert.pem testdata/broken.pem",
			"add testdata/certs.pem testdata/key.pem",
			"add testdata/cert.pem testdata/key.pem",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPrivateRootCA, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		pairs2 := pool2.GetPrivateRootCAPairs()
		testCompareCertPool(t, pool1, pool2, testExceptPrivateRootCA)
		require.True(t, len(pairs2)-len(pairs1) == 3)
		for i := 0; i < len(pairs1); i++ {
			require.Equal(t, pairs1[i], pairs2[i])
		}
	})
}

func TestManager_PrivateRootCA_Delete(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		pairs1 := pool1.GetPrivateRootCAPairs()

		for _, cmd := range []string{
			"private", "root-ca",

			"delete 0",
			"delete", "delete id",
			"delete 9999",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPrivateRootCA, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		pairs2 := pool2.GetPrivateRootCAPairs()
		testCompareCertPool(t, pool1, pool2, testExceptPrivateRootCA)
		require.True(t, len(pairs1)-len(pairs2) == 1)
		for i := 0; i < len(pairs2); i++ {
			require.Equal(t, pairs1[i+1], pairs2[i])
		}
	})
}

func TestManager_PrivateRootCA_Export(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool

		for _, cmd := range []string{
			"private", "root-ca",

			"export 0 " + testExportCert + " " + testExportKey,
			"export", "export id path1 path2",
			"export 9999 path1 path2",
			"export 0 testdata testdata",
			"export 0 " + testExportCert + " testdata",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPrivateRootCA, mgr.prefix)

		crt, key := pool1.GetPrivateRootCAPairs()[0].Encode()
		data, err := os.ReadFile(testExportCert)
		require.NoError(t, err)
		block, _ := pem.Decode(data)
		require.NotNil(t, block)
		require.Equal(t, crt, block.Bytes)
		data, err = os.ReadFile(testExportKey)
		require.NoError(t, err)
		block, _ = pem.Decode(data)
		require.NotNil(t, block)
		require.Equal(t, key, block.Bytes)

		pool2 := testGetCertPool(mgr, pool1)
		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
}

func TestManager_PrivateClientCA(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool

		for _, cmd := range []string{
			"private", "client-ca",

			"print 0",
			"print", "print id",
			"print -1", "print 9999",

			"list", "save", "reload",
			"help", " ", "invalid-cmd",
			"clear", "reset", "cls",
			"return", "client-ca", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPrivateClientCA, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
}

func TestManager_PrivateClientCA_Add(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		pairs1 := pool1.GetPrivateClientCAPairs()

		for _, cmd := range []string{
			"private", "client-ca",

			"add testdata/cert.pem testdata/key.pem",
			"add testdata/certs.pem testdata/keys.pem",
			"add ",
			"add testdata/foo.pem testdata/foo.pem",
			"add testdata/broken.pem testdata/foo.pem",
			"add testdata/cert.pem testdata/foo.pem",
			"add testdata/cert.pem testdata/broken.pem",
			"add testdata/certs.pem testdata/key.pem",
			"add testdata/cert.pem testdata/key.pem",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPrivateClientCA, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		pairs2 := pool2.GetPrivateClientCAPairs()
		testCompareCertPool(t, pool1, pool2, testExceptPrivateClientCA)
		require.True(t, len(pairs2)-len(pairs1) == 3)
		for i := 0; i < len(pairs1); i++ {
			require.Equal(t, pairs1[i], pairs2[i])
		}
	})
}

func TestManager_PrivateClientCA_Delete(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		pairs1 := pool1.GetPrivateClientCAPairs()

		for _, cmd := range []string{
			"private", "client-ca",

			"delete 0",
			"delete", "delete id",
			"delete 9999",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPrivateClientCA, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		pairs2 := pool2.GetPrivateClientCAPairs()
		testCompareCertPool(t, pool1, pool2, testExceptPrivateClientCA)
		require.True(t, len(pairs1)-len(pairs2) == 1)
		for i := 0; i < len(pairs2); i++ {
			require.Equal(t, pairs1[i+1], pairs2[i])
		}
	})
}

func TestManager_PrivateClientCA_Export(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool

		for _, cmd := range []string{
			"private", "client-ca",

			"export 0 " + testExportCert + " " + testExportKey,
			"export", "export id path1 path2",
			"export 9999 path1 path2",
			"export 0 testdata testdata",
			"export 0 " + testExportCert + " testdata",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPrivateClientCA, mgr.prefix)

		crt, key := pool1.GetPrivateClientCAPairs()[0].Encode()
		data, err := os.ReadFile(testExportCert)
		require.NoError(t, err)
		block, _ := pem.Decode(data)
		require.NotNil(t, block)
		require.Equal(t, crt, block.Bytes)
		data, err = os.ReadFile(testExportKey)
		require.NoError(t, err)
		block, _ = pem.Decode(data)
		require.NotNil(t, block)
		require.Equal(t, key, block.Bytes)

		pool2 := testGetCertPool(mgr, pool1)
		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
}

func TestManager_PrivateClient(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool

		for _, cmd := range []string{
			"private", "client",

			"print 0",
			"print", "print id",
			"print -1", "print 9999",

			"list", "save", "reload",
			"help", " ", "invalid-cmd",
			"clear", "reset", "cls",
			"return", "client", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPrivateClient, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
}

func TestManager_PrivateClient_Add(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		pairs1 := pool1.GetPrivateClientPairs()

		for _, cmd := range []string{
			"private", "client",

			"add testdata/cert.pem testdata/key.pem",
			"add testdata/certs.pem testdata/keys.pem",
			"add ",
			"add testdata/foo.pem testdata/foo.pem",
			"add testdata/broken.pem testdata/foo.pem",
			"add testdata/cert.pem testdata/foo.pem",
			"add testdata/cert.pem testdata/broken.pem",
			"add testdata/certs.pem testdata/key.pem",
			"add testdata/cert.pem testdata/key.pem",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPrivateClient, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		pairs2 := pool2.GetPrivateClientPairs()
		testCompareCertPool(t, pool1, pool2, testExceptPrivateClient)
		require.True(t, len(pairs2)-len(pairs1) == 3)
		for i := 0; i < len(pairs1); i++ {
			require.Equal(t, pairs1[i], pairs2[i])
		}
	})
}

func TestManager_PrivateClient_Delete(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool
		pairs1 := pool1.GetPrivateClientPairs()

		for _, cmd := range []string{
			"private", "client",

			"delete 0",
			"delete", "delete id",
			"delete 9999",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPrivateClient, mgr.prefix)

		pool2 := testGetCertPool(mgr, pool1)
		pairs2 := pool2.GetPrivateClientPairs()
		testCompareCertPool(t, pool1, pool2, testExceptPrivateClient)
		require.True(t, len(pairs1)-len(pairs2) == 1)
		for i := 0; i < len(pairs2); i++ {
			require.Equal(t, pairs1[i+1], pairs2[i])
		}
	})
}

func TestManager_PrivateClient_Export(t *testing.T) {
	testCleanTestData(t)
	defer testCleanTestData(t)

	testManager(t, func(mgr *Manager, w io.Writer) {
		pool1 := mgr.pool

		for _, cmd := range []string{
			"private", "client",

			"export 0 " + testExportCert + " " + testExportKey,
			"export", "export id path1 path2",
			"export 9999 path1 path2",
			"export 0 testdata testdata",
			"export 0 " + testExportCert + " testdata",

			"save", "reload", "exit",
		} {
			_, err := w.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}
		require.Equal(t, prefixPrivateClient, mgr.prefix)

		crt, key := pool1.GetPrivateClientPairs()[0].Encode()
		data, err := os.ReadFile(testExportCert)
		require.NoError(t, err)
		block, _ := pem.Decode(data)
		require.NotNil(t, block)
		require.Equal(t, crt, block.Bytes)
		data, err = os.ReadFile(testExportKey)
		require.NoError(t, err)
		block, _ = pem.Decode(data)
		require.NotNil(t, block)
		require.Equal(t, key, block.Bytes)

		pool2 := testGetCertPool(mgr, pool1)
		testCompareCertPool(t, pool1, pool2, testExceptNone)
	})
}
