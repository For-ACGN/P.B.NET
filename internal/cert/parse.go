package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/pkg/errors"
)

// ErrInvalidPEMBlock is the error about the PEM block.
var ErrInvalidPEMBlock = errors.New("invalid PEM block")

// ParseCertificateDER is used to parse certificate from the given ASN.1 DER data.
func ParseCertificateDER(der []byte) (*x509.Certificate, error) {
	return x509.ParseCertificate(der)
}

// ParseCertificatePEM is used to parse certificate from the PEM data.
func ParseCertificatePEM(pb []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pb)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}
	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("invalid PEM block type: %s", block.Type)
	}
	return ParseCertificateDER(block.Bytes)
}

// ParseCertificatesPEM is used to parse certificates from the PEM data.
func ParseCertificatesPEM(pb []byte) ([]*x509.Certificate, error) {
	var (
		certs []*x509.Certificate
		block *pem.Block
	)
	for {
		block, pb = pem.Decode(pb)
		if block == nil {
			return nil, ErrInvalidPEMBlock
		}
		if block.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("invalid PEM block type: %s", block.Type)
		}
		cert, err := ParseCertificateDER(block.Bytes)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
		if len(pb) == 0 {
			break
		}
	}
	return certs, nil
}

// ParsePrivateKeyDER is used to parse private key from the given ASN.1 DER data.
func ParsePrivateKeyDER(der []byte) (interface{}, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}
	return nil, errors.New("failed to parse private key")
}

// ParsePrivateKeyPEM is used to parse private key from the PEM data.
func ParsePrivateKeyPEM(pb []byte) (interface{}, error) {
	block, _ := pem.Decode(pb)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}
	return ParsePrivateKeyDER(block.Bytes)
}

// ParsePrivateKeysPEM is used to parse private keys from the PEM data.
func ParsePrivateKeysPEM(pb []byte) ([]interface{}, error) {
	var (
		keys  []interface{}
		block *pem.Block
	)
	for {
		block, pb = pem.Decode(pb)
		if block == nil {
			return nil, ErrInvalidPEMBlock
		}
		key, err := ParsePrivateKeyDER(block.Bytes)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
		if len(pb) == 0 {
			break
		}
	}
	return keys, nil
}
