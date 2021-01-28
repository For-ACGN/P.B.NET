package testsuite

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/random"
)

// TLSCertificate is used to generate CA ASN1 data, signed certificate.
func TLSCertificate(t testing.TB, ipv4 string) (caASN1 []byte, cPEMBlock, cPriPEMBlock []byte) {
	// generate CA certificate
	caCert := &x509.Certificate{
		SerialNumber: big.NewInt(random.Int63()),
		SubjectKeyId: random.Bytes(4),
		NotBefore:    time.Now().AddDate(0, 0, -1),
		NotAfter:     time.Now().AddDate(0, 0, 1),
	}
	caCert.Subject.CommonName = "testsuite CA " + random.String(4)
	caCert.KeyUsage = x509.KeyUsageCertSign
	caCert.BasicConstraintsValid = true
	caCert.IsCA = true
	caPri, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	caPub := &caPri.PublicKey
	caASN1, err = x509.CreateCertificate(rand.Reader, caCert, caCert, caPub, caPri)
	require.NoError(t, err)
	caCert, err = x509.ParseCertificate(caASN1)
	require.NoError(t, err)

	// sign certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(random.Int63()),
		SubjectKeyId: random.Bytes(4),
		NotBefore:    time.Now().AddDate(0, 0, -1),
		NotAfter:     time.Now().AddDate(0, 0, 1),
	}
	cert.Subject.CommonName = "testsuite certificate " + random.String(4)
	cert.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	cert.DNSNames = []string{"localhost"}
	cert.IPAddresses = []net.IP{net.ParseIP(ipv4), net.ParseIP("::1")}
	cPri, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	cPub := &cPri.PublicKey
	cASN1, err := x509.CreateCertificate(rand.Reader, cert, caCert, cPub, caPri)
	require.NoError(t, err)

	// encode with pem
	cPEMBlock = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cASN1,
	})
	cPriPEMBlock = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(cPri),
	})
	return
}

// TLSConfigPair is used to build server and client *tls.Config.
func TLSConfigPair(t testing.TB, ipv4 string) (server, client *tls.Config) {
	// certificates about server
	caASN1, certPEMBlock, keyPEMBlock := TLSCertificate(t, ipv4)
	caCert, err := x509.ParseCertificate(caASN1)
	require.NoError(t, err)
	tlsCert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	require.NoError(t, err)

	// create tls config
	server = &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{tlsCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}
	client = &tls.Config{MinVersion: tls.VersionTLS12,
		RootCAs: x509.NewCertPool(),
	}
	client.RootCAs.AddCert(caCert)

	// certificates about client
	caASN1, certPEMBlock, keyPEMBlock = TLSCertificate(t, ipv4)
	caCert, err = x509.ParseCertificate(caASN1)
	require.NoError(t, err)
	tlsCert, err = tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	require.NoError(t, err)

	server.ClientCAs = x509.NewCertPool()
	server.ClientCAs.AddCert(caCert)
	client.Certificates = []tls.Certificate{tlsCert}
	return
}
