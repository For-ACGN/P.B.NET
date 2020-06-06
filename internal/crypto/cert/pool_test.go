package cert

import (
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/crypto/cert/certpool"
	"project/internal/patch/monkey"
	"project/internal/security"
	"project/internal/testsuite"
)

func TestPair_ToPair(t *testing.T) {
	defer testsuite.DeferForPanic(t)

	sb := security.NewBytes(make([]byte, 1024))
	pair := pair{PrivateKey: sb}

	pair.ToPair()
}

func testGeneratePair(t *testing.T) *Pair {
	pair, err := GenerateCA(nil)
	require.NoError(t, err)
	return pair
}

func TestLoadPair(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		pair := testGeneratePair(t)

		p, err := loadPair(pair.Encode())
		require.NoError(t, err)
		require.NotNil(t, p)
	})

	t.Run("no certificate", func(t *testing.T) {
		_, err := loadPair(nil, nil)
		require.Error(t, err)
	})

	t.Run("no private key", func(t *testing.T) {
		_, err := loadPair(make([]byte, 1024), nil)
		require.Error(t, err)
	})

	t.Run("invalid certificate", func(t *testing.T) {
		padding := make([]byte, 1024)
		_, err := loadPair(padding, padding)
		require.Error(t, err)
	})

	t.Run("invalid private key", func(t *testing.T) {
		pair := testGeneratePair(t)
		cert, _ := pair.Encode()

		_, err := loadPair(cert, make([]byte, 1024))
		require.Error(t, err)
	})

	t.Run("mismatched private key", func(t *testing.T) {
		pair1 := testGeneratePair(t)
		cert := pair1.ASN1()

		pair2 := testGeneratePair(t)
		_, key := pair2.Encode()

		_, err := loadPair(cert, key)
		require.Error(t, err)
	})

	t.Run("failed to marshal PKCS8 private key", func(t *testing.T) {
		pair := testGeneratePair(t)
		cert, key := pair.Encode()

		patch := func(interface{}) ([]byte, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(x509.MarshalPKCS8PrivateKey, patch)
		defer pg.Unpatch()

		_, err := loadPair(cert, key)
		monkey.IsMonkeyError(t, err)
	})
}

func TestLoadCertToPair(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		pair := testGeneratePair(t)

		p, err := loadCertToPair(pair.ASN1())
		require.NoError(t, err)
		require.NotNil(t, p)
		require.Nil(t, p.PrivateKey)
	})

	t.Run("no certificate", func(t *testing.T) {
		_, err := loadCertToPair(nil)
		require.Error(t, err)
	})

	t.Run("invalid certificate", func(t *testing.T) {
		_, err := loadCertToPair(make([]byte, 1024))
		require.Error(t, err)
	})
}

