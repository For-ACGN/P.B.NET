// +build windows

package certpool

import (
	"crypto/x509"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

func TestSystem_Windows(t *testing.T) {
	// must clean systemCerts, otherwise it will return a
	// copy of cache if you run other tests in this package.
	func() {
		systemCertsMu.Lock()
		defer systemCertsMu.Unlock()
		systemCerts = nil
	}()

	patch := func() ([]*x509.Certificate, error) {
		return nil, monkey.Error
	}
	pg := monkey.Patch(loadSystemCerts, patch)
	defer pg.Unpatch()

	_, err := System()
	monkey.IsExistMonkeyError(t, err)
}

func TestLoadSystemCerts(t *testing.T) {
	patch := func(string) ([][]byte, error) {
		return nil, monkey.Error
	}
	pg := monkey.Patch(LoadSystemCertsWithName, patch)
	defer pg.Unpatch()

	_, err := loadSystemCerts()
	monkey.IsMonkeyError(t, err)
}

func TestLoadSystemCertsWithName(t *testing.T) {
	t.Run("ROOT and CA", func(t *testing.T) {
		root, err := LoadSystemCertsWithName("ROOT")
		require.NoError(t, err)
		ca, err := LoadSystemCertsWithName("CA")
		require.NoError(t, err)

		certs := append(root, ca...)
		t.Log("raw number:", len(certs))

		count := 0
		for i := 0; i < len(certs); i++ {
			cert, err := x509.ParseCertificate(certs[i])
			if err != nil {
				t.Log(err)
				continue
			}
			count++

			// print CA information
			const format = "V%d %s\n"
			switch {
			case cert.Subject.CommonName != "":
				t.Logf(format, cert.Version, cert.Subject.CommonName)
			case len(cert.Subject.Organization) != 0:
				t.Logf(format, cert.Version, cert.Subject.Organization[0])
			default:
				t.Logf(format, cert.Version, cert.Subject)
			}
		}
		t.Log("actual number:", count)
	})

	t.Run("UTF16PtrFromString", func(t *testing.T) {
		patch := func(string) (*uint16, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(syscall.UTF16PtrFromString, patch)
		defer pg.Unpatch()

		_, err := LoadSystemCertsWithName("ROOT")
		monkey.IsMonkeyError(t, err)
	})

	t.Run("CertOpenSystemStore", func(t *testing.T) {
		patch := func(syscall.Handle, *uint16) (syscall.Handle, error) {
			return 0, monkey.Error
		}
		pg := monkey.Patch(syscall.CertOpenSystemStore, patch)
		defer pg.Unpatch()

		_, err := LoadSystemCertsWithName("ROOT")
		monkey.IsMonkeyError(t, err)
	})

	t.Run("CertEnumCertificatesInStore error", func(t *testing.T) {
		patch := func(syscall.Handle, *syscall.CertContext) (*syscall.CertContext, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(syscall.CertEnumCertificatesInStore, patch)
		defer pg.Unpatch()

		_, err := LoadSystemCertsWithName("ROOT")
		monkey.IsMonkeyError(t, err)
	})

	t.Run("CertEnumCertificatesInStore nil", func(t *testing.T) {
		patch := func(syscall.Handle, *syscall.CertContext) (*syscall.CertContext, error) {
			return nil, nil
		}
		pg := monkey.Patch(syscall.CertEnumCertificatesInStore, patch)
		defer pg.Unpatch()

		certs, err := LoadSystemCertsWithName("ROOT")
		require.NoError(t, err)
		require.Empty(t, certs)
	})
}
