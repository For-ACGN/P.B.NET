package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/pkg/errors"
)

// ErrInvalidPEMBlock is the error about the PEM block.
var ErrInvalidPEMBlock = errors.New("invalid PEM block")

// ParseCertificate is used to parse certificate from PEM.
func ParseCertificate(pemBlock []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemBlock)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}
	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("invalid PEM block type: %s", block.Type)
	}
	return x509.ParseCertificate(block.Bytes)
}

// ParseCertificates is used to parse certificates from PEM.
func ParseCertificates(pemBlock []byte) ([]*x509.Certificate, error) {
	var (
		certs []*x509.Certificate
		block *pem.Block
	)
	for {
		block, pemBlock = pem.Decode(pemBlock)
		if block == nil {
			return nil, ErrInvalidPEMBlock
		}
		if block.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("invalid PEM block type: %s", block.Type)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
		if len(pemBlock) == 0 {
			break
		}
	}
	return certs, nil
}

// ParsePrivateKey is used to parse private key from PEM.
func ParsePrivateKey(pemBlock []byte) (interface{}, error) {
	block, _ := pem.Decode(pemBlock)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}
	return ParsePrivateKeyBytes(block.Bytes)
}

// ParsePrivateKeys is used to parse private keys from PEM.
func ParsePrivateKeys(pemBlock []byte) ([]interface{}, error) {
	var (
		keys  []interface{}
		block *pem.Block
	)
	for {
		block, pemBlock = pem.Decode(pemBlock)
		if block == nil {
			return nil, ErrInvalidPEMBlock
		}
		key, err := ParsePrivateKeyBytes(block.Bytes)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
		if len(pemBlock) == 0 {
			break
		}
	}
	return keys, nil
}

// ParsePrivateKeyBytes is used to parse private key from bytes.
func ParsePrivateKeyBytes(bytes []byte) (interface{}, error) {
	if key, err := x509.ParsePKCS1PrivateKey(bytes); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(bytes); err == nil {
		return key, nil
	}
	if key, err := x509.ParseECPrivateKey(bytes); err == nil {
		return key, nil
	}
	return nil, errors.New("failed to parse private key")
}
