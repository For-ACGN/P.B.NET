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

func TestNBCertPool_GetCertsFromPool(t *testing.T) {
	pair := testGenerateCert(t)
	c, k := pair.Encode()

	pool := certpool.NewPool()

	err := pool.AddPublicRootCACert(c)
	require.NoError(t, err)
	err = pool.AddPublicClientCACert(c)
	require.NoError(t, err)
	err = pool.AddPublicClientPair(c, k)
	require.NoError(t, err)
	err = pool.AddPrivateRootCAPair(c, k)
	require.NoError(t, err)
	err = pool.AddPrivateClientCAPair(c, k)
	require.NoError(t, err)
	err = pool.AddPrivateClientPair(c, k)
	require.NoError(t, err)

	cp := new(CertPool)
	cp.Load(pool)

	require.Len(t, cp.PublicRootCACerts, 1)
	require.Len(t, cp.PublicClientCACerts, 1)
	require.Len(t, cp.PublicClientPairs, 1)
	require.Len(t, cp.PrivateRootCACerts, 1)
	require.Len(t, cp.PrivateClientCACerts, 1)
	require.Len(t, cp.PrivateClientPairs, 1)
}

func TestNBCertPool_ToPool(t *testing.T) {
	cp := new(CertPool)

	t.Run("public root ca cert", func(t *testing.T) {
		pair := testGenerateCert(t)

		cp.PublicRootCACerts = [][]byte{pair.ASN1()}
		pool, err := cp.ToPool()
		require.NoError(t, err)

		certs := pool.GetPublicRootCACerts()
		require.Len(t, certs, 1)
		require.Equal(t, pair.ASN1(), certs[0].Raw)

		// already exists
		cp.PublicRootCACerts = append(cp.PublicRootCACerts, pair.ASN1())
		_, err = cp.ToPool()
		require.Error(t, err)

		cp.PublicRootCACerts = [][]byte{pair.ASN1()}
	})

	t.Run("public client ca cert", func(t *testing.T) {
		pair := testGenerateCert(t)

		cp.PublicClientCACerts = [][]byte{pair.ASN1()}
		pool, err := cp.ToPool()
		require.NoError(t, err)

		certs := pool.GetPublicClientCACerts()
		require.Len(t, certs, 1)
		require.Equal(t, pair.ASN1(), certs[0].Raw)

		// already exists
		cp.PublicClientCACerts = append(cp.PublicClientCACerts, pair.ASN1())
		_, err = cp.ToPool()
		require.Error(t, err)

		cp.PublicClientCACerts = [][]byte{pair.ASN1()}
	})

	t.Run("public client cert", func(t *testing.T) {
		pair := testGenerateCert(t)

		c, k := pair.Encode()
		cp.PublicClientPairs = []struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{
			{Cert: c, Key: k},
		}

		pool, err := cp.ToPool()
		require.NoError(t, err)

		certs := pool.GetPublicClientPairs()
		require.Len(t, certs, 1)
		dCert, dKey := certs[0].Encode()
		require.Equal(t, c, dCert)
		require.Equal(t, k, dKey)

		// already exists
		cp.PublicClientPairs = append(cp.PublicClientPairs, struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{
			Cert: c, Key: k,
		})
		_, err = cp.ToPool()
		require.Error(t, err)

		cp.PublicClientPairs = []struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{
			{Cert: c, Key: k},
		}
	})

	t.Run("private root ca cert", func(t *testing.T) {
		pair := testGenerateCert(t)

		cp.PrivateRootCACerts = [][]byte{pair.ASN1()}
		pool, err := cp.ToPool()
		require.NoError(t, err)

		certs := pool.GetPrivateRootCACerts()
		require.Len(t, certs, 1)
		require.Equal(t, pair.ASN1(), certs[0].Raw)

		// already exists
		cp.PrivateRootCACerts = append(cp.PrivateRootCACerts, pair.ASN1())
		_, err = cp.ToPool()
		require.Error(t, err)

		cp.PrivateRootCACerts = [][]byte{pair.ASN1()}
	})

	t.Run("private client ca cert", func(t *testing.T) {
		pair := testGenerateCert(t)

		cp.PrivateClientCACerts = [][]byte{pair.ASN1()}
		pool, err := cp.ToPool()
		require.NoError(t, err)

		certs := pool.GetPrivateClientCACerts()
		require.Len(t, certs, 1)
		require.Equal(t, pair.ASN1(), certs[0].Raw)

		// already exists
		cp.PrivateClientCACerts = append(cp.PrivateClientCACerts, pair.ASN1())
		_, err = cp.ToPool()
		require.Error(t, err)

		cp.PrivateClientCACerts = [][]byte{pair.ASN1()}
	})

	t.Run("private client cert", func(t *testing.T) {
		pair := testGenerateCert(t)

		c, k := pair.Encode()
		cp.PrivateClientPairs = []struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{
			{Cert: c, Key: k},
		}

		pool, err := cp.ToPool()
		require.NoError(t, err)

		certs := pool.GetPrivateClientPairs()
		require.Len(t, certs, 1)
		dCert, dKey := certs[0].Encode()
		require.Equal(t, c, dCert)
		require.Equal(t, k, dKey)

		// already exists
		cp.PrivateClientPairs = append(cp.PrivateClientPairs, struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{
			Cert: c, Key: k,
		})
		_, err = cp.ToPool()
		require.Error(t, err)

		cp.PrivateClientPairs = []struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{
			{Cert: c, Key: k},
		}
	})
}
