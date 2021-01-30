package cert

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"project/internal/security"
)

// Pair contains certificate and private key.
type Pair struct {
	Certificate *x509.Certificate
	PrivateKey  interface{}
}

// ASN1 is used to get certificate ASN1 data.
func (p *Pair) ASN1() []byte {
	asn1Data := make([]byte, len(p.Certificate.Raw))
	copy(asn1Data, p.Certificate.Raw)
	return asn1Data
}

// Encode is used to get certificate ASN1 data and encode private key to PKCS8.
func (p *Pair) Encode() ([]byte, []byte) {
	cert := p.ASN1()
	key, err := x509.MarshalPKCS8PrivateKey(p.PrivateKey)
	if err != nil {
		panic(fmt.Sprintf("cert: internal error: %s", err))
	}
	return cert, key
}

// EncodeToPEM is used to encode certificate and private key to PEM data.
func (p *Pair) EncodeToPEM() ([]byte, []byte) {
	cert, key := p.Encode()
	defer func() {
		security.CoverBytes(cert)
		security.CoverBytes(key)
	}()
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	}
	keyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: key,
	}
	return pem.EncodeToMemory(certBlock), pem.EncodeToMemory(keyBlock)
}

// TLSCertificate is used to create tls certificate.
func (p *Pair) TLSCertificate() tls.Certificate {
	var cert tls.Certificate
	cert.Certificate = make([][]byte, 1)
	cert.Certificate[0] = make([]byte, len(p.Certificate.Raw))
	copy(cert.Certificate[0], p.Certificate.Raw)
	cert.PrivateKey = p.PrivateKey
	return cert
}
