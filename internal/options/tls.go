package options

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"project/internal/security"
)

var (
	ErrInvalidPEMBlock     = errors.New("invalid PEM block")
	ErrInvalidPEMBlockType = errors.New("invalid PEM block type")
)

type TLSConfig struct {
	Certificates       []X509KeyPair `toml:"certificates"` // tls.X509KeyPair
	RootCAs            []string      `toml:"root_ca"`      // PEM
	ClientCAs          []string      `toml:"client_ca"`    // PEM
	NextProtos         []string      `toml:"next_protos"`
	InsecureSkipVerify bool          `toml:"insecure_skip_verify"`
}

type X509KeyPair struct {
	Cert string `toml:"cert"` // PEM
	Key  string `toml:"key"`  // PEM
}

func (t *TLSConfig) failed(err error) error {
	return fmt.Errorf("failed to apply tls config: %s", err)
}

func (t *TLSConfig) Apply() (*tls.Config, error) {
	nextProtos := make([]string, len(t.NextProtos))
	copy(nextProtos, t.NextProtos)
	config := &tls.Config{
		NextProtos:         nextProtos,
		InsecureSkipVerify: t.InsecureSkipVerify,
	}
	l := len(t.Certificates)
	if l != 0 {
		config.Certificates = make([]tls.Certificate, l)
		for i := 0; i < l; i++ {
			c := []byte(t.Certificates[i].Cert)
			k := []byte(t.Certificates[i].Key)
			tlsCert, err := tls.X509KeyPair(c, k)
			if err != nil {
				return nil, t.failed(err)
			}
			security.FlushBytes(c)
			security.FlushBytes(k)
			config.Certificates[i] = tlsCert
		}
	}
	l = len(t.RootCAs)
	if l != 0 {
		config.RootCAs = x509.NewCertPool()
		for i := 0; i < l; i++ {
			cert, err := parseCertificate([]byte(t.RootCAs[i]))
			if err != nil {
				return nil, t.failed(err)
			}
			config.RootCAs.AddCert(cert)
		}
	}
	l = len(t.ClientCAs)
	if l != 0 {
		config.ClientCAs = x509.NewCertPool()
		for i := 0; i < l; i++ {
			cert, err := parseCertificate([]byte(t.ClientCAs[i]))
			if err != nil {
				return nil, t.failed(err)
			}
			config.ClientCAs.AddCert(cert)
		}
	}
	return config, nil
}

func parseCertificate(cert []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(cert)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}
	if block.Type != "CERTIFICATE" {
		return nil, ErrInvalidPEMBlockType
	}
	return x509.ParseCertificate(block.Bytes)
}
