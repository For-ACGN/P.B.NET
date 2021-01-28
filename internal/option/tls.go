package option

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"project/internal/cert"
	"project/internal/crypto/rand"
	"project/internal/security"
)

// TLSConfig contains options about tls.Config.
type TLSConfig struct {
	// Certificates contains one or more certificate chains to present to
	// the other side of the connection. The first certificate compatible
	// with the peer's requirements is selected automatically.
	//
	// Server configurations must set one of Certificates, clients doing
	// client authentication may set either Certificates.
	//
	// Note: if there are multiple Certificates, and they don't have the
	// optional field Leaf set, certificate selection will incur a significant
	// per-handshake performance cost.
	Certificates []X509KeyPair `toml:"certificates"`

	// RootCAs defines the set of root certificate authorities that clients
	// use when verifying server certificates. Certificates encoded by pem.
	RootCAs []string `toml:"root_ca"`

	// ClientCAs defines the set of root certificate authorities that servers
	// use if required to verify a client certificate by the policy in ClientAuth.
	// Certificates encoded by pem.
	ClientCAs []string `toml:"client_ca"`

	// ClientAuth determines the server's policy for TLS Client Authentication.
	// The default is NoClientCert.
	ClientAuth tls.ClientAuthType `toml:"client_auth"`

	// ServerName is used to verify the hostname on the returned certificates
	// unless InsecureSkipVerify is given. It is also included in the client's
	// handshake to support virtual hosting unless it is an IP address.
	ServerName string `toml:"server_name"`

	// NextProtos is a list of supported application level protocols, in
	// of preference.
	NextProtos []string `toml:"next_protos"`

	// MinVersion contains the minimum TLS version that is acceptable. If zero,
	//  TLS 1.0 is currently taken as the minimum.
	MinVersion uint16 `toml:"min_version"`

	// MaxVersion contains the maximum TLS version that is acceptable. If zero,
	// the maximum version supported by this package is used.
	MaxVersion uint16 `toml:"max_version"`

	// CipherSuites is a list of supported cipher suites for TLS versions
	// up to TLS 1.2. If CipherSuites is nil, a default list of secure cipher
	// suites is used, with a preference order based on hardware performance.
	// The default cipher suites might change over Go versions.
	CipherSuites []uint16 `toml:"cipher_suites"`

	// CertPoolConfig is used to add certificates from certificate pool manually.
	// Public will be loaded automatically and Private need be loaded manually.
	CertPoolConfig struct {
		SkipPublicRootCA      bool `toml:"skip_public_root_ca"`
		SkipPublicClientCA    bool `toml:"skip_public_client_ca"`
		SkipPublicClientCert  bool `toml:"skip_public_client_cert"`
		LoadPrivateRootCA     bool `toml:"load_private_root_ca"`
		LoadPrivateClientCA   bool `toml:"load_private_client_ca"`
		LoadPrivateClientCert bool `toml:"load_private_client_cert"`
	} `toml:"cert_pool"`

	// CertPool is the certificate pool.
	CertPool *cert.Pool `toml:"-" msgpack:"-" testsuite:"-"`

	// ServerSide is used to mark this configuration for server side, like
	// listeners or http server need set it true for GetCertificates().
	ServerSide bool `toml:"-" msgpack:"-" testsuite:"-"`
}

// X509KeyPair contain certificate and private key encoded by pem.
type X509KeyPair struct {
	Cert string `toml:"cert"`
	Key  string `toml:"key"`
}

// GetCertificates is used to make TLS certificates.
func (t *TLSConfig) GetCertificates() ([]tls.Certificate, error) {
	var certs []tls.Certificate
	for i := 0; i < len(t.Certificates); i++ {
		c := []byte(t.Certificates[i].Cert)
		k := []byte(t.Certificates[i].Key)
		tlsCert, err := tls.X509KeyPair(c, k)
		if err != nil {
			return nil, err
		}
		security.CoverBytes(c)
		security.CoverBytes(k)
		certs = append(certs, tlsCert)
	}
	if t.CertPool == nil {
		return certs, nil
	}
	if !t.ServerSide && !t.CertPoolConfig.SkipPublicClientCert {
		pairs := t.CertPool.GetPublicClientPairs()
		certs = append(certs, makeTLSCertificates(pairs)...)
	}
	if !t.ServerSide && t.CertPoolConfig.LoadPrivateClientCert {
		pairs := t.CertPool.GetPrivateClientPairs()
		certs = append(certs, makeTLSCertificates(pairs)...)
	}
	return certs, nil
}

