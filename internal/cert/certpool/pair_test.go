package certpool

import (
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
	"project/internal/security"
	"project/internal/testsuite"
)

func TestPair_ToCertPair(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		pair := testGeneratePair(t)

		p, err := loadPair(pair.Encode())
		require.NoError(t, err)

		cp := p.ToCertPair()
		require.NotNil(t, cp)
	})

	t.Run("panic", func(t *testing.T) {
		sb := security.NewBytes(make([]byte, 1024))
		pair := pair{PrivateKey: sb}

		defer testsuite.DeferForPanic(t)
		pair.ToCertPair()
	})
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
		crt, _ := pair.Encode()

		_, err := loadPair(crt, make([]byte, 1024))
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

	t.Run("failed to marshal private key", func(t *testing.T) {
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
