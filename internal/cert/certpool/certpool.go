// +build !windows

package certpool

import (
	"crypto/x509"
)

// System is used to return certificate pool from system.
func System() (*x509.CertPool, error) {
	return x509.SystemCertPool()
}