func makeTLSCertificates(pairs []*cert.Pair) []tls.Certificate {
	l := len(pairs)
	clientCerts := make([]tls.Certificate, l)
	for i := 0; i < l; i++ {
		clientCerts[i] = pairs[i].TLSCertificate()
	}
	return clientCerts
}

// GetRootCAs is used to parse TLSConfig.RootCAs.
func (t *TLSConfig) GetRootCAs() ([]*x509.Certificate, error) {
	if t.ServerSide {
		return nil, nil
	}
	rootCAs, err := t.parseCertificates(t.RootCAs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse root ca: %s", err)
	}
	if t.CertPool == nil {
		return rootCAs, nil
	}
	if !t.CertPoolConfig.SkipPublicRootCA {
		rootCAs = append(rootCAs, t.CertPool.GetPublicRootCACerts()...)
	}
	if t.CertPoolConfig.LoadPrivateRootCA {
		rootCAs = append(rootCAs, t.CertPool.GetPrivateRootCACerts()...)
	}
	return rootCAs, nil
}

// GetClientCAs is used to parse TLSConfig.ClientCAs.
func (t *TLSConfig) GetClientCAs() ([]*x509.Certificate, error) {
	if !t.ServerSide {
		return nil, nil
	}
	clientCAs, err := t.parseCertificates(t.ClientCAs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse client ca: %s", err)
	}
	if t.CertPool == nil {
		return clientCAs, nil
	}
	if !t.CertPoolConfig.SkipPublicClientCA {
		clientCAs = append(clientCAs, t.CertPool.GetPublicClientCACerts()...)
	}
	if t.CertPoolConfig.LoadPrivateClientCA {
		clientCAs = append(clientCAs, t.CertPool.GetPrivateClientCACerts()...)
	}
	return clientCAs, nil
}

func (t *TLSConfig) parseCertificates(pem []string) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	for _, p := range pem {
		c, err := cert.ParseCertificates([]byte(p))
		if err != nil {
			return nil, err
		}
		certs = append(certs, c...)
	}
	return certs, nil
}

// Apply is used to create *tls.Config.
func (t *TLSConfig) Apply() (*tls.Config, error) {
	config := tls.Config{
		Rand:       rand.Reader,
		ServerName: t.ServerName,
		MinVersion: t.MinVersion,
		MaxVersion: t.MaxVersion,
		ClientAuth: t.ClientAuth,
	}
	// set certificates
	certs, err := t.GetCertificates()
	if err != nil {
		return nil, t.error(err)
	}
	config.Certificates = certs
	// set root CAs
	rootCAs, err := t.GetRootCAs()
	if err != nil {
		return nil, t.error(err)
	}
	config.RootCAs = x509.NewCertPool()
	for i := 0; i < len(rootCAs); i++ {
		config.RootCAs.AddCert(rootCAs[i])
	}
	// set client CAs
	clientCAs, err := t.GetClientCAs()
	if err != nil {
		return nil, t.error(err)
	}
	config.ClientCAs = x509.NewCertPool()
	for i := 0; i < len(clientCAs); i++ {
		config.ClientCAs.AddCert(clientCAs[i])
	}
	// set next protocols
	l := len(t.NextProtos)
	if l > 0 {
		config.NextProtos = make([]string, l)
		copy(config.NextProtos, t.NextProtos)
	}
	// set cipher suites
	l = len(t.CipherSuites)
	if l > 0 {
		config.CipherSuites = make([]uint16, l)
		copy(config.CipherSuites, t.CipherSuites)
	}
	// set default minimum version
	if config.MinVersion == 0 {
		config.MinVersion = tls.VersionTLS12
	}
	return &config, nil
}

func (t *TLSConfig) error(err error) error {
	return fmt.Errorf("failed to apply tls configuration: %s", err)
}
