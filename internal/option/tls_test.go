package option

import (
	"crypto/tls"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/toml"
	"project/internal/testsuite"
	"project/internal/testsuite/testcert"
)

func TestTLSConfigDefault(t *testing.T) {
	t.Run("client side", func(t *testing.T) {
		tlsConfig := TLSConfig{}

		t.Run("without cert pool", func(t *testing.T) {
			config, err := tlsConfig.Apply()
			require.NoError(t, err)

			require.Len(t, config.Certificates, 0)
			require.Len(t, config.RootCAs.Certs(), 0)
			require.Len(t, config.ClientCAs.Certs(), 0)
		})

		t.Run("with cert pool", func(t *testing.T) {
			tlsConfig.CertPool = testcert.CertPool(t)
			config, err := tlsConfig.Apply()
			require.NoError(t, err)

			require.Len(t, config.Certificates, testcert.PublicClientCertNum)
			require.Len(t, config.RootCAs.Certs(), testcert.PublicRootCANum)
			require.Len(t, config.ClientCAs.Certs(), 0)
		})
	})

	t.Run("server side", func(t *testing.T) {
		tlsConfig := TLSConfig{ServerSide: true}

		t.Run("without cert pool", func(t *testing.T) {
			config, err := tlsConfig.Apply()
			require.NoError(t, err)

			require.Len(t, config.Certificates, 0)
			require.Len(t, config.RootCAs.Certs(), 0)
			require.Len(t, config.ClientCAs.Certs(), 0)
		})

		t.Run("with cert pool", func(t *testing.T) {
			tlsConfig.CertPool = testcert.CertPool(t)
			config, err := tlsConfig.Apply()
			require.NoError(t, err)

			require.Len(t, config.Certificates, 0)
			require.Len(t, config.RootCAs.Certs(), 0)
			require.Len(t, config.ClientCAs.Certs(), testcert.PublicClientCANum)
		})
	})

	t.Run("common", func(t *testing.T) {
		config, err := new(TLSConfig).Apply()
		require.NoError(t, err)

		require.Equal(t, tls.ClientAuthType(0), config.ClientAuth)
		require.Zero(t, config.ServerName)
		require.Nil(t, config.NextProtos)
		require.Equal(t, uint16(tls.VersionTLS12), config.MinVersion)
		require.Equal(t, uint16(0), config.MaxVersion)
		require.Nil(t, config.CipherSuites)
		require.Equal(t, false, config.InsecureSkipVerify)
	})
}

// the number of the certificate in testdata/tls.toml
const (
	testRootCANum      = 3
	testClientCANum    = 2
	testCertificateNum = 1
)

func TestTLSConfig(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/tls.toml")
	require.NoError(t, err)

	// check unnecessary field
	tlsConfig := TLSConfig{}
	err = toml.Unmarshal(data, &tlsConfig)
	require.NoError(t, err)

	// check zero value
	testsuite.CheckOptions(t, tlsConfig)

	t.Run("client side", func(t *testing.T) {
		t.Run("without cert pool", func(t *testing.T) {
			config, err := tlsConfig.Apply()
			require.NoError(t, err)

			require.Len(t, config.Certificates, testCertificateNum)
			require.Len(t, config.RootCAs.Certs(), testRootCANum)
			require.Len(t, config.ClientCAs.Certs(), 0)
		})

		t.Run("with cert pool", func(t *testing.T) {
			tlsConfig.CertPool = testcert.CertPool(t)
			config, err := tlsConfig.Apply()
			require.NoError(t, err)

			clientCertNum := testCertificateNum + testcert.PrivateClientCertNum
			require.Len(t, config.Certificates, clientCertNum)
			rootCANum := testRootCANum + testcert.PrivateRootCANum
			require.Len(t, config.RootCAs.Certs(), rootCANum)
			require.Len(t, config.ClientCAs.Certs(), 0)
		})
	})

	t.Run("server side", func(t *testing.T) {
		tlsConfig.CertPool = nil
		tlsConfig.ServerSide = true

		t.Run("without cert pool", func(t *testing.T) {
			config, err := tlsConfig.Apply()
			require.NoError(t, err)

			require.Len(t, config.Certificates, testCertificateNum)
			require.Len(t, config.RootCAs.Certs(), 0)
			require.Len(t, config.ClientCAs.Certs(), testClientCANum)
		})

		t.Run("with cert pool", func(t *testing.T) {
			tlsConfig.CertPool = testcert.CertPool(t)
			config, err := tlsConfig.Apply()
			require.NoError(t, err)

			require.Len(t, config.Certificates, testCertificateNum)
			require.Len(t, config.RootCAs.Certs(), 0)
			clientCANum := testClientCANum + testcert.PrivateClientCANum
			require.Len(t, config.ClientCAs.Certs(), clientCANum)
		})
	})

	t.Run("common", func(t *testing.T) {
		tlsConfig.CertPool = nil
		config, err := tlsConfig.Apply()
		require.NoError(t, err)

		testdata := [...]*struct {
			expected interface{}
			actual   interface{}
		}{
			{expected: tls.ClientAuthType(4), actual: config.ClientAuth},
			{expected: "test.com", actual: config.ServerName},
			{expected: []string{"h2", "h2c"}, actual: config.NextProtos},
			{expected: uint16(tls.VersionTLS10), actual: config.MinVersion},
			{expected: uint16(tls.VersionTLS11), actual: config.MaxVersion},
			{expected: []uint16{tls.TLS_RSA_WITH_AES_128_GCM_SHA256}, actual: config.CipherSuites},
			{expected: false, actual: config.InsecureSkipVerify},
		}
		for _, td := range testdata {
			require.Equal(t, td.expected, td.actual)
		}
	})
}

func TestTLSConfig_Apply(t *testing.T) {
	config := TLSConfig{}

	t.Run("invalid certificates", func(t *testing.T) {
		config.Certificates = append(config.Certificates, X509KeyPair{
			Cert: "foo data",
			Key:  "foo data",
		})
		_, err := config.Apply()
		require.Error(t, err)
	})

	t.Run("invalid Root CAs", func(t *testing.T) {
		config.Certificates = nil
		config.RootCAs = []string{"foo data"}
		_, err := config.Apply()
		require.Error(t, err)
	})

	t.Run("invalid Client CAs", func(t *testing.T) {
		config.ServerSide = true
		config.ClientCAs = []string{"foo data"}
		_, err := config.Apply()
		require.Error(t, err)
	})
}
