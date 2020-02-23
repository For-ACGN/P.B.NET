package testcert

import (
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/crypto/cert/certutil"
)

func TestCertPool(t *testing.T) {
	pool := CertPool(t)

	systemCertPool, err := certutil.SystemCertPool()
	require.NoError(t, err)
	certs := systemCertPool.Certs()
	require.Equal(t, len(certs), len(pool.GetPublicRootCACerts()))

	require.Equal(t, PublicClientCANum, len(pool.GetPublicClientCACerts()))
	require.Equal(t, PublicClientCertNum, len(pool.GetPublicClientPairs()))

	require.Equal(t, PrivateRootCANum, len(pool.GetPrivateRootCAPairs()))
	require.Equal(t, PrivateClientCANum, len(pool.GetPrivateClientCAPairs()))
	require.Equal(t, PrivateClientCertNum, len(pool.GetPrivateClientPairs()))
}
