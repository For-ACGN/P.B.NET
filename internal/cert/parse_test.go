package cert

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCertificate(t *testing.T) {
	certPEMBlock, err := os.ReadFile("testdata/certs.pem")
	require.NoError(t, err)
	cert, err := ParseCertificate(certPEMBlock)
	require.NoError(t, err)
	t.Log(cert.Issuer)

	// parse invalid PEM data
	_, err = ParseCertificate([]byte{0, 1, 2, 3})
	require.Equal(t, ErrInvalidPEMBlock, err)

	// invalid Type
	certPEMBlock = []byte(`
-----BEGIN INVALID TYPE-----
-----END INVALID TYPE-----
`)
	_, err = ParseCertificate(certPEMBlock)
	require.EqualError(t, err, "invalid PEM block type: INVALID TYPE")

	// invalid certificate data
	certPEMBlock = []byte(`
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
`)
	_, err = ParseCertificate(certPEMBlock)
	require.Error(t, err)
}

func TestParseCertificates(t *testing.T) {
	certPEMBlock, err := os.ReadFile("testdata/certs.pem")
	require.NoError(t, err)
	certs, err := ParseCertificates(certPEMBlock)
	require.NoError(t, err)
	t.Log(certs[0].Issuer)
	t.Log(certs[1].Issuer)

	// parse invalid PEM data
	_, err = ParseCertificates([]byte{0, 1, 2, 3})
	require.Equal(t, ErrInvalidPEMBlock, err)

	// invalid Type
	certPEMBlock = []byte(`
-----BEGIN INVALID TYPE-----
-----END INVALID TYPE-----
`)
	_, err = ParseCertificates(certPEMBlock)
	require.EqualError(t, err, "invalid PEM block type: INVALID TYPE")

	// invalid certificate data
	certPEMBlock = []byte(`
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
`)
	_, err = ParseCertificates(certPEMBlock)
	require.Error(t, err)
}

func TestParsePrivateKey(t *testing.T) {
	for _, file := range [...]string{
		"pkcs1.key", "pkcs8.key", "ecp.key",
	} {
		keyPEMBlock, err := os.ReadFile("testdata/" + file)
		require.NoError(t, err)
		_, err = ParsePrivateKey(keyPEMBlock)
		require.NoError(t, err)
	}

	// parse invalid PEM data
	_, err := ParsePrivateKey([]byte{0, 1, 2, 3})
	require.Equal(t, ErrInvalidPEMBlock, err)

	// invalid certificate data
	keyPEMBlock := []byte(`
-----BEGIN PRIVATE KEY-----
-----END PRIVATE KEY-----
`)
	_, err = ParsePrivateKey(keyPEMBlock)
	require.Error(t, err)
}

func TestParsePrivateKeys(t *testing.T) {
	keyPEMBlock, err := os.ReadFile("testdata/keys.pem")
	require.NoError(t, err)
	keys, err := ParsePrivateKeys(keyPEMBlock)
	require.NoError(t, err)
	require.Len(t, keys, 2)

	// parse invalid PEM data
	_, err = ParsePrivateKeys([]byte{0, 1, 2, 3})
	require.Equal(t, ErrInvalidPEMBlock, err)

	// invalid certificate data
	keyPEMBlock = []byte(`
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
`)
	_, err = ParsePrivateKeys(keyPEMBlock)
	require.Error(t, err)
}