func TestPool_AddPublicRootCACert(t *testing.T) {
	pair := testGeneratePair(t)

	t.Run("ok", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPublicRootCACert(pair.Certificate.Raw)
		require.NoError(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("exist", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPublicRootCACert(pair.Certificate.Raw)
		require.NoError(t, err)
		err = pool.AddPublicRootCACert(pair.Certificate.Raw)
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid certificate", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPublicRootCACert(nil)
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_AddPublicClientCACert(t *testing.T) {
	pair := testGeneratePair(t)

	t.Run("ok", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPublicClientCACert(pair.Certificate.Raw)
		require.NoError(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("exist", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPublicClientCACert(pair.Certificate.Raw)
		require.NoError(t, err)
		err = pool.AddPublicClientCACert(pair.Certificate.Raw)
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid certificate", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPublicClientCACert(nil)
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_AddPublicClientPair(t *testing.T) {
	pair := testGeneratePair(t)

	t.Run("ok", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPublicClientPair(pair.Encode())
		require.NoError(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("exist", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPublicClientPair(pair.Encode())
		require.NoError(t, err)
		err = pool.AddPublicClientPair(pair.Encode())
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid pair", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPublicClientPair(nil, nil)
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_AddPrivateRootCAPair(t *testing.T) {
	pair := testGeneratePair(t)

	t.Run("ok", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPrivateRootCAPair(pair.Encode())
		require.NoError(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("exist", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPrivateRootCAPair(pair.Encode())
		require.NoError(t, err)
		err = pool.AddPrivateRootCAPair(pair.Encode())
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid pair", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPrivateRootCAPair(nil, nil)
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_AddPrivateRootCACert(t *testing.T) {
	pair := testGeneratePair(t)

	t.Run("ok", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPrivateRootCACert(pair.Certificate.Raw)
		require.NoError(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("exist", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPrivateRootCACert(pair.Certificate.Raw)
		require.NoError(t, err)
		err = pool.AddPrivateRootCACert(pair.Certificate.Raw)
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid certificate", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPrivateRootCACert(nil)
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_AddPrivateClientCAPair(t *testing.T) {
	pair := testGeneratePair(t)

	t.Run("ok", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPrivateClientCAPair(pair.Encode())
		require.NoError(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("exist", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPrivateClientCAPair(pair.Encode())
		require.NoError(t, err)
		err = pool.AddPrivateClientCAPair(pair.Encode())
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid pair", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPrivateClientCAPair(nil, nil)
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_AddPrivateClientCACert(t *testing.T) {
	pair := testGeneratePair(t)

	t.Run("ok", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPrivateClientCACert(pair.Certificate.Raw)
		require.NoError(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("exist", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPrivateClientCACert(pair.Certificate.Raw)
		require.NoError(t, err)
		err = pool.AddPrivateClientCACert(pair.Certificate.Raw)
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid certificate", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPrivateClientCACert(nil)
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_AddPrivateClientPair(t *testing.T) {
	pair := testGeneratePair(t)

	t.Run("ok", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPrivateClientPair(pair.Encode())
		require.NoError(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("exist", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPrivateClientPair(pair.Encode())
		require.NoError(t, err)
		err = pool.AddPrivateClientPair(pair.Encode())
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid pair", func(t *testing.T) {
		pool := NewPool()

		err := pool.AddPrivateClientPair(nil, nil)
		require.Error(t, err)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_DeletePublicRootCACert(t *testing.T) {
	pair := testGeneratePair(t)

	t.Run("ok", func(t *testing.T) {
		pool := NewPool()
		err := pool.AddPublicRootCACert(pair.Certificate.Raw)
		require.NoError(t, err)

		err = pool.DeletePublicRootCACert(0)
		require.NoError(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("doesn't exist", func(t *testing.T) {
		pool := NewPool()
		err := pool.AddPublicRootCACert(pair.Certificate.Raw)
		require.NoError(t, err)
		err = pool.DeletePublicRootCACert(0)
		require.NoError(t, err)

		err = pool.DeletePublicRootCACert(0)
		require.Error(t, err)
		t.Log(err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid id", func(t *testing.T) {
		pool := NewPool()

		for _, id := range []int{
			-1, 0, 1,
		} {
			err := pool.DeletePublicRootCACert(id)
			require.Error(t, err)
			t.Log(err)
		}

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_DeletePublicClientCACert(t *testing.T) {
	pair := testGeneratePair(t)

	t.Run("ok", func(t *testing.T) {
		pool := NewPool()
		err := pool.AddPublicClientCACert(pair.Certificate.Raw)
		require.NoError(t, err)

		err = pool.DeletePublicClientCACert(0)
		require.NoError(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("doesn't exist", func(t *testing.T) {
		pool := NewPool()
		err := pool.AddPublicClientCACert(pair.Certificate.Raw)
		require.NoError(t, err)
		err = pool.DeletePublicClientCACert(0)
		require.NoError(t, err)

		err = pool.DeletePublicClientCACert(0)
		require.Error(t, err)
		t.Log(err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid id", func(t *testing.T) {
		pool := NewPool()

		for _, id := range []int{
			-1, 0, 1,
		} {
			err := pool.DeletePublicClientCACert(id)
			require.Error(t, err)
			t.Log(err)
		}

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_DeletePublicClientCert(t *testing.T) {
	pair := testGeneratePair(t)

	t.Run("ok", func(t *testing.T) {
		pool := NewPool()
		err := pool.AddPublicClientPair(pair.Encode())
		require.NoError(t, err)

		err = pool.DeletePublicClientCert(0)
		require.NoError(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("doesn't exist", func(t *testing.T) {
		pool := NewPool()
		err := pool.AddPublicClientPair(pair.Encode())
		require.NoError(t, err)
		err = pool.DeletePublicClientCert(0)
		require.NoError(t, err)

		err = pool.DeletePublicClientCert(0)
		require.Error(t, err)
		t.Log(err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid id", func(t *testing.T) {
		pool := NewPool()

		for _, id := range []int{
			-1, 0, 1,
		} {
			err := pool.DeletePublicClientCert(id)
			require.Error(t, err)
			t.Log(err)
		}

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_DeletePrivateRootCACert(t *testing.T) {
	pair := testGeneratePair(t)

	t.Run("ok", func(t *testing.T) {
		pool := NewPool()
		err := pool.AddPrivateRootCACert(pair.Certificate.Raw)
		require.NoError(t, err)

		err = pool.DeletePrivateRootCACert(0)
		require.NoError(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("doesn't exist", func(t *testing.T) {
		pool := NewPool()
		err := pool.AddPrivateRootCACert(pair.Certificate.Raw)
		require.NoError(t, err)
		err = pool.DeletePrivateRootCACert(0)
		require.NoError(t, err)

		err = pool.DeletePrivateRootCACert(0)
		require.Error(t, err)
		t.Log(err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid id", func(t *testing.T) {
		pool := NewPool()

		for _, id := range []int{
			-1, 0, 1,
		} {
			err := pool.DeletePrivateRootCACert(id)
			require.Error(t, err)
			t.Log(err)
		}

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_DeletePrivateClientCACert(t *testing.T) {
	pair := testGeneratePair(t)

	t.Run("ok", func(t *testing.T) {
		pool := NewPool()
		err := pool.AddPrivateClientCACert(pair.Certificate.Raw)
		require.NoError(t, err)

		err = pool.DeletePrivateClientCACert(0)
		require.NoError(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("doesn't exist", func(t *testing.T) {
		pool := NewPool()
		err := pool.AddPrivateClientCACert(pair.Certificate.Raw)
		require.NoError(t, err)
		err = pool.DeletePrivateClientCACert(0)
		require.NoError(t, err)

		err = pool.DeletePrivateClientCACert(0)
		require.Error(t, err)
		t.Log(err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid id", func(t *testing.T) {
		pool := NewPool()

		for _, id := range []int{
			-1, 0, 1,
		} {
			err := pool.DeletePrivateClientCACert(id)
			require.Error(t, err)
			t.Log(err)
		}

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_DeletePrivateClientPair(t *testing.T) {
	pair := testGeneratePair(t)

	t.Run("ok", func(t *testing.T) {
		pool := NewPool()
		err := pool.AddPrivateClientPair(pair.Encode())
		require.NoError(t, err)

		err = pool.DeletePrivateClientCert(0)
		require.NoError(t, err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("doesn't exist", func(t *testing.T) {
		pool := NewPool()
		err := pool.AddPrivateClientPair(pair.Encode())
		require.NoError(t, err)
		err = pool.DeletePrivateClientCert(0)
		require.NoError(t, err)

		err = pool.DeletePrivateClientCert(0)
		require.Error(t, err)
		t.Log(err)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid id", func(t *testing.T) {
		pool := NewPool()

		for _, id := range []int{
			-1, 0, 1,
		} {
			err := pool.DeletePrivateClientCert(id)
			require.Error(t, err)
			t.Log(err)
		}

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_GetPublicRootCACert(t *testing.T) {
	pair := testGeneratePair(t)
	pool := NewPool()
	err := pool.AddPublicRootCACert(pair.Certificate.Raw)
	require.NoError(t, err)

	certs := pool.GetPublicRootCACerts()
	require.True(t, certs[0].Equal(pair.Certificate))

	testsuite.IsDestroyed(t, pool)
}

func TestPool_GetPublicClientCACert(t *testing.T) {
	pair := testGeneratePair(t)
	pool := NewPool()
	err := pool.AddPublicClientCACert(pair.Certificate.Raw)
	require.NoError(t, err)

	certs := pool.GetPublicClientCACerts()
	require.True(t, certs[0].Equal(pair.Certificate))

	testsuite.IsDestroyed(t, pool)
}

func TestPool_GetPublicClientPair(t *testing.T) {
	pair := testGeneratePair(t)
	pool := NewPool()
	err := pool.AddPublicClientPair(pair.Encode())
	require.NoError(t, err)

	pairs := pool.GetPublicClientPairs()
	require.Equal(t, pair, pairs[0])

	testsuite.IsDestroyed(t, pool)
}

func TestPool_GetPrivateRootCAPair(t *testing.T) {
	pair := testGeneratePair(t)
	pool := NewPool()
	err := pool.AddPrivateRootCAPair(pair.Encode())
	require.NoError(t, err)

	pairs := pool.GetPrivateRootCAPairs()
	require.Equal(t, pair, pairs[0])

	testsuite.IsDestroyed(t, pool)
}

func TestPool_GetPrivateRootCACert(t *testing.T) {
	pair := testGeneratePair(t)
	pool := NewPool()
	err := pool.AddPrivateRootCAPair(pair.Encode())
	require.NoError(t, err)

	certs := pool.GetPrivateRootCACerts()
	require.True(t, certs[0].Equal(pair.Certificate))

	testsuite.IsDestroyed(t, pool)
}

func TestPool_GetPrivateClientCAPair(t *testing.T) {
	pair := testGeneratePair(t)
	pool := NewPool()
	err := pool.AddPrivateClientCAPair(pair.Encode())
	require.NoError(t, err)

	pairs := pool.GetPrivateClientCAPairs()
	require.Equal(t, pair, pairs[0])

	testsuite.IsDestroyed(t, pool)
}

func TestPool_GetPrivateClientCACert(t *testing.T) {
	pair := testGeneratePair(t)
	pool := NewPool()
	err := pool.AddPrivateClientCAPair(pair.Encode())
	require.NoError(t, err)

	certs := pool.GetPrivateClientCACerts()
	require.True(t, certs[0].Equal(pair.Certificate))

	testsuite.IsDestroyed(t, pool)
}

func TestPool_GetPrivateClientPair(t *testing.T) {
	pair := testGeneratePair(t)
	pool := NewPool()
	err := pool.AddPrivateClientPair(pair.Encode())
	require.NoError(t, err)

	pairs := pool.GetPrivateClientPairs()
	require.Equal(t, pair, pairs[0])

	testsuite.IsDestroyed(t, pool)
}

func TestPool_ExportPublicRootCACert(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		pair := testGeneratePair(t)
		pool := NewPool()
		err := pool.AddPublicRootCACert(pair.Certificate.Raw)
		require.NoError(t, err)

		cert, err := pool.ExportPublicRootCACert(0)
		require.NoError(t, err)

		c, _ := pair.EncodeToPEM()
		require.Equal(t, c, cert)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid id", func(t *testing.T) {
		pool := NewPool()

		for _, id := range []int{
			-1, 0, 1,
		} {
			cert, err := pool.ExportPublicRootCACert(id)
			require.Error(t, err)
			require.Nil(t, cert)
			t.Log(err)
		}

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_ExportPublicClientCACert(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		pair := testGeneratePair(t)
		pool := NewPool()
		err := pool.AddPublicClientCACert(pair.Certificate.Raw)
		require.NoError(t, err)

		cert, err := pool.ExportPublicClientCACert(0)
		require.NoError(t, err)

		c, _ := pair.EncodeToPEM()
		require.Equal(t, c, cert)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid id", func(t *testing.T) {
		pool := NewPool()

		for _, id := range []int{
			-1, 0, 1,
		} {
			cert, err := pool.ExportPublicClientCACert(id)
			require.Error(t, err)
			require.Nil(t, cert)
			t.Log(err)
		}

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_ExportPublicClientCert(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		pair := testGeneratePair(t)
		pool := NewPool()
		err := pool.AddPublicClientPair(pair.Encode())
		require.NoError(t, err)

		cert, key, err := pool.ExportPublicClientPair(0)
		require.NoError(t, err)

		c, k := pair.EncodeToPEM()
		require.Equal(t, c, cert)
		require.Equal(t, k, key)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid id", func(t *testing.T) {
		pool := NewPool()

		for _, id := range []int{
			-1, 0, 1,
		} {
			cert, key, err := pool.ExportPublicClientPair(id)
			require.Error(t, err)
			require.Nil(t, cert)
			require.Nil(t, key)
			t.Log(err)
		}

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_ExportPrivateRootCACert(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		pair := testGeneratePair(t)
		pool := NewPool()
		err := pool.AddPrivateRootCAPair(pair.Encode())
		require.NoError(t, err)

		cert, key, err := pool.ExportPrivateRootCAPair(0)
		require.NoError(t, err)

		c, k := pair.EncodeToPEM()
		require.Equal(t, c, cert)
		require.Equal(t, k, key)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid id", func(t *testing.T) {
		pool := NewPool()

		for _, id := range []int{
			-1, 0, 1,
		} {
			cert, key, err := pool.ExportPrivateRootCAPair(id)
			require.Error(t, err)
			require.Nil(t, cert)
			require.Nil(t, key)
			t.Log(err)
		}

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_ExportPrivateClientCACert(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		pair := testGeneratePair(t)
		pool := NewPool()
		err := pool.AddPrivateClientCAPair(pair.Encode())
		require.NoError(t, err)

		cert, key, err := pool.ExportPrivateClientCAPair(0)
		require.NoError(t, err)

		c, k := pair.EncodeToPEM()
		require.Equal(t, c, cert)
		require.Equal(t, k, key)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid id", func(t *testing.T) {
		pool := NewPool()

		for _, id := range []int{
			-1, 0, 1,
		} {
			cert, key, err := pool.ExportPrivateClientCAPair(id)
			require.Error(t, err)
			require.Nil(t, cert)
			require.Nil(t, key)
			t.Log(err)
		}

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_ExportPrivateClientPair(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		pair := testGeneratePair(t)
		pool := NewPool()
		err := pool.AddPrivateClientPair(pair.Encode())
		require.NoError(t, err)

		cert, key, err := pool.ExportPrivateClientPair(0)
		require.NoError(t, err)

		c, k := pair.EncodeToPEM()
		require.Equal(t, c, cert)
		require.Equal(t, k, key)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("invalid id", func(t *testing.T) {
		pool := NewPool()

		for _, id := range []int{
			-1, 0, 1,
		} {
			cert, key, err := pool.ExportPrivateClientPair(id)
			require.Error(t, err)
			require.Nil(t, cert)
			require.Nil(t, key)
			t.Log(err)
		}

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_AddPublicRootCA_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		add1 := func() {
			err := pool.AddPublicRootCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.AddPublicRootCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.AddPublicRootCACert(nil)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.GetPublicRootCACerts()
			require.Len(t, certs, 2)

			err := pool.DeletePublicRootCACert(0)
			require.NoError(t, err)
			err = pool.DeletePublicRootCACert(0)
			require.NoError(t, err)

			certs = pool.GetPublicRootCACerts()
			require.Len(t, certs, 0)
		}
		testsuite.RunParallel(100, nil, cleanup, add1, add2, add3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()
		}
		add1 := func() {
			err := pool.AddPublicRootCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.AddPublicRootCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.AddPublicRootCACert(nil)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.GetPublicRootCACerts()
			require.Len(t, certs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, add1, add2, add3)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_AddPublicClientCA_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		add1 := func() {
			err := pool.AddPublicClientCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.AddPublicClientCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.AddPublicClientCACert(nil)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.GetPublicClientCACerts()
			require.Len(t, certs, 2)

			err := pool.DeletePublicClientCACert(0)
			require.NoError(t, err)
			err = pool.DeletePublicClientCACert(0)
			require.NoError(t, err)

			certs = pool.GetPublicClientCACerts()
			require.Len(t, certs, 0)
		}
		testsuite.RunParallel(100, nil, cleanup, add1, add2, add3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()
		}
		add1 := func() {
			err := pool.AddPublicClientCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.AddPublicClientCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.AddPublicClientCACert(nil)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.GetPublicClientCACerts()
			require.Len(t, certs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, add1, add2, add3)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_PublicRootCA_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("add", func(t *testing.T) {

	})

	t.Run("delete", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)

		init := func() {
			pool = NewPool()

			err := pool.AddPublicRootCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPublicRootCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		delete1 := func() {
			err := pool.DeletePublicRootCACert(0)
			require.NoError(t, err)
		}
		delete2 := func() {
			err := pool.DeletePublicRootCACert(0)
			require.NoError(t, err)
		}
		delete3 := func() {
			err := pool.DeletePublicRootCACert(3)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.GetPublicRootCACerts()
			require.Len(t, certs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("get", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)

		init := func() {
			pool = NewPool()

			err := pool.AddPublicRootCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPublicRootCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		get := func() {
			certs := pool.GetPublicRootCACerts()
			require.Len(t, certs, 2)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("export", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, _ := pair1.EncodeToPEM()
		cert2, _ := pair2.EncodeToPEM()

		init := func() {
			pool = NewPool()

			err := pool.AddPublicRootCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPublicRootCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		export1 := func() {
			cert, err := pool.ExportPublicRootCACert(0)
			require.NoError(t, err)
			require.Equal(t, cert1, cert)
		}
		export2 := func() {
			cert, err := pool.ExportPublicRootCACert(1)
			require.NoError(t, err)
			require.Equal(t, cert2, cert)
		}
		export3 := func() {
			cert, err := pool.ExportPublicRootCACert(2)
			require.Error(t, err)
			require.Nil(t, cert)
		}
		cleanup := func() {
			certs := pool.GetPublicRootCACerts()
			require.Len(t, certs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_PublicClientCA_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("add", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)

		init := func() {
			pool = NewPool()
		}
		add1 := func() {
			err := pool.AddPublicClientCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.AddPublicClientCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.AddPublicClientCACert(nil)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.GetPublicClientCACerts()
			require.Len(t, certs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, add1, add2, add3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("delete", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)

		init := func() {
			pool = NewPool()

			err := pool.AddPublicClientCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPublicClientCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		delete1 := func() {
			err := pool.DeletePublicClientCACert(0)
			require.NoError(t, err)
		}
		delete2 := func() {
			err := pool.DeletePublicClientCACert(0)
			require.NoError(t, err)
		}
		delete3 := func() {
			err := pool.DeletePublicClientCACert(3)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.GetPublicClientCACerts()
			require.Len(t, certs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("get", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)

		init := func() {
			pool = NewPool()

			err := pool.AddPublicClientCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPublicClientCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		get := func() {
			certs := pool.GetPublicClientCACerts()
			require.Len(t, certs, 2)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("export", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, _ := pair1.EncodeToPEM()
		cert2, _ := pair2.EncodeToPEM()

		init := func() {
			pool = NewPool()

			err := pool.AddPublicClientCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPublicClientCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		export1 := func() {
			cert, err := pool.ExportPublicClientCACert(0)
			require.NoError(t, err)
			require.Equal(t, cert1, cert)
		}
		export2 := func() {
			cert, err := pool.ExportPublicClientCACert(1)
			require.NoError(t, err)
			require.Equal(t, cert2, cert)
		}
		export3 := func() {
			cert, err := pool.ExportPublicClientCACert(2)
			require.Error(t, err)
			require.Nil(t, cert)
		}
		cleanup := func() {
			certs := pool.GetPublicClientCACerts()
			require.Len(t, certs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_PublicClient_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("add", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()

		init := func() {
			pool = NewPool()
		}
		add1 := func() {
			err := pool.AddPublicClientPair(cert1, key1)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.AddPublicClientPair(cert2, key2)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.AddPublicClientPair(nil, nil)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPublicClientPairs()
			require.Len(t, pairs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, add1, add2, add3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("delete", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()

		init := func() {
			pool = NewPool()

			err := pool.AddPublicClientPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPublicClientPair(cert2, key2)
			require.NoError(t, err)
		}
		delete1 := func() {
			err := pool.DeletePublicClientCert(0)
			require.NoError(t, err)
		}
		delete2 := func() {
			err := pool.DeletePublicClientCert(0)
			require.NoError(t, err)
		}
		delete3 := func() {
			err := pool.DeletePublicClientCert(3)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPublicClientPairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("get", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()

		init := func() {
			pool = NewPool()

			err := pool.AddPublicClientPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPublicClientPair(cert2, key2)
			require.NoError(t, err)
		}
		get := func() {
			pairs := pool.GetPublicClientPairs()
			require.Len(t, pairs, 2)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("export", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()
		cert1PEM, key1PEM := pair1.EncodeToPEM()
		cert2PEM, key2PEM := pair2.EncodeToPEM()

		init := func() {
			pool = NewPool()

			err := pool.AddPublicClientPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPublicClientPair(cert2, key2)
			require.NoError(t, err)
		}
		export1 := func() {
			cert, key, err := pool.ExportPublicClientPair(0)
			require.NoError(t, err)
			require.Equal(t, cert1PEM, cert)
			require.Equal(t, key1PEM, key)
		}
		export2 := func() {
			cert, key, err := pool.ExportPublicClientPair(1)
			require.NoError(t, err)
			require.Equal(t, cert2PEM, cert)
			require.Equal(t, key2PEM, key)
		}
		export3 := func() {
			cert, key, err := pool.ExportPublicClientPair(2)
			require.Error(t, err)
			require.Nil(t, cert)
			require.Nil(t, key)
		}
		cleanup := func() {
			pairs := pool.GetPublicClientPairs()
			require.Len(t, pairs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_PrivateRootCA_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("add pair", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()

		init := func() {
			pool = NewPool()
		}
		add1 := func() {
			err := pool.AddPrivateRootCAPair(cert1, key1)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.AddPrivateRootCAPair(cert2, key2)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.AddPrivateRootCAPair(nil, nil)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPrivateRootCAPairs()
			require.Len(t, pairs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, add1, add2, add3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("add certificate", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, _ := pair1.Encode()
		cert2, _ := pair2.Encode()

		init := func() {
			pool = NewPool()
		}
		add1 := func() {
			err := pool.AddPrivateRootCACert(cert1)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.AddPrivateRootCACert(cert2)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.AddPrivateRootCACert(nil)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.GetPrivateRootCACerts()
			require.Len(t, certs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, add1, add2, add3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("delete", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateRootCAPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPrivateRootCAPair(cert2, key2)
			require.NoError(t, err)
		}
		delete1 := func() {
			err := pool.DeletePrivateRootCACert(0)
			require.NoError(t, err)
		}
		delete2 := func() {
			err := pool.DeletePrivateRootCACert(0)
			require.NoError(t, err)
		}
		delete3 := func() {
			err := pool.DeletePrivateRootCACert(3)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPrivateRootCAPairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("get pair", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateRootCAPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPrivateRootCAPair(cert2, key2)
			require.NoError(t, err)
		}
		get := func() {
			pairs := pool.GetPrivateRootCAPairs()
			require.Len(t, pairs, 2)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("get certificate", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, _ := pair1.Encode()
		cert2, _ := pair2.Encode()

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateRootCACert(cert1)
			require.NoError(t, err)
			err = pool.AddPrivateRootCACert(cert2)
			require.NoError(t, err)
		}
		get := func() {
			certs := pool.GetPrivateRootCACerts()
			require.Len(t, certs, 2)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("export", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()
		cert1PEM, key1PEM := pair1.EncodeToPEM()
		cert2PEM, key2PEM := pair2.EncodeToPEM()

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateRootCAPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPrivateRootCAPair(cert2, key2)
			require.NoError(t, err)
		}
		export1 := func() {
			cert, key, err := pool.ExportPrivateRootCAPair(0)
			require.NoError(t, err)
			require.Equal(t, cert1PEM, cert)
			require.Equal(t, key1PEM, key)
		}
		export2 := func() {
			cert, key, err := pool.ExportPrivateRootCAPair(1)
			require.NoError(t, err)
			require.Equal(t, cert2PEM, cert)
			require.Equal(t, key2PEM, key)
		}
		export3 := func() {
			cert, key, err := pool.ExportPrivateRootCAPair(2)
			require.Error(t, err)
			require.Nil(t, cert)
			require.Nil(t, key)
		}
		cleanup := func() {
			pairs := pool.GetPrivateRootCAPairs()
			require.Len(t, pairs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_PrivateClientCA_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("add pair", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()

		init := func() {
			pool = NewPool()
		}
		add1 := func() {
			err := pool.AddPrivateClientCAPair(cert1, key1)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.AddPrivateClientCAPair(cert2, key2)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.AddPrivateClientCAPair(nil, nil)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPrivateClientCAPairs()
			require.Len(t, pairs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, add1, add2, add3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("add certificate", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, _ := pair1.Encode()
		cert2, _ := pair2.Encode()

		init := func() {
			pool = NewPool()
		}
		add1 := func() {
			err := pool.AddPrivateClientCACert(cert1)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.AddPrivateClientCACert(cert2)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.AddPrivateClientCACert(nil)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.GetPrivateClientCACerts()
			require.Len(t, certs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, add1, add2, add3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("delete", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateClientCAPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPrivateClientCAPair(cert2, key2)
			require.NoError(t, err)
		}
		delete1 := func() {
			err := pool.DeletePrivateClientCACert(0)
			require.NoError(t, err)
		}
		delete2 := func() {
			err := pool.DeletePrivateClientCACert(0)
			require.NoError(t, err)
		}
		delete3 := func() {
			err := pool.DeletePrivateClientCACert(3)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPrivateClientCACerts()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("get pair", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateClientCAPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPrivateClientCAPair(cert2, key2)
			require.NoError(t, err)
		}
		get := func() {
			pairs := pool.GetPrivateClientCACerts()
			require.Len(t, pairs, 2)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("get certificate", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, _ := pair1.Encode()
		cert2, _ := pair2.Encode()

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateClientCACert(cert1)
			require.NoError(t, err)
			err = pool.AddPrivateClientCACert(cert2)
			require.NoError(t, err)
		}
		get := func() {
			certs := pool.GetPrivateClientCACerts()
			require.Len(t, certs, 2)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("export", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()
		cert1PEM, key1PEM := pair1.EncodeToPEM()
		cert2PEM, key2PEM := pair2.EncodeToPEM()

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateClientCAPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPrivateClientCAPair(cert2, key2)
			require.NoError(t, err)
		}
		export1 := func() {
			cert, key, err := pool.ExportPrivateClientCAPair(0)
			require.NoError(t, err)
			require.Equal(t, cert1PEM, cert)
			require.Equal(t, key1PEM, key)
		}
		export2 := func() {
			cert, key, err := pool.ExportPrivateClientCAPair(1)
			require.NoError(t, err)
			require.Equal(t, cert2PEM, cert)
			require.Equal(t, key2PEM, key)
		}
		export3 := func() {
			cert, key, err := pool.ExportPrivateClientCAPair(2)
			require.Error(t, err)
			require.Nil(t, cert)
			require.Nil(t, key)
		}
		cleanup := func() {
			pairs := pool.GetPrivateClientCAPairs()
			require.Len(t, pairs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_PrivateClient_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("add", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()

		init := func() {
			pool = NewPool()
		}
		add1 := func() {
			err := pool.AddPrivateClientPair(cert1, key1)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.AddPrivateClientPair(cert2, key2)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.AddPrivateClientPair(nil, nil)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPrivateClientPairs()
			require.Len(t, pairs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, add1, add2, add3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("delete", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateClientPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPrivateClientPair(cert2, key2)
			require.NoError(t, err)
		}
		delete1 := func() {
			err := pool.DeletePrivateClientCert(0)
			require.NoError(t, err)
		}
		delete2 := func() {
			err := pool.DeletePrivateClientCert(0)
			require.NoError(t, err)
		}
		delete3 := func() {
			err := pool.DeletePrivateClientCert(3)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPrivateClientPairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("get", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateClientPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPrivateClientPair(cert2, key2)
			require.NoError(t, err)
		}
		get := func() {
			pairs := pool.GetPrivateClientPairs()
			require.Len(t, pairs, 2)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("export", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()
		cert1PEM, key1PEM := pair1.EncodeToPEM()
		cert2PEM, key2PEM := pair2.EncodeToPEM()

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateClientPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPrivateClientPair(cert2, key2)
			require.NoError(t, err)
		}
		export1 := func() {
			cert, key, err := pool.ExportPrivateClientPair(0)
			require.NoError(t, err)
			require.Equal(t, cert1PEM, cert)
			require.Equal(t, key1PEM, key)
		}
		export2 := func() {
			cert, key, err := pool.ExportPrivateClientPair(1)
			require.NoError(t, err)
			require.Equal(t, cert2PEM, cert)
			require.Equal(t, key2PEM, key)
		}
		export3 := func() {
			cert, key, err := pool.ExportPrivateClientPair(2)
			require.Error(t, err)
			require.Nil(t, cert)
			require.Nil(t, key)
		}
		cleanup := func() {
			pairs := pool.GetPrivateClientPairs()
			require.Len(t, pairs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestPool_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("pair only (controller)", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()

		init := func() {
			pool = NewPool()
		}
		fns := []func(){
			// add
			func() { _ = pool.AddPublicRootCACert(pair1.Certificate.Raw) },
			func() { _ = pool.AddPublicRootCACert(pair2.Certificate.Raw) },
			func() { _ = pool.AddPublicClientCACert(pair1.Certificate.Raw) },
			func() { _ = pool.AddPublicClientCACert(pair2.Certificate.Raw) },
			func() { _ = pool.AddPublicClientPair(cert1, key1) },
			func() { _ = pool.AddPublicClientPair(cert2, key2) },
			func() { _ = pool.AddPrivateRootCAPair(cert1, key1) },
			func() { _ = pool.AddPrivateRootCAPair(cert2, key2) },
			func() { _ = pool.AddPrivateClientCAPair(cert1, key1) },
			func() { _ = pool.AddPrivateClientCAPair(cert2, key2) },
			func() { _ = pool.AddPrivateClientPair(cert1, key1) },
			func() { _ = pool.AddPrivateClientPair(cert2, key2) },

			// delete
			func() { _ = pool.DeletePublicRootCACert(0) },
			func() { _ = pool.DeletePublicRootCACert(0) },
			func() { _ = pool.DeletePublicClientCACert(0) },
			func() { _ = pool.DeletePublicClientCACert(0) },
			func() { _ = pool.DeletePublicClientCert(0) },
			func() { _ = pool.DeletePublicClientCert(0) },
			func() { _ = pool.DeletePrivateRootCACert(0) },
			func() { _ = pool.DeletePrivateRootCACert(0) },
			func() { _ = pool.DeletePrivateClientCACert(0) },
			func() { _ = pool.DeletePrivateClientCACert(0) },
			func() { _ = pool.DeletePrivateClientCert(0) },
			func() { _ = pool.DeletePrivateClientCert(0) },

			// get
			func() { _ = pool.GetPublicRootCACerts() },
			func() { _ = pool.GetPublicClientCACerts() },
			func() { _ = pool.GetPublicClientPairs() },
			func() { _ = pool.GetPrivateRootCACerts() },
			func() { _ = pool.GetPrivateRootCAPairs() },
			func() { _ = pool.GetPrivateClientCACerts() },
			func() { _ = pool.GetPrivateClientCAPairs() },
			func() { _ = pool.GetPrivateClientPairs() },

			// export
			func() { _, _ = pool.ExportPublicRootCACert(0) },
			func() { _, _ = pool.ExportPublicClientCACert(0) },
			func() { _, _, _ = pool.ExportPublicClientPair(0) },
			func() { _, _, _ = pool.ExportPrivateRootCAPair(0) },
			func() { _, _, _ = pool.ExportPrivateClientCAPair(0) },
			func() { _, _, _ = pool.ExportPrivateClientPair(0) },
		}
		testsuite.RunParallel(100, init, nil, fns...)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("cert only (node and beacon)", func(t *testing.T) {
		var pool *Pool
		pair1 := testGeneratePair(t)
		pair2 := testGeneratePair(t)
		cert1, key1 := pair1.Encode()
		cert2, key2 := pair2.Encode()

		init := func() {
			pool = NewPool()
		}
		fns := []func(){
			// add
			func() { _ = pool.AddPublicRootCACert(pair1.Certificate.Raw) },
			func() { _ = pool.AddPublicRootCACert(pair2.Certificate.Raw) },
			func() { _ = pool.AddPublicClientCACert(pair1.Certificate.Raw) },
			func() { _ = pool.AddPublicClientCACert(pair2.Certificate.Raw) },
			func() { _ = pool.AddPublicClientPair(cert1, key1) },
			func() { _ = pool.AddPublicClientPair(cert2, key2) },
			func() { _ = pool.AddPrivateRootCACert(cert1) },
			func() { _ = pool.AddPrivateRootCACert(cert2) },
			func() { _ = pool.AddPrivateClientCACert(cert1) },
			func() { _ = pool.AddPrivateClientCACert(cert2) },
			func() { _ = pool.AddPrivateClientPair(cert1, key1) },
			func() { _ = pool.AddPrivateClientPair(cert2, key2) },

			// delete
			func() { _ = pool.DeletePublicRootCACert(0) },
			func() { _ = pool.DeletePublicRootCACert(0) },
			func() { _ = pool.DeletePublicClientCACert(0) },
			func() { _ = pool.DeletePublicClientCACert(0) },
			func() { _ = pool.DeletePublicClientCert(0) },
			func() { _ = pool.DeletePublicClientCert(0) },
			func() { _ = pool.DeletePrivateRootCACert(0) },
			func() { _ = pool.DeletePrivateRootCACert(0) },
			func() { _ = pool.DeletePrivateClientCACert(0) },
			func() { _ = pool.DeletePrivateClientCACert(0) },
			func() { _ = pool.DeletePrivateClientCert(0) },
			func() { _ = pool.DeletePrivateClientCert(0) },

			// get
			func() { _ = pool.GetPublicRootCACerts() },
			func() { _ = pool.GetPublicClientCACerts() },
			func() { _ = pool.GetPublicClientPairs() },
			func() { _ = pool.GetPrivateRootCACerts() },
			func() { _ = pool.GetPrivateClientCACerts() },
			func() { _ = pool.GetPrivateClientPairs() },
		}
		testsuite.RunParallel(100, init, nil, fns...)

		testsuite.IsDestroyed(t, pool)
	})
}

func TestNewPoolWithSystemCerts(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		_, err := NewPoolWithSystemCerts()
		require.NoError(t, err)
	})

	t.Run("failed to call SystemCertPool", func(t *testing.T) {
		patch := func() (*x509.CertPool, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(certpool.System, patch)
		defer pg.Unpatch()

		_, err := NewPoolWithSystemCerts()
		monkey.IsMonkeyError(t, err)
	})

	t.Run("failed to AddPublicRootCACert", func(t *testing.T) {
		pool := NewPool()

		patch := func(*Pool, []byte) error {
			return monkey.Error
		}
		pg := monkey.PatchInstanceMethod(pool, "AddPublicRootCACert", patch)
		defer pg.Unpatch()

		_, err := NewPoolWithSystemCerts()
		monkey.IsMonkeyError(t, err)

		testsuite.IsDestroyed(t, pool)
	})
}
