package cert

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestPair_Encode(t *testing.T) {
	ca, err := GenerateCA(nil)
	require.NoError(t, err)

	p := &Pair{Certificate: ca.Certificate}

	defer testsuite.DeferForPanic(t)
	p.Encode()
}

func TestPair_EncodeToPEM(t *testing.T) {
	ca, err := GenerateCA(nil)
	require.NoError(t, err)

	_, err = tls.X509KeyPair(ca.EncodeToPEM())
	require.NoError(t, err)
}

func TestPair_TLSCertificate(t *testing.T) {
	ca, err := GenerateCA(nil)
	require.NoError(t, err)

	cert := ca.TLSCertificate()
	require.NotNil(t, cert)
}
