package cert

import (
	"crypto/x509"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/crypto/cert/certutil"
	"project/internal/patch/monkey"
	"project/internal/security"
)

func TestPair_ToPair(t *testing.T) {
	defer func() {
		require.NotNil(t, recover())
	}()
	pair := pair{
		PrivateKey: security.NewBytes(make([]byte, 1024)),
	}
	pair.ToPair()
}

func TestLoadCertWithPrivateKey(t *testing.T) {
	t.Run("invalid certificate", func(t *testing.T) {
		_, err := loadCertWithPrivateKey(make([]byte, 1024), nil)
		require.Error(t, err)
	})

	t.Run("invalid private key", func(t *testing.T) {
		pair := testGenerateCert(t)
		cert, _ := pair.Encode()
		_, err := loadCertWithPrivateKey(cert, make([]byte, 1024))
		require.Error(t, err)
	})

	t.Run("mismatched private key", func(t *testing.T) {
		pair1 := testGenerateCert(t)
		cert := pair1.ASN1()
		pair2 := testGenerateCert(t)
		_, key := pair2.Encode()
		_, err := loadCertWithPrivateKey(cert, key)
		require.Error(t, err)
	})

	t.Run("MarshalPKCS8PrivateKey", func(t *testing.T) {
		pair := testGenerateCert(t)
		cert, key := pair.Encode() // must before patch
		patchFunc := func(_ interface{}) ([]byte, error) {
			return nil, monkey.ErrMonkey
		}
		pg := monkey.Patch(x509.MarshalPKCS8PrivateKey, patchFunc)
		defer pg.Unpatch()
		_, err := loadCertWithPrivateKey(cert, key)
		monkey.IsMonkeyError(t, err)
	})
}

func testGenerateCert(t *testing.T) *Pair {
	pair, err := GenerateCA(nil)
	require.NoError(t, err)
	return pair
}

func testRunParallel(f ...func()) {
	l := len(f)
	if l == 0 {
		return
	}
	wg := sync.WaitGroup{}
	for i := 0; i < l; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			f[i]()
		}(i)
	}
	wg.Wait()
}

