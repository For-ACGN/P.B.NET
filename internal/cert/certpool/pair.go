package certpool

import (
	"crypto/x509"
	"fmt"

	"github.com/pkg/errors"

	"project/internal/cert"
	"project/internal/security"
)

// pair is used to protect private key about certificate.
type pair struct {
	Certificate *x509.Certificate
	PrivateKey  *security.Bytes // PKCS8
}

// ToCertPair is used to convert *pair to *cert.Pair.
func (p *pair) ToCertPair() *cert.Pair {
	pkcs8 := p.PrivateKey.Get()
	defer p.PrivateKey.Put(pkcs8)
	pri, err := x509.ParsePKCS8PrivateKey(pkcs8)
	if err != nil {
		panic(fmt.Sprintf("certpool: internal error: %s", err))
	}
	return &cert.Pair{Certificate: p.Certificate, PrivateKey: pri}
}

func loadPair(crt, pri []byte) (*pair, error) {
	if len(crt) == 0 {
		return nil, errors.New("empty certificate data")
	}
	if len(pri) == 0 {
		return nil, errors.New("empty private key data")
	}
	raw := make([]byte, len(crt))
	copy(raw, crt)
	certCp, err := cert.ParseCertificateDER(raw)
	if err != nil {
		return nil, err
	}
	privateKey, err := cert.ParsePrivateKeyDER(pri)
	if err != nil {
		return nil, err
	}
	if !cert.IsMatchPrivateKey(certCp, privateKey) {
		return nil, errors.New("private key in certificate is not matched")
	}
	priBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return &pair{
		Certificate: certCp,
		PrivateKey:  security.NewBytes(priBytes),
	}, nil
}

func loadCertToPair(cert []byte) (*pair, error) {
	if len(cert) == 0 {
		return nil, errors.New("empty certificate data")
	}
	raw := make([]byte, len(cert))
	copy(raw, cert)
	certCopy, err := x509.ParseCertificate(raw)
	if err != nil {
		return nil, err
	}
	return &pair{Certificate: certCopy}, nil
}
