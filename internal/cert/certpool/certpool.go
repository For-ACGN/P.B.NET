package certpool

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"sync"

	"github.com/pkg/errors"

	"project/internal/cert"
	"project/internal/security"
)

// Pool include all certificates from public and private place.
type Pool struct {
	// public means these certificates are from the common organization,
	// like Let's Encrypt, GlobalSign ...
	pubRootCACerts   []*x509.Certificate
	pubClientCACerts []*x509.Certificate
	pubClientCerts   []*pair

	// private means these certificates are from the Controller or self.
	priRootCACerts   []*pair // only Controller contain the Private Key
	priClientCACerts []*pair // only Controller contain the Private Key
	priClientCerts   []*pair

	rwm sync.RWMutex
}

// NewPool is used to create a new certificate pool.
func NewPool() *Pool {
	security.PaddingMemory()
	defer security.FlushMemory()
	memory := security.NewMemory()
	defer memory.Flush()
	return new(Pool)
}

// AddPublicRootCACert is used to add public root CA certificate.
func (p *Pool) AddPublicRootCACert(cert []byte) error {
	// must copy
	raw := make([]byte, len(cert))
	copy(raw, cert)
	certCopy, err := x509.ParseCertificate(raw)
	if err != nil {
		return errors.Wrap(err, "failed to add public root ca certificate")
	}
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if isCertExist(p.pubRootCACerts, certCopy) {
		return errors.New("this public root ca certificate is already exists")
	}
	p.pubRootCACerts = append(p.pubRootCACerts, certCopy)
	return nil
}

// AddPublicClientCACert is used to add public client CA certificate.
func (p *Pool) AddPublicClientCACert(cert []byte) error {
	// must copy
	raw := make([]byte, len(cert))
	copy(raw, cert)
	certCopy, err := x509.ParseCertificate(raw)
	if err != nil {
		return errors.Wrap(err, "failed to add public client ca certificate")
	}
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if isCertExist(p.pubClientCACerts, certCopy) {
		return errors.New("this public client ca certificate is already exists")
	}
	p.pubClientCACerts = append(p.pubClientCACerts, certCopy)
	return nil
}

// AddPublicClientPair is used to add public client certificate.
func (p *Pool) AddPublicClientPair(cert, pri []byte) error {
	pair, err := loadPair(cert, pri)
	if err != nil {
		return errors.Wrap(err, "failed to add public client certificate")
	}
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if isPairExist(p.pubClientCerts, pair) {
		return errors.New("this public client certificate is already exists")
	}
	p.pubClientCerts = append(p.pubClientCerts, pair)
	return nil
}

// AddPrivateRootCAPair is used to add private root CA certificate with private key.
func (p *Pool) AddPrivateRootCAPair(cert, pri []byte) error {
	pair, err := loadPair(cert, pri)
	if err != nil {
		return errors.Wrap(err, "failed to add private root ca certificate")
	}
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if isPairExist(p.priRootCACerts, pair) {
		return errors.New("this private root ca certificate is already exists")
	}
	p.priRootCACerts = append(p.priRootCACerts, pair)
	return nil
}

// AddPrivateRootCACert is used to add private root CA certificate.
func (p *Pool) AddPrivateRootCACert(cert []byte) error {
	pair, err := loadCertToPair(cert)
	if err != nil {
		return errors.Wrap(err, "failed to add private root ca certificate")
	}
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if isPairExist(p.priRootCACerts, pair) {
		return errors.New("this private root ca certificate is already exists")
	}
	p.priRootCACerts = append(p.priRootCACerts, pair)
	return nil
}

// AddPrivateClientCAPair is used to add private client CA certificate with private key.
func (p *Pool) AddPrivateClientCAPair(cert, pri []byte) error {
	pair, err := loadPair(cert, pri)
	if err != nil {
		return errors.Wrap(err, "failed to add private client ca certificate")
	}
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if isPairExist(p.priClientCACerts, pair) {
		return errors.New("this private client ca certificate is already exists")
	}
	p.priClientCACerts = append(p.priClientCACerts, pair)
	return nil
}

// AddPrivateClientCACert is used to add private client CA certificate with private key.
func (p *Pool) AddPrivateClientCACert(cert []byte) error {
	pair, err := loadCertToPair(cert)
	if err != nil {
		return errors.Wrap(err, "failed to add private client ca certificate")
	}
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if isPairExist(p.priClientCACerts, pair) {
		return errors.New("this private client ca certificate is already exists")
	}
	p.priClientCACerts = append(p.priClientCACerts, pair)
	return nil
}