func TestPool(t *testing.T) {
	pool := NewPool()

	t.Run("PublicRootCACert", func(t *testing.T) {
		require.Equal(t, 0, len(pool.GetPublicRootCACerts()))

		pair := testGenerateCert(t)

		t.Run("add", func(t *testing.T) {
			err := pool.AddPublicRootCACert(pair.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPublicRootCACert(pair.Certificate.Raw)
			require.Error(t, err)
		})

		t.Run("get", func(t *testing.T) {
			certs := pool.GetPublicRootCACerts()
			require.True(t, certs[0].Equal(pair.Certificate))
		})

		t.Run("parallel", func(t *testing.T) {
			pair := testGenerateCert(t)
			add := func() {
				err := pool.AddPublicRootCACert(pair.Certificate.Raw)
				require.NoError(t, err)
			}
			get := func() {
				pool.GetPublicRootCACerts()
			}
			testRunParallel(add, get)
		})

		t.Run("invalid certificate", func(t *testing.T) {
			err := pool.AddPublicRootCACert(make([]byte, 1024))
			require.Error(t, err)
		})
	})

	t.Run("PublicClientCACert", func(t *testing.T) {
		require.Equal(t, 0, len(pool.GetPublicClientCACerts()))

		pair := testGenerateCert(t)

		t.Run("add", func(t *testing.T) {
			err := pool.AddPublicClientCACert(pair.Certificate.Raw)
			require.NoError(t, err)
			err = pool.AddPublicClientCACert(pair.Certificate.Raw)
			require.Error(t, err)
		})

		t.Run("get", func(t *testing.T) {
			certs := pool.GetPublicClientCACerts()
			require.True(t, certs[0].Equal(pair.Certificate))
		})

		t.Run("parallel", func(t *testing.T) {
			pair := testGenerateCert(t)
			add := func() {
				err := pool.AddPublicClientCACert(pair.Certificate.Raw)
				require.NoError(t, err)
			}
			get := func() {
				pool.GetPublicClientCACerts()
			}
			testRunParallel(add, get)
		})

		t.Run("invalid certificate", func(t *testing.T) {
			err := pool.AddPublicClientCACert(make([]byte, 1024))
			require.Error(t, err)
		})
	})

	t.Run("PublicClientCert", func(t *testing.T) {
		require.Equal(t, 0, len(pool.GetPublicClientPairs()))

		pair := testGenerateCert(t)

		t.Run("add", func(t *testing.T) {
			cert, key := pair.Encode()
			err := pool.AddPublicClientCert(cert, key)
			require.NoError(t, err)
			err = pool.AddPublicClientCert(cert, key)
			require.Error(t, err)
			err = pool.AddPublicClientCert(cert, nil)
			require.Error(t, err)
		})

		t.Run("get", func(t *testing.T) {
			pairs := pool.GetPublicClientPairs()
			require.Equal(t, pair, pairs[0])
		})

		t.Run("parallel", func(t *testing.T) {
			pair := testGenerateCert(t)
			add := func() {
				err := pool.AddPublicClientCert(pair.Encode())
				require.NoError(t, err)
			}
			get := func() {
				pool.GetPublicClientPairs()
			}
			testRunParallel(add, get)
		})

		t.Run("loadCertWithPrivateKey", func(t *testing.T) {
			err := pool.AddPublicClientCert(nil, nil)
			require.Error(t, err)
		})
	})

	t.Run("PrivateRootCACert", func(t *testing.T) {
		require.Equal(t, 0, len(pool.GetPrivateRootCACerts()))

		pair := testGenerateCert(t)

		t.Run("add", func(t *testing.T) {
			cert, key := pair.Encode()
			err := pool.AddPrivateRootCACert(cert, key)
			require.NoError(t, err)
			err = pool.AddPrivateRootCACert(cert, key)
			require.Error(t, err)
			err = pool.AddPrivateRootCACert(cert, []byte{})
			require.Error(t, err)
		})

		t.Run("get certs", func(t *testing.T) {
			certs := pool.GetPrivateRootCACerts()
			require.True(t, certs[0].Equal(pair.Certificate))
		})

		t.Run("get pairs", func(t *testing.T) {
			pairs := pool.GetPrivateRootCAPairs()
			require.Equal(t, pair, pairs[0])
		})

		t.Run("parallel", func(t *testing.T) {
			pair := testGenerateCert(t)
			add := func() {
				err := pool.AddPrivateRootCACert(pair.Encode())
				require.NoError(t, err)
			}
			getCerts := func() {
				pool.GetPrivateRootCACerts()
			}
			getPairs := func() {
				pool.GetPrivateRootCAPairs()
			}
			testRunParallel(add, getCerts, getPairs)
		})

		t.Run("loadCertWithPrivateKey", func(t *testing.T) {
			err := pool.AddPrivateRootCACert(nil, nil)
			require.Error(t, err)
		})
	})

	t.Run("PrivateClientCACert", func(t *testing.T) {
		require.Equal(t, 0, len(pool.GetPrivateClientCACerts()))

		pair := testGenerateCert(t)

		t.Run("add", func(t *testing.T) {
			cert, key := pair.Encode()
			err := pool.AddPrivateClientCACert(cert, key)
			require.NoError(t, err)
			err = pool.AddPrivateClientCACert(cert, key)
			require.Error(t, err)
			err = pool.AddPrivateClientCACert(cert, []byte{})
			require.Error(t, err)
		})

		t.Run("get certs", func(t *testing.T) {
			certs := pool.GetPrivateClientCACerts()
			require.True(t, certs[0].Equal(pair.Certificate))
		})

		t.Run("get pairs", func(t *testing.T) {
			pairs := pool.GetPrivateClientCAPairs()
			require.Equal(t, pair, pairs[0])
		})

		t.Run("parallel", func(t *testing.T) {
			pair := testGenerateCert(t)
			add := func() {
				err := pool.AddPrivateClientCACert(pair.Encode())
				require.NoError(t, err)
			}
			getCerts := func() {
				pool.GetPrivateClientCACerts()
			}
			getPairs := func() {
				pool.GetPrivateClientCAPairs()
			}
			testRunParallel(add, getCerts, getPairs)
		})

		t.Run("loadCertWithPrivateKey", func(t *testing.T) {
			err := pool.AddPrivateClientCACert(nil, nil)
			require.Error(t, err)
		})
	})

	t.Run("PrivateClientCert", func(t *testing.T) {
		require.Equal(t, 0, len(pool.GetPrivateClientPairs()))

		pair := testGenerateCert(t)

		t.Run("add", func(t *testing.T) {
			cert, key := pair.Encode()
			err := pool.AddPrivateClientCert(cert, key)
			require.NoError(t, err)
			err = pool.AddPrivateClientCert(cert, key)
			require.Error(t, err)
			err = pool.AddPrivateClientCert(cert, []byte{})
			require.Error(t, err)
		})

		t.Run("get", func(t *testing.T) {
			pairs := pool.GetPrivateClientPairs()
			require.Equal(t, pair, pairs[0])
		})

		t.Run("parallel", func(t *testing.T) {
			pair := testGenerateCert(t)
			add := func() {
				err := pool.AddPrivateClientCert(pair.Encode())
				require.NoError(t, err)
			}
			get := func() {
				pool.GetPrivateClientPairs()
			}
			testRunParallel(add, get)
		})

		t.Run("loadCertWithPrivateKey", func(t *testing.T) {
			err := pool.AddPrivateClientCert(nil, nil)
			require.Error(t, err)
		})
	})
}

func TestNewPoolWithSystemCerts(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		_, err := NewPoolWithSystemCerts()
		require.NoError(t, err)
	})

	t.Run("failed to call SystemCertPool", func(t *testing.T) {
		patchFunc := func() (*x509.CertPool, error) {
			return nil, monkey.ErrMonkey
		}
		pg := monkey.Patch(certutil.SystemCertPool, patchFunc)
		defer pg.Unpatch()
		_, err := NewPoolWithSystemCerts()
		monkey.IsMonkeyError(t, err)
	})

	t.Run("failed to AddPublicRootCACert", func(t *testing.T) {
		pool := NewPool()
		patchFunc := func(_ *Pool, _ []byte) error {
			return monkey.ErrMonkey
		}
		pg := monkey.PatchInstanceMethod(pool, "AddPublicRootCACert", patchFunc)
		defer pg.Unpatch()
		_, err := NewPoolWithSystemCerts()
		monkey.IsMonkeyError(t, err)
	})
}
