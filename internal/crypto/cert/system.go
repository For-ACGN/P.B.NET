// +build windows

package cert

import (
	"crypto/x509"
	"errors"
	"sync"
	"syscall"
	"unsafe"
)

var (
	systemCert      []*x509.Certificate
	systemCertErr   = errors.New("no system certificates")
	systemCertMutex sync.Mutex
)

func systemCertPool() (*x509.CertPool, error) {
	var certs []*x509.Certificate
	systemCertMutex.Lock()
	if systemCertErr == nil {
		systemCertMutex.Unlock()
		certs = make([]*x509.Certificate, len(systemCert))
		copy(certs, systemCert)
	} else {
		c, err := loadSystemCert()
		if err != nil {
			systemCertMutex.Unlock()
			return nil, err
		}
		systemCert = c
		systemCertErr = nil
		systemCertMutex.Unlock()
		certs = make([]*x509.Certificate, len(systemCert))
		copy(certs, systemCert)
	}
	// must new pool
	pool := x509.NewCertPool()
	for i := 0; i < len(certs); i++ {
		pool.AddCert(certs[i])
	}
	return pool, nil
}

func loadSystemCert() ([]*x509.Certificate, error) {
	root, err := loadSystemCertWithName("ROOT")
	if err != nil {
		return nil, err
	}
	ca, err := loadSystemCertWithName("CA")
	if err != nil {
		return nil, err
	}
	var pool []*x509.Certificate
	certs := append(root, ca...)
	for i := 0; i < len(certs); i++ {
		cert, err := x509.ParseCertificate(certs[i])
		if err == nil {
			pool = append(pool, cert)
		}
	}
	return pool, nil
}

func loadSystemCertWithName(n string) ([][]byte, error) {
	name, err := syscall.UTF16PtrFromString(n)
	if err != nil {
		return nil, err
	}
	store, err := syscall.CertOpenSystemStore(0, name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = syscall.CertCloseStore(store, 0) }()
	var certs [][]byte
	var cert *syscall.CertContext
	for {
		cert, err = syscall.CertEnumCertificatesInStore(store, cert)
		if err != nil {
			if errno, ok := err.(syscall.Errno); ok {
				// CRYPT_E_NOT_FOUND = 0x80092004
				if errno == 0x80092004 {
					break
				}
			}
			return nil, err
		}
		if cert == nil {
			break
		}
		// Copy the buf, since ParseCertificate does not create its own copy.
		buf := (*[1 << 20]byte)(unsafe.Pointer(cert.EncodedCert))[:]
		buf2 := make([]byte, cert.Length)
		copy(buf2, buf)
		certs = append(certs, buf2)
	}
	return certs, nil
}
