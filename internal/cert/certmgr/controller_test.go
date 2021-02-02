package certmgr

import (
	"compress/flate"
	"crypto/sha256"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/cert"
	"project/internal/cert/certpool"
	"project/internal/crypto/aes"
	"project/internal/crypto/hmac"
	"project/internal/patch/monkey"
	"project/internal/patch/msgpack"
)

func TestCtrlCertMgr_Dump(t *testing.T) {
	invalidCert := []byte("foo")
	invalidPair := struct {
		Cert []byte `msgpack:"a"`
		Key  []byte `msgpack:"b"`
	}{
		Cert: []byte("foo"),
		Key:  []byte("bar"),
	}

	pool := certpool.NewPool()
	ccm := new(ctrlCertMgr)

	t.Run("invalid public root ca cert", func(t *testing.T) {
		ccm.PublicRootCACerts = [][]byte{invalidCert}
		err := ccm.Dump(pool)
		require.Error(t, err)
		ccm.PublicRootCACerts = nil
	})

	t.Run("invalid public client ca cert", func(t *testing.T) {
		ccm.PublicClientCACerts = [][]byte{invalidCert}
		err := ccm.Dump(pool)
		require.Error(t, err)
		ccm.PublicClientCACerts = nil
	})

	t.Run("invalid public client pair", func(t *testing.T) {
		ccm.PublicClientPairs = []struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{invalidPair}
		err := ccm.Dump(pool)
		require.Error(t, err)
		ccm.PublicClientPairs = nil
	})

	t.Run("invalid private root ca pair", func(t *testing.T) {
		ccm.PrivateRootCAPairs = []struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{invalidPair}
		err := ccm.Dump(pool)
		require.Error(t, err)
		ccm.PrivateRootCAPairs = nil
	})

	t.Run("invalid private client ca pair", func(t *testing.T) {
		ccm.PrivateClientCAPairs = []struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{invalidPair}
		err := ccm.Dump(pool)
		require.Error(t, err)
		ccm.PrivateClientCAPairs = nil
	})

	t.Run("invalid private client pair", func(t *testing.T) {
		ccm.PrivateClientPairs = []struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{invalidPair}
		err := ccm.Dump(pool)
		require.Error(t, err)
		ccm.PrivateClientPairs = nil
	})
}

var testPassword = []byte("admin")

func testGenerateCertPool(t *testing.T) *certpool.Pool {
	// load system certificates
	pool, err := certpool.NewPoolWithSystem()
	require.NoError(t, err)

	// create Root CA certificate
	rootCA, err := cert.GenerateCA(nil)
	require.NoError(t, err)
	err = pool.AddPrivateRootCAPair(rootCA.Encode())
	require.NoError(t, err)

	// create Client CA certificate
	clientCA, err := cert.GenerateCA(nil)
	require.NoError(t, err)
	err = pool.AddPublicClientCACert(clientCA.ASN1())
	require.NoError(t, err)
	err = pool.AddPrivateClientCAPair(clientCA.Encode())
	require.NoError(t, err)

	// generate a client certificate and use client CA sign it
	clientCert, err := cert.Generate(clientCA.Certificate, clientCA.PrivateKey, nil)
	require.NoError(t, err)
	err = pool.AddPublicClientPair(clientCert.Encode())
	require.NoError(t, err)
	err = pool.AddPrivateClientPair(clientCert.Encode())
	require.NoError(t, err)
	return pool
}

