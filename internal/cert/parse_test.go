package cert

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCertificatePEM(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		pb, err := os.ReadFile("testdata/certs.pem")
		require.NoError(t, err)
		cert, err := ParseCertificatePEM(pb)
		require.NoError(t, err)
		t.Log(cert.Issuer)
	})

	t.Run("invalid PEM data", func(t *testing.T) {
		cert, err := ParseCertificatePEM([]byte{0, 1, 2, 3})
		require.Equal(t, ErrInvalidPEMBlock, err)
		require.Nil(t, cert)
	})

	t.Run("invalid type", func(t *testing.T) {
		pb := []byte(`
-----BEGIN INVALID TYPE-----
-----END INVALID TYPE-----
`)
		cert, err := ParseCertificatePEM(pb)
		require.EqualError(t, err, "invalid PEM block type: INVALID TYPE")
		require.Nil(t, cert)
	})

	t.Run("invalid certificate data", func(t *testing.T) {
		pb := []byte(`
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
`)
		cert, err := ParseCertificatePEM(pb)
		require.Error(t, err)
		require.Nil(t, cert)
	})
}

func TestParseCertificatesPEM(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		pb, err := os.ReadFile("testdata/certs.pem")
		require.NoError(t, err)
		certs, err := ParseCertificatesPEM(pb)
		require.NoError(t, err)
		t.Log(certs[0].Issuer)
		t.Log(certs[1].Issuer)
	})

	t.Run("invalid PEM data", func(t *testing.T) {
		certs, err := ParseCertificatesPEM([]byte{0, 1, 2, 3})
		require.Equal(t, ErrInvalidPEMBlock, err)
		require.Nil(t, certs)
	})

	t.Run("invalid type", func(t *testing.T) {
		pb := []byte(`
-----BEGIN INVALID TYPE-----
-----END INVALID TYPE-----
`)
		certs, err := ParseCertificatesPEM(pb)
		require.EqualError(t, err, "invalid PEM block type: INVALID TYPE")
		require.Nil(t, certs)
	})

	t.Run("invalid certificate data", func(t *testing.T) {
		pb := []byte(`
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
`)
		certs, err := ParseCertificatesPEM(pb)
		require.Error(t, err)
		require.Nil(t, certs)
	})
}

func TestParsePrivateKeyPEM(t *testing.T) {
	for _, file := range [...]string{
		"pkcs1", "pkcs8", "ecp",
	} {
		t.Run(file, func(t *testing.T) {
			der, err := os.ReadFile(fmt.Sprintf("testdata/%s.key", file))
			require.NoError(t, err)
			key, err := ParsePrivateKeyPEM(der)
			require.NoError(t, err)
			require.NotNil(t, key)
		})
	}

	t.Run("invalid PEM data", func(t *testing.T) {
		key, err := ParsePrivateKeyPEM([]byte{0, 1, 2, 3})
		require.Equal(t, ErrInvalidPEMBlock, err)
		require.Nil(t, key)
	})

	t.Run("invalid private key data", func(t *testing.T) {
		pb := []byte(`
-----BEGIN PRIVATE KEY-----
-----END PRIVATE KEY-----
`)
		key, err := ParsePrivateKeyPEM(pb)
		require.Error(t, err)
		require.Nil(t, key)
	})
}

func TestParsePrivateKeysPEM(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		pb, err := os.ReadFile("testdata/keys.pem")
		require.NoError(t, err)
		keys, err := ParsePrivateKeysPEM(pb)
		require.NoError(t, err)
		require.Len(t, keys, 2)
	})

	t.Run("invalid PEM data", func(t *testing.T) {
		keys, err := ParsePrivateKeysPEM([]byte{0, 1, 2, 3})
		require.Equal(t, ErrInvalidPEMBlock, err)
		require.Nil(t, keys)
	})

	t.Run("invalid private key data", func(t *testing.T) {
		pb := []byte(`
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
`)
		keys, err := ParsePrivateKeysPEM(pb)
		require.Error(t, err)
		require.Nil(t, keys)
	})
}
