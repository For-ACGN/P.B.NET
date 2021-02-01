package certpool

import (
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/cert"
	"project/internal/patch/monkey"
	"project/internal/testsuite"
)

func testGeneratePair(t *testing.T) *cert.Pair {
	pair, err := cert.GenerateCA(nil)
	require.NoError(t, err)
	return pair
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

	t.Run("is not exist", func(t *testing.T) {
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

		for _, id := range [...]int{
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

	t.Run("is not exist", func(t *testing.T) {
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

		for _, id := range [...]int{
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

	t.Run("is not exist", func(t *testing.T) {
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

		for _, id := range [...]int{
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

	t.Run("is not exist", func(t *testing.T) {
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

		for _, id := range [...]int{
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

	t.Run("is not exist", func(t *testing.T) {
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

		for _, id := range [...]int{
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

	t.Run("is not exist", func(t *testing.T) {
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

		for _, id := range [...]int{
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

		for _, id := range [...]int{
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

		for _, id := range [...]int{
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

		for _, id := range [...]int{
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

		for _, id := range [...]int{
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

		for _, id := range [...]int{
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

		for _, id := range [...]int{
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

func TestPool_AddPublicRootCACert_Parallel(t *testing.T) {
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

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_AddPublicClientCACert_Parallel(t *testing.T) {
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

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_AddPublicClientPair_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

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

			err := pool.DeletePublicClientCert(0)
			require.NoError(t, err)
			err = pool.DeletePublicClientCert(0)
			require.NoError(t, err)

			pairs = pool.GetPublicClientPairs()
			require.Len(t, pairs, 0)
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

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_AddPrivateRootCAPair_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

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

			err := pool.DeletePrivateRootCACert(0)
			require.NoError(t, err)
			err = pool.DeletePrivateRootCACert(0)
			require.NoError(t, err)

			pairs = pool.GetPrivateRootCAPairs()
			require.Len(t, pairs, 0)
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

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_AddPrivateRootCACert_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		add1 := func() {
			err := pool.AddPrivateRootCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.AddPrivateRootCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.AddPrivateRootCACert(nil)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.GetPrivateRootCACerts()
			require.Len(t, certs, 2)

			err := pool.DeletePrivateRootCACert(0)
			require.NoError(t, err)
			err = pool.DeletePrivateRootCACert(0)
			require.NoError(t, err)

			certs = pool.GetPrivateRootCACerts()
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
			err := pool.AddPrivateRootCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.AddPrivateRootCACert(pair2.Certificate.Raw)
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

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_AddPrivateClientCAPair_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

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

			err := pool.DeletePrivateClientCACert(0)
			require.NoError(t, err)
			err = pool.DeletePrivateClientCACert(0)
			require.NoError(t, err)

			pairs = pool.GetPrivateClientCAPairs()
			require.Len(t, pairs, 0)
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

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_AddPrivateClientCACert_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		add1 := func() {
			err := pool.AddPrivateClientCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.AddPrivateClientCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.AddPrivateClientCACert(nil)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.GetPrivateClientCACerts()
			require.Len(t, certs, 2)

			err := pool.DeletePrivateClientCACert(0)
			require.NoError(t, err)
			err = pool.DeletePrivateClientCACert(0)
			require.NoError(t, err)

			certs = pool.GetPrivateClientCACerts()
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
			err := pool.AddPrivateClientCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.AddPrivateClientCACert(pair2.Certificate.Raw)
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

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_AddPrivateClientPair_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

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

			err := pool.DeletePrivateClientCert(0)
			require.NoError(t, err)
			err = pool.DeletePrivateClientCert(0)
			require.NoError(t, err)

			pairs = pool.GetPrivateClientPairs()
			require.Len(t, pairs, 0)
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

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_DeletePublicRootCACert_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
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
			err := pool.DeletePublicRootCACert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.GetPublicRootCACerts()
			require.Len(t, certs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

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
			err := pool.DeletePublicRootCACert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.GetPublicRootCACerts()
			require.Len(t, certs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_DeletePublicClientCACert_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
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
			err := pool.DeletePublicClientCACert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.GetPublicClientCACerts()
			require.Len(t, certs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

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
			err := pool.DeletePublicClientCACert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.GetPublicClientCACerts()
			require.Len(t, certs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_DeletePublicClientCert_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
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
			err := pool.DeletePublicClientCert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPublicClientPairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

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
			err := pool.DeletePublicClientCert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPublicClientPairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_DeletePrivateRootCACert_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
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
			err := pool.DeletePrivateRootCACert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPrivateRootCAPairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

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
			err := pool.DeletePrivateRootCACert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPrivateRootCAPairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_DeletePrivateClientCACert_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
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
			err := pool.DeletePrivateClientCACert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPrivateClientCAPairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

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
			err := pool.DeletePrivateClientCACert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPrivateClientCAPairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_DeletePrivateClientCert_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
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
			err := pool.DeletePrivateClientCert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPrivateClientPairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

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
			err := pool.DeletePrivateClientCert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.GetPrivateClientPairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_GetPublicRootCACerts_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
			err := pool.AddPublicRootCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPublicRootCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		get := func() {
			certs := pool.GetPublicRootCACerts()
			expected := []*x509.Certificate{pair1.Certificate, pair2.Certificate}
			require.Equal(t, expected, certs)
		}
		cleanup := func() {
			err := pool.DeletePublicRootCACert(0)
			require.NoError(t, err)
			err = pool.DeletePublicRootCACert(0)
			require.NoError(t, err)
		}
		testsuite.RunParallel(100, init, cleanup, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()

			err := pool.AddPublicRootCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPublicRootCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		get := func() {
			certs := pool.GetPublicRootCACerts()
			expected := []*x509.Certificate{pair1.Certificate, pair2.Certificate}
			require.Equal(t, expected, certs)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_GetPublicClientCACerts_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
			err := pool.AddPublicClientCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPublicClientCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		get := func() {
			certs := pool.GetPublicClientCACerts()
			expected := []*x509.Certificate{pair1.Certificate, pair2.Certificate}
			require.Equal(t, expected, certs)
		}
		cleanup := func() {
			err := pool.DeletePublicClientCACert(0)
			require.NoError(t, err)
			err = pool.DeletePublicClientCACert(0)
			require.NoError(t, err)
		}
		testsuite.RunParallel(100, init, cleanup, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()

			err := pool.AddPublicClientCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPublicClientCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		get := func() {
			certs := pool.GetPublicClientCACerts()
			expected := []*x509.Certificate{pair1.Certificate, pair2.Certificate}
			require.Equal(t, expected, certs)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_GetPublicClientPairs_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
			err := pool.AddPublicClientPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPublicClientPair(cert2, key2)
			require.NoError(t, err)
		}
		get := func() {
			pairs := pool.GetPublicClientPairs()
			expected := []*cert.Pair{pair1, pair2}
			require.Equal(t, expected, pairs)
		}
		cleanup := func() {
			err := pool.DeletePublicClientCert(0)
			require.NoError(t, err)
			err = pool.DeletePublicClientCert(0)
			require.NoError(t, err)
		}
		testsuite.RunParallel(100, init, cleanup, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()

			err := pool.AddPublicClientPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPublicClientPair(cert2, key2)
			require.NoError(t, err)
		}
		get := func() {
			pairs := pool.GetPublicClientPairs()
			expected := []*cert.Pair{pair1, pair2}
			require.Equal(t, expected, pairs)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_GetPrivateRootCAPairs_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
			err := pool.AddPrivateRootCAPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPrivateRootCAPair(cert2, key2)
			require.NoError(t, err)
		}
		get := func() {
			pairs := pool.GetPrivateRootCAPairs()
			expected := []*cert.Pair{pair1, pair2}
			require.Equal(t, expected, pairs)
		}
		cleanup := func() {
			err := pool.DeletePrivateRootCACert(0)
			require.NoError(t, err)
			err = pool.DeletePrivateRootCACert(0)
			require.NoError(t, err)
		}
		testsuite.RunParallel(100, init, cleanup, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateRootCAPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPrivateRootCAPair(cert2, key2)
			require.NoError(t, err)
		}
		get := func() {
			pairs := pool.GetPrivateRootCAPairs()
			expected := []*cert.Pair{pair1, pair2}
			require.Equal(t, expected, pairs)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_GetPrivateRootCACerts_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
			err := pool.AddPrivateRootCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPrivateRootCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		get := func() {
			certs := pool.GetPrivateRootCACerts()
			expected := []*x509.Certificate{pair1.Certificate, pair2.Certificate}
			require.Equal(t, expected, certs)
		}
		cleanup := func() {
			err := pool.DeletePrivateRootCACert(0)
			require.NoError(t, err)
			err = pool.DeletePrivateRootCACert(0)
			require.NoError(t, err)
		}
		testsuite.RunParallel(100, init, cleanup, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateRootCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPrivateRootCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		get := func() {
			certs := pool.GetPrivateRootCACerts()
			expected := []*x509.Certificate{pair1.Certificate, pair2.Certificate}
			require.Equal(t, expected, certs)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_GetPrivateClientCAPairs_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
			err := pool.AddPrivateClientCAPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPrivateClientCAPair(cert2, key2)
			require.NoError(t, err)
		}
		get := func() {
			pairs := pool.GetPrivateClientCAPairs()
			expected := []*cert.Pair{pair1, pair2}
			require.Equal(t, expected, pairs)
		}
		cleanup := func() {
			err := pool.DeletePrivateClientCACert(0)
			require.NoError(t, err)
			err = pool.DeletePrivateClientCACert(0)
			require.NoError(t, err)
		}
		testsuite.RunParallel(100, init, cleanup, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateClientCAPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPrivateClientCAPair(cert2, key2)
			require.NoError(t, err)
		}
		get := func() {
			pairs := pool.GetPrivateClientCAPairs()
			expected := []*cert.Pair{pair1, pair2}
			require.Equal(t, expected, pairs)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_GetPrivateClientCACerts_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
			err := pool.AddPrivateClientCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPrivateClientCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		get := func() {
			certs := pool.GetPrivateClientCACerts()
			expected := []*x509.Certificate{pair1.Certificate, pair2.Certificate}
			require.Equal(t, expected, certs)
		}
		cleanup := func() {
			err := pool.DeletePrivateClientCACert(0)
			require.NoError(t, err)
			err = pool.DeletePrivateClientCACert(0)
			require.NoError(t, err)
		}
		testsuite.RunParallel(100, init, cleanup, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateClientCACert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPrivateClientCACert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		get := func() {
			certs := pool.GetPrivateClientCACerts()
			expected := []*x509.Certificate{pair1.Certificate, pair2.Certificate}
			require.Equal(t, expected, certs)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_GetPrivateClientPairs_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
			err := pool.AddPrivateClientPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPrivateClientPair(cert2, key2)
			require.NoError(t, err)
		}
		get := func() {
			pairs := pool.GetPrivateClientPairs()
			expected := []*cert.Pair{pair1, pair2}
			require.Equal(t, expected, pairs)
		}
		cleanup := func() {
			err := pool.DeletePrivateClientCert(0)
			require.NoError(t, err)
			err = pool.DeletePrivateClientCert(0)
			require.NoError(t, err)
		}
		testsuite.RunParallel(100, init, cleanup, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()

			err := pool.AddPrivateClientPair(cert1, key1)
			require.NoError(t, err)
			err = pool.AddPrivateClientPair(cert2, key2)
			require.NoError(t, err)
		}
		get := func() {
			pairs := pool.GetPrivateClientPairs()
			expected := []*cert.Pair{pair1, pair2}
			require.Equal(t, expected, pairs)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_ExportPublicRootCACert_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, _ := pair1.EncodeToPEM()
	cert2, _ := pair2.EncodeToPEM()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
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

			err := pool.DeletePublicRootCACert(0)
			require.NoError(t, err)
			err = pool.DeletePublicRootCACert(0)
			require.NoError(t, err)

			certs = pool.GetPublicRootCACerts()
			require.Len(t, certs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

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

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_ExportPublicClientCACert_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, _ := pair1.EncodeToPEM()
	cert2, _ := pair2.EncodeToPEM()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
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

			err := pool.DeletePublicClientCACert(0)
			require.NoError(t, err)
			err = pool.DeletePublicClientCACert(0)
			require.NoError(t, err)

			certs = pool.GetPublicClientCACerts()
			require.Len(t, certs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

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

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_ExportPublicClientPair_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()
	cert1PEM, key1PEM := pair1.EncodeToPEM()
	cert2PEM, key2PEM := pair2.EncodeToPEM()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
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

			err := pool.DeletePublicClientCert(0)
			require.NoError(t, err)
			err = pool.DeletePublicClientCert(0)
			require.NoError(t, err)

			pairs = pool.GetPublicClientPairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

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

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_ExportPrivateRootCAPair_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()
	cert1PEM, key1PEM := pair1.EncodeToPEM()
	cert2PEM, key2PEM := pair2.EncodeToPEM()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
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

			err := pool.DeletePrivateRootCACert(0)
			require.NoError(t, err)
			err = pool.DeletePrivateRootCACert(0)
			require.NoError(t, err)

			pairs = pool.GetPrivateRootCAPairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

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

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_ExportPrivateClientCAPair_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()
	cert1PEM, key1PEM := pair1.EncodeToPEM()
	cert2PEM, key2PEM := pair2.EncodeToPEM()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
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

			err := pool.DeletePrivateClientCACert(0)
			require.NoError(t, err)
			err = pool.DeletePrivateClientCACert(0)
			require.NoError(t, err)

			pairs = pool.GetPrivateClientCAPairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

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

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_ExportPrivateClientPair_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()
	cert1PEM, key1PEM := pair1.EncodeToPEM()
	cert2PEM, key2PEM := pair2.EncodeToPEM()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
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

			err := pool.DeletePrivateClientCert(0)
			require.NoError(t, err)
			err = pool.DeletePrivateClientCert(0)
			require.NoError(t, err)

			pairs = pool.GetPrivateClientPairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

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

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestPool_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("pair only (controller)", func(t *testing.T) {
		t.Run("part", func(t *testing.T) {
			pool := NewPool()

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
			cleanup := func() {
				_ = pool.DeletePublicRootCACert(0)
				_ = pool.DeletePublicRootCACert(0)
				_ = pool.DeletePublicClientCACert(0)
				_ = pool.DeletePublicClientCACert(0)
				_ = pool.DeletePublicClientCert(0)
				_ = pool.DeletePublicClientCert(0)
				_ = pool.DeletePrivateRootCACert(0)
				_ = pool.DeletePrivateRootCACert(0)
				_ = pool.DeletePrivateClientCACert(0)
				_ = pool.DeletePrivateClientCACert(0)
				_ = pool.DeletePrivateClientCert(0)
				_ = pool.DeletePrivateClientCert(0)
			}
			testsuite.RunParallel(100, nil, cleanup, fns...)

			testsuite.IsDestroyed(t, pool)
		})

		t.Run("whole", func(t *testing.T) {
			var pool *Pool

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
	})

	t.Run("cert only (node and beacon)", func(t *testing.T) {
		t.Run("part", func(t *testing.T) {
			pool := NewPool()

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
			cleanup := func() {
				_ = pool.DeletePublicRootCACert(0)
				_ = pool.DeletePublicRootCACert(0)
				_ = pool.DeletePublicClientCACert(0)
				_ = pool.DeletePublicClientCACert(0)
				_ = pool.DeletePublicClientCert(0)
				_ = pool.DeletePublicClientCert(0)
				_ = pool.DeletePrivateRootCACert(0)
				_ = pool.DeletePrivateRootCACert(0)
				_ = pool.DeletePrivateClientCACert(0)
				_ = pool.DeletePrivateClientCACert(0)
				_ = pool.DeletePrivateClientCert(0)
				_ = pool.DeletePrivateClientCert(0)
			}
			testsuite.RunParallel(100, nil, cleanup, fns...)

			testsuite.IsDestroyed(t, pool)
		})

		t.Run("whole", func(t *testing.T) {
			var pool *Pool

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
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}

func TestNewPoolWithSystem(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		_, err := NewPoolWithSystem()
		require.NoError(t, err)
	})

	t.Run("failed to call SystemCertPool", func(t *testing.T) {
		patch := func() (*x509.CertPool, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(System, patch)
		defer pg.Unpatch()

		_, err := NewPoolWithSystem()
		monkey.IsMonkeyError(t, err)
	})

	t.Run("failed to AddPublicRootCACert", func(t *testing.T) {
		pool := NewPool()

		patch := func(*Pool, []byte) error {
			return monkey.Error
		}
		pg := monkey.PatchInstanceMethod(pool, "AddPublicRootCACert", patch)
		defer pg.Unpatch()

		_, err := NewPoolWithSystem()
		monkey.IsMonkeyError(t, err)

		testsuite.IsDestroyed(t, pool)
	})
}
