package cert

import (
	"fmt"
	"strings"
)

func generateTestPoolAddCertParallel(f1 string) {
	const template = `
func TestPool_Add<f1>Cert_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		add1 := func() {
			err := pool.Add<f1>Cert(pair1.Certificate.Raw)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.Add<f1>Cert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.Add<f1>Cert(nil)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.Get<f1>Certs()
			require.Len(t, certs, 2)

			err := pool.Delete<f1>Cert(0)
			require.NoError(t, err)
			err = pool.Delete<f1>Cert(0)
			require.NoError(t, err)

			certs = pool.Get<f1>Certs()
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
			err := pool.Add<f1>Cert(pair1.Certificate.Raw)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.Add<f1>Cert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.Add<f1>Cert(nil)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.Get<f1>Certs()
			require.Len(t, certs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, add1, add2, add3)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}
`
	code := strings.ReplaceAll(template, "<f1>", f1)
	fmt.Println(code)
}

func generateTestPoolAddPairParallel(f1 string) {
	const template = `
func TestPool_Add<f1>Pair_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		add1 := func() {
			err := pool.Add<f1>Pair(cert1, key1)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.Add<f1>Pair(cert2, key2)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.Add<f1>Pair(nil, nil)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.Get<f1>Pairs()
			require.Len(t, pairs, 2)

			err := pool.Delete<f1>Cert(0)
			require.NoError(t, err)
			err = pool.Delete<f1>Cert(0)
			require.NoError(t, err)

			pairs = pool.Get<f1>Pairs()
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
			err := pool.Add<f1>Pair(cert1, key1)
			require.NoError(t, err)
		}
		add2 := func() {
			err := pool.Add<f1>Pair(cert2, key2)
			require.NoError(t, err)
		}
		add3 := func() {
			err := pool.Add<f1>Pair(nil, nil)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.Get<f1>Pairs()
			require.Len(t, pairs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, add1, add2, add3)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}
`
	code := strings.ReplaceAll(template, "<f1>", f1)
	fmt.Println(code)
}

func generateTestPoolDeleteCertParallel(f1 string) {
	const template = `
func TestPool_Delete<f1>Cert_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
			err := pool.Add<f1>Cert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.Add<f1>Cert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		delete1 := func() {
			err := pool.Delete<f1>Cert(0)
			require.NoError(t, err)
		}
		delete2 := func() {
			err := pool.Delete<f1>Cert(0)
			require.NoError(t, err)
		}
		delete3 := func() {
			err := pool.Delete<f1>Cert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.Get<f1>Certs()
			require.Len(t, certs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()

			err := pool.Add<f1>Cert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.Add<f1>Cert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		delete1 := func() {
			err := pool.Delete<f1>Cert(0)
			require.NoError(t, err)
		}
		delete2 := func() {
			err := pool.Delete<f1>Cert(0)
			require.NoError(t, err)
		}
		delete3 := func() {
			err := pool.Delete<f1>Cert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			certs := pool.Get<f1>Certs()
			require.Len(t, certs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}
`
	code := strings.ReplaceAll(template, "<f1>", f1)
	fmt.Println(code)
}

func generateTestPoolDeletePairParallel(f1 string) {
	const template = `
func TestPool_Delete<f1>Cert_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
			err := pool.Add<f1>Pair(cert1, key1)
			require.NoError(t, err)
			err = pool.Add<f1>Pair(cert2, key2)
			require.NoError(t, err)
		}
		delete1 := func() {
			err := pool.Delete<f1>Cert(0)
			require.NoError(t, err)
		}
		delete2 := func() {
			err := pool.Delete<f1>Cert(0)
			require.NoError(t, err)
		}
		delete3 := func() {
			err := pool.Delete<f1>Cert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.Get<f1>Pairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()

			err := pool.Add<f1>Pair(cert1, key1)
			require.NoError(t, err)
			err = pool.Add<f1>Pair(cert2, key2)
			require.NoError(t, err)
		}
		delete1 := func() {
			err := pool.Delete<f1>Cert(0)
			require.NoError(t, err)
		}
		delete2 := func() {
			err := pool.Delete<f1>Cert(0)
			require.NoError(t, err)
		}
		delete3 := func() {
			err := pool.Delete<f1>Cert(2)
			require.Error(t, err)
		}
		cleanup := func() {
			pairs := pool.Get<f1>Pairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, delete1, delete2, delete3)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}
`
	code := strings.ReplaceAll(template, "<f1>", f1)
	fmt.Println(code)
}

