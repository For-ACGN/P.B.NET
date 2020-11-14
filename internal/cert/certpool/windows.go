// +build windows

package certpool

import (
	"crypto/x509"
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

var (
	systemCerts   []*x509.Certificate
	systemCertsMu sync.Mutex
)

// System is used to return certificate pool from system.
// On windows, the number of the CA and ROOT Certificate is
// incorrect, because the CA "Root Agency" is used to test.
func System() (*x509.CertPool, error) {
	var certs []*x509.Certificate
	systemCertsMu.Lock()
	defer systemCertsMu.Unlock()
	l := len(systemCerts)
	if l != 0 {
		certs = make([]*x509.Certificate, l)
		copy(certs, systemCerts)
	} else {
		c, err := loadSystemCerts()
		if err != nil {
			return nil, fmt.Errorf("failed to load system certificate pool: %s", err)
		}
		certs = make([]*x509.Certificate, len(c))
		copy(certs, c)
		systemCerts = c
	}
	// must new pool
	pool := x509.NewCertPool()
	for i := 0; i < len(certs); i++ {
		pool.AddCert(certs[i])
	}
	return pool, nil
}

// reference:
// https://docs.microsoft.com/en-us/windows/win32/api/wincrypt/nf-wincrypt-certopensystemstorea

func loadSystemCerts() ([]*x509.Certificate, error) {
	var certs [][]byte
	names := []string{"ROOT", "CA", "MY"}
	for i := 0; i < len(names); i++ {
		raw, err := LoadSystemCertsWithName(names[i])
		if err != nil {
			return nil, err
		}
		certs = append(certs, raw...)
	}
	var pool []*x509.Certificate
	for i := 0; i < len(certs); i++ {
		cert, err := x509.ParseCertificate(certs[i])
		if err == nil {
			pool = append(pool, cert)
		}
	}
	return pool, nil
}

// LoadSystemCertsWithName is used to load system certificates by name.
// Usually name is "ROOT" or "CA".
func LoadSystemCertsWithName(name string) ([][]byte, error) {
	n, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}
	store, err := syscall.CertOpenSystemStore(0, n)
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
				if errno == 0x80092004 { // 0x80092004 is CRYPT_E_NOT_FOUND
					break
				}
			}
			return nil, err
		}
		if cert == nil {
			break
		}
		// copy the buf, since ParseCertificate does not create its own copy.
		buf := (*[1024 * 1024]byte)(unsafe.Pointer(cert.EncodedCert))[:] // #nosec
		buf2 := make([]byte, cert.Length)
		copy(buf2, buf)
		certs = append(certs, buf2)
	}
	return certs, nil
}