// AddPrivateClientPair is used to add private client certificate.
func (p *Pool) AddPrivateClientPair(cert, pri []byte) error {
	pair, err := loadPair(cert, pri)
	if err != nil {
		return errors.Wrap(err, "failed to add private client certificate")
	}
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if isPairExist(p.priClientCerts, pair) {
		return errors.New("this private client certificate is already exists")
	}
	p.priClientCerts = append(p.priClientCerts, pair)
	return nil
}

// DeletePublicRootCACert is used to delete public root CA certificate.
func (p *Pool) DeletePublicRootCACert(i int) error {
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if i < 0 || i > len(p.pubRootCACerts)-1 {
		return errors.Errorf("invalid id: %d", i)
	}
	p.pubRootCACerts = append(p.pubRootCACerts[:i], p.pubRootCACerts[i+1:]...)
	return nil
}

// DeletePublicClientCACert is used to delete public client CA certificate.
func (p *Pool) DeletePublicClientCACert(i int) error {
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if i < 0 || i > len(p.pubClientCACerts)-1 {
		return errors.Errorf("invalid id: %d", i)
	}
	p.pubClientCACerts = append(p.pubClientCACerts[:i], p.pubClientCACerts[i+1:]...)
	return nil
}

// DeletePublicClientCert is used to delete public client certificate.
func (p *Pool) DeletePublicClientCert(i int) error {
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if i < 0 || i > len(p.pubClientCerts)-1 {
		return errors.Errorf("invalid id: %d", i)
	}
	p.pubClientCerts = append(p.pubClientCerts[:i], p.pubClientCerts[i+1:]...)
	return nil
}

// DeletePrivateRootCACert is used to delete private root CA certificate.
func (p *Pool) DeletePrivateRootCACert(i int) error {
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if i < 0 || i > len(p.priRootCACerts)-1 {
		return errors.Errorf("invalid id: %d", i)
	}
	p.priRootCACerts = append(p.priRootCACerts[:i], p.priRootCACerts[i+1:]...)
	return nil
}

// DeletePrivateClientCACert is used to delete private client CA certificate.
func (p *Pool) DeletePrivateClientCACert(i int) error {
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if i < 0 || i > len(p.priClientCACerts)-1 {
		return errors.Errorf("invalid id: %d", i)
	}
	p.priClientCACerts = append(p.priClientCACerts[:i], p.priClientCACerts[i+1:]...)
	return nil
}

// DeletePrivateClientCert is used to delete private client certificate.
func (p *Pool) DeletePrivateClientCert(i int) error {
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if i < 0 || i > len(p.priClientCerts)-1 {
		return errors.Errorf("invalid id: %d", i)
	}
	p.priClientCerts = append(p.priClientCerts[:i], p.priClientCerts[i+1:]...)
	return nil
}

// GetPublicRootCACerts is used to get all public root CA certificates.
func (p *Pool) GetPublicRootCACerts() []*x509.Certificate {
	p.rwm.RLock()
	defer p.rwm.RUnlock()
	certs := make([]*x509.Certificate, len(p.pubRootCACerts))
	copy(certs, p.pubRootCACerts)
	return certs
}

// GetPublicClientCACerts is used to get all public client CA certificates.
func (p *Pool) GetPublicClientCACerts() []*x509.Certificate {
	p.rwm.RLock()
	defer p.rwm.RUnlock()
	certs := make([]*x509.Certificate, len(p.pubClientCACerts))
	copy(certs, p.pubClientCACerts)
	return certs
}

// GetPublicClientPairs is used to get all public client certificates.
func (p *Pool) GetPublicClientPairs() []*cert.Pair {
	p.rwm.RLock()
	defer p.rwm.RUnlock()
	l := len(p.pubClientCerts)
	pairs := make([]*cert.Pair, l)
	for i := 0; i < l; i++ {
		pairs[i] = p.pubClientCerts[i].ToCertPair()
	}
	return pairs
}

// GetPrivateRootCAPairs is used to get all private root CA certificates.
func (p *Pool) GetPrivateRootCAPairs() []*cert.Pair {
	p.rwm.RLock()
	defer p.rwm.RUnlock()
	l := len(p.priRootCACerts)
	pairs := make([]*cert.Pair, l)
	for i := 0; i < l; i++ {
		pairs[i] = p.priRootCACerts[i].ToCertPair()
	}
	return pairs
}

// GetPrivateRootCACerts is used to get all private root CA certificates.
func (p *Pool) GetPrivateRootCACerts() []*x509.Certificate {
	p.rwm.RLock()
	defer p.rwm.RUnlock()
	l := len(p.priRootCACerts)
	certs := make([]*x509.Certificate, l)
	for i := 0; i < l; i++ {
		certs[i] = p.priRootCACerts[i].Certificate
	}
	return certs
}