func generateTestPoolGetCertsParallel(f1 string) {
	const template = `
func TestPool_Get<f1>Certs_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
			err := pool.Add<f1>Cert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.Add<f1>Cert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		get := func() {
			certs := pool.Get<f1>Certs()
			expected := []*x509.Certificate{pair1.Certificate, pair2.Certificate}
			require.Equal(t, expected, certs)
		}
		cleanup := func() {
			err := pool.Delete<f1>Cert(0)
			require.NoError(t, err)
			err = pool.Delete<f1>Cert(0)
			require.NoError(t, err)
		}
		testsuite.RunParallel(100, init, cleanup, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()

			err := pool.Add<f1>Cert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.Add<f1>Cert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		get := func() {
			certs := pool.Get<f1>Certs()
			expected := []*x509.Certificate{pair1.Certificate, pair2.Certificate}
			require.Equal(t, expected, certs)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}
`
	code := strings.ReplaceAll(template, "<f1>", f1)
	fmt.Println(code)
}

func generateTestPoolGetPairsParallel(f1 string) {
	const template = `
func TestPool_Get<f1>Pairs_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, key1 := pair1.Encode()
	cert2, key2 := pair2.Encode()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
			err := pool.Add<f1>Pair(cert1, key1)
			require.NoError(t, err)
			err = pool.Add<f1>Pair(cert2, key2)
			require.NoError(t, err)
		}
		get := func() {
			pairs := pool.Get<f1>Pairs()
			expected := []*Pair{pair1, pair2}
			require.Equal(t, expected, pairs)
		}
		cleanup := func() {
			err := pool.Delete<f1>Cert(0)
			require.NoError(t, err)
			err = pool.Delete<f1>Cert(0)
			require.NoError(t, err)
		}
		testsuite.RunParallel(100, init, cleanup, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()

			err := pool.Add<f1>Pair(cert1, key1)
			require.NoError(t, err)
			err = pool.Add<f1>Pair(cert2, key2)
			require.NoError(t, err)
		}
		get := func() {
			pairs := pool.Get<f1>Pairs()
			expected := []*Pair{pair1, pair2}
			require.Equal(t, expected, pairs)
		}
		testsuite.RunParallel(100, init, nil, get, get)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}
`
	code := strings.ReplaceAll(template, "<f1>", f1)
	fmt.Println(code)
}

