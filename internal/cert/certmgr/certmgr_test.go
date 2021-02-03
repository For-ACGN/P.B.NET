package certmgr

import (
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/cert"
	"project/internal/cert/certpool"
)

func testGenerateCert(t *testing.T) *cert.Pair {
	pair, err := cert.GenerateCA(nil)
	require.NoError(t, err)
	return pair
}

func TestManager_Load(t *testing.T) {
	pair := testGenerateCert(t)
	crt, key := pair.Encode()

	pool := certpool.NewPool()

	err := pool.AddPublicRootCACert(crt)
	require.NoError(t, err)
	err = pool.AddPublicClientCACert(crt)
	require.NoError(t, err)
	err = pool.AddPublicClientPair(crt, key)
	require.NoError(t, err)
	err = pool.AddPrivateRootCAPair(crt, key)
	require.NoError(t, err)
	err = pool.AddPrivateClientCAPair(crt, key)
	require.NoError(t, err)
	err = pool.AddPrivateClientPair(crt, key)
	require.NoError(t, err)

	mgr := new(Manager)
	mgr.Load(pool)

	require.Len(t, mgr.PublicRootCACerts, 1)
	require.Len(t, mgr.PublicClientCACerts, 1)
	require.Len(t, mgr.PublicClientPairs, 1)
	require.Len(t, mgr.PrivateRootCACerts, 1)
	require.Len(t, mgr.PrivateClientCACerts, 1)
	require.Len(t, mgr.PrivateClientPairs, 1)

	mgr.Clean()
}

func TestManager_ToCertPool(t *testing.T) {
	mgr := new(Manager)

	t.Run("public root ca cert", func(t *testing.T) {
		pair := testGenerateCert(t)

		mgr.PublicRootCACerts = [][]byte{pair.ASN1()}
		pool, err := mgr.ToCertPool()
		require.NoError(t, err)

		certs := pool.GetPublicRootCACerts()
		require.Len(t, certs, 1)
		require.Equal(t, pair.ASN1(), certs[0].Raw)

		// already exists
		mgr.PublicRootCACerts = append(mgr.PublicRootCACerts, pair.ASN1())
		_, err = mgr.ToCertPool()
		require.Error(t, err)

		mgr.PublicRootCACerts = [][]byte{pair.ASN1()}
	})

	t.Run("public client ca cert", func(t *testing.T) {
		pair := testGenerateCert(t)

		mgr.PublicClientCACerts = [][]byte{pair.ASN1()}
		pool, err := mgr.ToCertPool()
		require.NoError(t, err)

		certs := pool.GetPublicClientCACerts()
		require.Len(t, certs, 1)
		require.Equal(t, pair.ASN1(), certs[0].Raw)

		// already exists
		mgr.PublicClientCACerts = append(mgr.PublicClientCACerts, pair.ASN1())
		_, err = mgr.ToCertPool()
		require.Error(t, err)

		mgr.PublicClientCACerts = [][]byte{pair.ASN1()}
	})

	t.Run("public client pair", func(t *testing.T) {
		pair := testGenerateCert(t)
		crt, key := pair.Encode()
		mgr.PublicClientPairs = []struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{
			{Cert: crt, Key: key},
		}

		pool, err := mgr.ToCertPool()
		require.NoError(t, err)

		certs := pool.GetPublicClientPairs()
		require.Len(t, certs, 1)
		dCert, dKey := certs[0].Encode()
		require.Equal(t, crt, dCert)
		require.Equal(t, key, dKey)

		// already exists
		mgr.PublicClientPairs = append(mgr.PublicClientPairs, struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{
			Cert: crt, Key: key,
		})
		_, err = mgr.ToCertPool()
		require.Error(t, err)

		mgr.PublicClientPairs = []struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{
			{Cert: crt, Key: key},
		}
	})

	t.Run("private root ca cert", func(t *testing.T) {
		pair := testGenerateCert(t)

		mgr.PrivateRootCACerts = [][]byte{pair.ASN1()}
		pool, err := mgr.ToCertPool()
		require.NoError(t, err)

		certs := pool.GetPrivateRootCACerts()
		require.Len(t, certs, 1)
		require.Equal(t, pair.ASN1(), certs[0].Raw)

		// already exists
		mgr.PrivateRootCACerts = append(mgr.PrivateRootCACerts, pair.ASN1())
		_, err = mgr.ToCertPool()
		require.Error(t, err)

		mgr.PrivateRootCACerts = [][]byte{pair.ASN1()}
	})

	t.Run("private client ca cert", func(t *testing.T) {
		pair := testGenerateCert(t)

		mgr.PrivateClientCACerts = [][]byte{pair.ASN1()}
		pool, err := mgr.ToCertPool()
		require.NoError(t, err)

		certs := pool.GetPrivateClientCACerts()
		require.Len(t, certs, 1)
		require.Equal(t, pair.ASN1(), certs[0].Raw)

		// already exists
		mgr.PrivateClientCACerts = append(mgr.PrivateClientCACerts, pair.ASN1())
		_, err = mgr.ToCertPool()
		require.Error(t, err)

		mgr.PrivateClientCACerts = [][]byte{pair.ASN1()}
	})

	t.Run("private client pair", func(t *testing.T) {
		pair := testGenerateCert(t)
		crt, key := pair.Encode()
		mgr.PrivateClientPairs = []struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{
			{Cert: crt, Key: key},
		}

		pool, err := mgr.ToCertPool()
		require.NoError(t, err)

		certs := pool.GetPrivateClientPairs()
		require.Len(t, certs, 1)
		dCert, dKey := certs[0].Encode()
		require.Equal(t, crt, dCert)
		require.Equal(t, key, dKey)

		// already exists
		mgr.PrivateClientPairs = append(mgr.PrivateClientPairs, struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{
			Cert: crt, Key: key,
		})
		_, err = mgr.ToCertPool()
		require.Error(t, err)

		mgr.PrivateClientPairs = []struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{
			{Cert: crt, Key: key},
		}
	})
}