// GetPrivateClientCAPairs is used to get all private client CA certificates.
func (p *Pool) GetPrivateClientCAPairs() []*cert.Pair {
	p.rwm.RLock()
	defer p.rwm.RUnlock()
	l := len(p.priClientCACerts)
	pairs := make([]*cert.Pair, l)
	for i := 0; i < l; i++ {
		pairs[i] = p.priClientCACerts[i].ToCertPair()
	}
	return pairs
}

// GetPrivateClientCACerts is used to get all private client CA certificates.
func (p *Pool) GetPrivateClientCACerts() []*x509.Certificate {
	p.rwm.RLock()
	defer p.rwm.RUnlock()
	l := len(p.priClientCACerts)
	certs := make([]*x509.Certificate, l)
	for i := 0; i < l; i++ {
		certs[i] = p.priClientCACerts[i].Certificate
	}
	return certs
}

// GetPrivateClientPairs is used to get all private client certificates.
func (p *Pool) GetPrivateClientPairs() []*cert.Pair {
	p.rwm.RLock()
	defer p.rwm.RUnlock()
	l := len(p.priClientCerts)
	pairs := make([]*cert.Pair, l)
	for i := 0; i < l; i++ {
		pairs[i] = p.priClientCerts[i].ToCertPair()
	}
	return pairs
}

// ExportPublicRootCACert is used to export public root CA certificate.
func (p *Pool) ExportPublicRootCACert(i int) ([]byte, error) {
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if i < 0 || i > len(p.pubRootCACerts)-1 {
		return nil, errors.Errorf("invalid id: %d", i)
	}
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: p.pubRootCACerts[i].Raw,
	}
	return pem.EncodeToMemory(certBlock), nil
}

// ExportPublicClientCACert is used to export public client CA certificate.
func (p *Pool) ExportPublicClientCACert(i int) ([]byte, error) {
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if i < 0 || i > len(p.pubClientCACerts)-1 {
		return nil, errors.Errorf("invalid id: %d", i)
	}
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: p.pubClientCACerts[i].Raw,
	}
	return pem.EncodeToMemory(certBlock), nil
}

// ExportPublicClientPair is used to export public client CA certificate.
func (p *Pool) ExportPublicClientPair(i int) ([]byte, []byte, error) {
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if i < 0 || i > len(p.pubClientCerts)-1 {
		return nil, nil, errors.Errorf("invalid id: %d", i)
	}
	crt, key := p.pubClientCerts[i].ToCertPair().EncodeToPEM()
	return crt, key, nil
}

// ExportPrivateRootCAPair is used to export private root CA certificate.
func (p *Pool) ExportPrivateRootCAPair(i int) ([]byte, []byte, error) {
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if i < 0 || i > len(p.priRootCACerts)-1 {
		return nil, nil, errors.Errorf("invalid id: %d", i)
	}
	crt, key := p.priRootCACerts[i].ToCertPair().EncodeToPEM()
	return crt, key, nil
}

// ExportPrivateClientCAPair is used to export private client CA certificate.
func (p *Pool) ExportPrivateClientCAPair(i int) ([]byte, []byte, error) {
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if i < 0 || i > len(p.priClientCACerts)-1 {
		return nil, nil, errors.Errorf("invalid id: %d", i)
	}
	crt, key := p.priClientCACerts[i].ToCertPair().EncodeToPEM()
	return crt, key, nil
}

// ExportPrivateClientPair is used to export private client certificate.
func (p *Pool) ExportPrivateClientPair(i int) ([]byte, []byte, error) {
	p.rwm.Lock()
	defer p.rwm.Unlock()
	if i < 0 || i > len(p.priClientCerts)-1 {
		return nil, nil, errors.Errorf("invalid id: %d", i)
	}
	crt, key := p.priClientCerts[i].ToCertPair().EncodeToPEM()
	return crt, key, nil
}

func isCertExist(certs []*x509.Certificate, cert *x509.Certificate) bool {
	for i := 0; i < len(certs); i++ {
		if bytes.Equal(certs[i].Raw, cert.Raw) {
			return true
		}
	}
	return false
}

func isPairExist(pairs []*pair, pair *pair) bool {
	for i := 0; i < len(pairs); i++ {
		if bytes.Equal(pairs[i].Certificate.Raw, pair.Certificate.Raw) {
			return true
		}
	}
	return false
}

// NewPoolWithSystem is used to create a certificate pool with system certificates.
func NewPoolWithSystem() (*Pool, error) {
	certPool, err := System()
	if err != nil {
		return nil, err
	}
	pool := NewPool()
	certs := certPool.Certs()
	for i := 0; i < len(certs); i++ {
		err = pool.AddPublicRootCACert(certs[i].Raw)
		if err != nil {
			return nil, err
		}
	}
	return pool, nil
}