func generateTestPoolExportCertParallel(f1 string) {
	const template = `
func TestPool_Export<f1>Cert_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pair1 := testGeneratePair(t)
	pair2 := testGeneratePair(t)
	cert1, _ := pair1.EncodeToPEM()
	cert2, _ := pair2.EncodeToPEM()

	t.Run("part", func(t *testing.T) {
		pool := NewPool()

		init := func() {
			err := pool.Add<f1>Cert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.Add<f1>Cert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		export1 := func() {
			crt, err := pool.Export<f1>Cert(0)
			require.NoError(t, err)
			require.Equal(t, cert1, crt)
		}
		export2 := func() {
			crt, err := pool.Export<f1>Cert(1)
			require.NoError(t, err)
			require.Equal(t, cert2, crt)
		}
		export3 := func() {
			crt, err := pool.Export<f1>Cert(2)
			require.Error(t, err)
			require.Nil(t, crt)
		}
		cleanup := func() {
			certs := pool.Get<f1>Certs()
			require.Len(t, certs, 2)

			err := pool.Delete<f1>Cert(0)
			require.NoError(t, err)
			err = pool.Delete<f1>Cert(0)
			require.NoError(t, err)

			certs = pool.Get<f1>Certs()
			require.Len(t, certs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()

			err := pool.Add<f1>Cert(pair1.Certificate.Raw)
			require.NoError(t, err)
			err = pool.Add<f1>Cert(pair2.Certificate.Raw)
			require.NoError(t, err)
		}
		export1 := func() {
			crt, err := pool.Export<f1>Cert(0)
			require.NoError(t, err)
			require.Equal(t, cert1, crt)
		}
		export2 := func() {
			crt, err := pool.Export<f1>Cert(1)
			require.NoError(t, err)
			require.Equal(t, cert2, crt)
		}
		export3 := func() {
			crt, err := pool.Export<f1>Cert(2)
			require.Error(t, err)
			require.Nil(t, crt)
		}
		cleanup := func() {
			certs := pool.Get<f1>Certs()
			require.Len(t, certs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}
`
	code := strings.ReplaceAll(template, "<f1>", f1)
	fmt.Println(code)
}

func generateTestPoolExportPairParallel(f1 string) {
	const template = `
func TestPool_Export<f1>Pair_Parallel(t *testing.T) {
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
			err := pool.Add<f1>Pair(cert1, key1)
			require.NoError(t, err)
			err = pool.Add<f1>Pair(cert2, key2)
			require.NoError(t, err)
		}
		export1 := func() {
			crt, key, err := pool.Export<f1>Pair(0)
			require.NoError(t, err)
			require.Equal(t, cert1PEM, crt)
			require.Equal(t, key1PEM, key)
		}
		export2 := func() {
			crt, key, err := pool.Export<f1>Pair(1)
			require.NoError(t, err)
			require.Equal(t, cert2PEM, crt)
			require.Equal(t, key2PEM, key)
		}
		export3 := func() {
			crt, key, err := pool.Export<f1>Pair(2)
			require.Error(t, err)
			require.Nil(t, crt)
			require.Nil(t, key)
		}
		cleanup := func() {
			pairs := pool.Get<f1>Pairs()
			require.Len(t, pairs, 2)

			err := pool.Delete<f1>Cert(0)
			require.NoError(t, err)
			err = pool.Delete<f1>Cert(0)
			require.NoError(t, err)

			pairs = pool.Get<f1>Pairs()
			require.Len(t, pairs, 0)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})

	t.Run("whole", func(t *testing.T) {
		var pool *Pool

		init := func() {
			pool = NewPool()

			err := pool.Add<f1>Pair(cert1, key1)
			require.NoError(t, err)
			err = pool.Add<f1>Pair(cert2, key2)
			require.NoError(t, err)
		}
		export1 := func() {
			crt, key, err := pool.Export<f1>Pair(0)
			require.NoError(t, err)
			require.Equal(t, cert1PEM, crt)
			require.Equal(t, key1PEM, key)
		}
		export2 := func() {
			crt, key, err := pool.Export<f1>Pair(1)
			require.NoError(t, err)
			require.Equal(t, cert2PEM, crt)
			require.Equal(t, key2PEM, key)
		}
		export3 := func() {
			crt, key, err := pool.Export<f1>Pair(2)
			require.Error(t, err)
			require.Nil(t, crt)
			require.Nil(t, key)
		}
		cleanup := func() {
			pairs := pool.Get<f1>Pairs()
			require.Len(t, pairs, 2)
		}
		testsuite.RunParallel(100, init, cleanup, export1, export2, export3)

		testsuite.IsDestroyed(t, pool)
	})

	testsuite.IsDestroyed(t, pair1)
	testsuite.IsDestroyed(t, pair2)
}
`
	code := strings.ReplaceAll(template, "<f1>", f1)
	fmt.Println(code)
}