func TestSaveCtrlCertPool(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		pool := testGenerateCertPool(t)
		data, err := SaveCtrlCertPool(pool, testPassword)
		require.NoError(t, err)
		require.NotNil(t, data)
	})

	pool := testGenerateCertPool(t)

	t.Run("invalid structure", func(t *testing.T) {
		patch := func(interface{}) ([]byte, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(msgpack.Marshal, patch)
		defer pg.Unpatch()

		data, err := SaveCtrlCertPool(pool, testPassword)
		monkey.IsMonkeyError(t, err)
		require.Nil(t, data)
	})

	t.Run("failed to NewWriter", func(t *testing.T) {
		patch := func(io.Writer, int) (*flate.Writer, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(flate.NewWriter, patch)
		defer pg.Unpatch()

		data, err := SaveCtrlCertPool(pool, testPassword)
		monkey.IsExistMonkeyError(t, err)
		require.Nil(t, data)
	})

	t.Run("failed to write about compress", func(t *testing.T) {
		writer := new(flate.Writer)
		patch := func(interface{}, []byte) (int, error) {
			return 0, monkey.Error
		}
		pg := monkey.PatchInstanceMethod(writer, "Write", patch)
		defer pg.Unpatch()

		data, err := SaveCtrlCertPool(pool, testPassword)
		monkey.IsExistMonkeyError(t, err)
		require.Nil(t, data)
	})

	t.Run("failed to close about compress", func(t *testing.T) {
		writer := new(flate.Writer)
		patch := func(interface{}) error {
			return monkey.Error
		}
		pg := monkey.PatchInstanceMethod(writer, "Close", patch)
		defer pg.Unpatch()

		data, err := SaveCtrlCertPool(pool, testPassword)
		monkey.IsExistMonkeyError(t, err)
		require.Nil(t, data)
	})

	t.Run("failed to encrypt data", func(t *testing.T) {
		patch := func([]byte, []byte) ([]byte, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(aes.CTREncrypt, patch)
		defer pg.Unpatch()

		data, err := SaveCtrlCertPool(pool, testPassword)
		monkey.IsExistMonkeyError(t, err)
		require.Nil(t, data)
	})
}

func TestLoadCtrlCertPool(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		pool := testGenerateCertPool(t)
		data, err := SaveCtrlCertPool(pool, testPassword)
		require.NoError(t, err)

		pool = certpool.NewPool()
		err = LoadCtrlCertPool(pool, data, testPassword)
		require.NoError(t, err)

		fmt.Println(len(pool.GetPublicRootCACerts()))
		fmt.Println(len(pool.GetPublicClientCACerts()))
		fmt.Println(len(pool.GetPublicClientPairs()))
		fmt.Println(len(pool.GetPrivateRootCACerts()))
		fmt.Println(len(pool.GetPrivateClientCACerts()))
		fmt.Println(len(pool.GetPrivateClientPairs()))
	})

	t.Run("invalid cert pool file size", func(t *testing.T) {
		pool := certpool.NewPool()

		err := LoadCtrlCertPool(pool, nil, testPassword)
		require.Error(t, err)
	})

	t.Run("invalid mac", func(t *testing.T) {
		pool := certpool.NewPool()

		data := make([]byte, 4096)

		err := LoadCtrlCertPool(pool, data, testPassword)
		require.Error(t, err)
	})

	t.Run("invalid cipher data", func(t *testing.T) {
		patch := func([]byte, []byte) ([]byte, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(aes.CTRDecrypt, patch)
		defer pg.Unpatch()

		pool := certpool.NewPool()

		data := make([]byte, sha256.Size+aes.IVSize+8)
		aesKey := calculateAESKey(testPassword)
		hash := hmac.New(sha256.New, aesKey)
		hash.Write(data[sha256.Size:])
		copy(data, hash.Sum(nil))

		err := LoadCtrlCertPool(pool, data, testPassword)
		monkey.IsExistMonkeyError(t, err)
	})

	t.Run("invalid compressed data", func(t *testing.T) {
		pool := certpool.NewPool()

		data := make([]byte, sha256.Size+aes.IVSize+8)
		aesKey := calculateAESKey(testPassword)
		hash := hmac.New(sha256.New, aesKey)
		hash.Write(data[sha256.Size:])
		copy(data, hash.Sum(nil))

		err := LoadCtrlCertPool(pool, data, testPassword)
		require.Error(t, err)
	})

	pool := testGenerateCertPool(t)
	cpData, err := SaveCtrlCertPool(pool, testPassword)
	require.NoError(t, err)

	t.Run("failed to close deflate reader", func(t *testing.T) {
		reader := flate.NewReader(nil)
		patch := func(interface{}) error {
			return monkey.Error
		}
		pg := monkey.PatchInstanceMethod(reader, "Close", patch)
		defer pg.Unpatch()

		err := LoadCtrlCertPool(pool, cpData, testPassword)
		require.Error(t, err)
	})

	t.Run("failed to unmarshal", func(t *testing.T) {
		patch := func([]byte, interface{}) error {
			return monkey.Error
		}
		pg := monkey.Patch(msgpack.Unmarshal, patch)
		defer pg.Unpatch()

		err := LoadCtrlCertPool(pool, cpData, testPassword)
		monkey.IsExistMonkeyError(t, err)
	})
}
