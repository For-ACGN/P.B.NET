package cert

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

func testGenerateOptions() *Options {
	opts := Options{
		DNSNames:       []string{"test.com", "foo.com"},
		IPAddresses:    []string{"1.1.1.1", "1234::1234"},
		EmailAddresses: []string{"admin@test.com", "user@test.com"},
		URLs:           []string{"https://1.1.1.1/", "http://example.com/"},
	}
	opts.Subject.CommonName = "test"
	opts.Subject.Organization = []string{"org a", "org b"}
	opts.Subject.OrganizationalUnit = []string{"unit a", "unit b"}
	opts.Subject.Country = []string{"country a", "country b"}
	opts.Subject.Locality = []string{"locality a", "locality b"}
	opts.Subject.Province = []string{"province a", "province b"}
	opts.Subject.StreetAddress = []string{"street address a", "street address b"}
	opts.Subject.PostalCode = []string{"postal code a", "postal code b"}
	opts.Subject.SerialNumber = "12345678"
	return &opts
}

func TestDump(t *testing.T) {
	opts := testGenerateOptions()

	t.Run("rsa", func(t *testing.T) {
		opts.Algorithm = "rsa|2048"

		ca, err := GenerateCA(opts)
		require.NoError(t, err)
		ca.Certificate.AuthorityKeyId = ca.Certificate.SubjectKeyId

		Dump(ca.Certificate)
	})

	t.Run("ecdsa", func(t *testing.T) {
		opts.Algorithm = "ecdsa|p256"

		ca, err := GenerateCA(opts)
		require.NoError(t, err)
		ca.Certificate.AuthorityKeyId = ca.Certificate.SubjectKeyId

		Dump(ca.Certificate)
	})

	t.Run("ed25519", func(t *testing.T) {
		opts.Algorithm = "ed25519"

		ca, err := GenerateCA(opts)
		require.NoError(t, err)
		ca.Certificate.AuthorityKeyId = ca.Certificate.SubjectKeyId

		Dump(ca.Certificate)
	})
}

func TestSdump(t *testing.T) {
	opts := testGenerateOptions()

	ca, err := GenerateCA(opts)
	require.NoError(t, err)

	output := Sdump(ca.Certificate)
	fmt.Println(output)
}

func TestFdump(t *testing.T) {
	opts := testGenerateOptions()
	ca, err := GenerateCA(opts)
	require.NoError(t, err)
	buf := bytes.NewBuffer(make([]byte, 0, 512))

	t.Run("common", func(t *testing.T) {
		n, err := Fdump(buf, ca.Certificate)
		require.NoError(t, err)
		require.Equal(t, n, buf.Len())

		fmt.Println(buf)
	})

	t.Run("empty alternate", func(t *testing.T) {
		pair, err := GenerateCA(nil)
		require.NoError(t, err)

		buf.Reset()

		n, err := Fdump(buf, pair.Certificate)
		require.NoError(t, err)
		require.Equal(t, n, buf.Len())

		fmt.Println(buf)
	})

	t.Run("empty key usage", func(t *testing.T) {
		ku := ca.Certificate.KeyUsage
		defer func() { ca.Certificate.KeyUsage = ku }()
		ca.Certificate.KeyUsage = 0

		buf.Reset()

		n, err := Fdump(buf, ca.Certificate)
		require.NoError(t, err)
		require.Equal(t, n, buf.Len())

		fmt.Println(buf)
	})

	t.Run("empty serial number", func(t *testing.T) {
		ku := ca.Certificate.SerialNumber
		defer func() { ca.Certificate.SerialNumber = ku }()
		ca.Certificate.SerialNumber = new(big.Int)

		buf.Reset()

		n, err := Fdump(buf, ca.Certificate)
		require.NoError(t, err)
		require.Equal(t, n, buf.Len())

		fmt.Println(buf)
	})

	t.Run("failed to dump public key", func(t *testing.T) {
		publicKey := ca.Certificate.PublicKey
		defer func() { ca.Certificate.PublicKey = publicKey }()
		ca.Certificate.PublicKey = nil

		buf.Reset()

		_, err = Fdump(buf, ca.Certificate)
		require.Error(t, err)

		fmt.Println(buf)
	})

	t.Run("failed to dump basic", func(t *testing.T) {
		var pg *monkey.PatchGuard
		patch := func(w io.Writer, format string, a ...interface{}) (int, error) {
			if strings.Contains(format, "Version: %d") {
				return 0, monkey.Error
			}
			pg.Unpatch()
			defer pg.Restore()
			return fmt.Fprintf(w, format, a...)
		}
		pg = monkey.Patch(fmt.Fprintf, patch)
		defer pg.Unpatch()

		buf.Reset()

		_, err = Fdump(buf, ca.Certificate)
		monkey.IsMonkeyError(t, err)

		fmt.Println(buf)
	})

	t.Run("failed to dump alternate", func(t *testing.T) {
		var pg *monkey.PatchGuard
		patch := func(io.Writer, ...interface{}) (int, error) {
			return 0, monkey.Error
		}
		pg = monkey.Patch(fmt.Fprint, patch)
		defer pg.Unpatch()

		buf.Reset()

		_, err = Fdump(buf, ca.Certificate)
		monkey.IsMonkeyError(t, err)

		fmt.Println(buf)
	})

	t.Run("failed to dump dns names", func(t *testing.T) {
		var pg *monkey.PatchGuard
		patch := func(w io.Writer, format string, a ...interface{}) (int, error) {
			if strings.Contains(format, "DNS names") {
				return 0, monkey.Error
			}
			pg.Unpatch()
			defer pg.Restore()
			return fmt.Fprintf(w, format, a...)
		}
		pg = monkey.Patch(fmt.Fprintf, patch)
		defer pg.Unpatch()

		buf.Reset()

		_, err = Fdump(buf, ca.Certificate)
		monkey.IsMonkeyError(t, err)

		fmt.Println(buf)
	})

	t.Run("failed to dump ip address", func(t *testing.T) {
		var pg *monkey.PatchGuard
		patch := func(w io.Writer, format string, a ...interface{}) (int, error) {
			if strings.Contains(format, "IP addresses") {
				return 0, monkey.Error
			}
			pg.Unpatch()
			defer pg.Restore()
			return fmt.Fprintf(w, format, a...)
		}
		pg = monkey.Patch(fmt.Fprintf, patch)
		defer pg.Unpatch()

		buf.Reset()

		_, err = Fdump(buf, ca.Certificate)
		monkey.IsMonkeyError(t, err)

		fmt.Println(buf)
	})

	t.Run("failed to dump email addresses", func(t *testing.T) {
		var pg *monkey.PatchGuard
		patch := func(w io.Writer, format string, a ...interface{}) (int, error) {
			if strings.Contains(format, "Email addresses") {
				return 0, monkey.Error
			}
			pg.Unpatch()
			defer pg.Restore()
			return fmt.Fprintf(w, format, a...)
		}
		pg = monkey.Patch(fmt.Fprintf, patch)
		defer pg.Unpatch()

		buf.Reset()

		_, err = Fdump(buf, ca.Certificate)
		monkey.IsMonkeyError(t, err)

		fmt.Println(buf)
	})

	t.Run("failed to dump uris", func(t *testing.T) {
		var pg *monkey.PatchGuard
		patch := func(w io.Writer, format string, a ...interface{}) (int, error) {
			if strings.Contains(format, "URIs") {
				return 0, monkey.Error
			}
			pg.Unpatch()
			defer pg.Restore()
			return fmt.Fprintf(w, format, a...)
		}
		pg = monkey.Patch(fmt.Fprintf, patch)
		defer pg.Unpatch()

		buf.Reset()

		_, err = Fdump(buf, ca.Certificate)
		monkey.IsMonkeyError(t, err)

		fmt.Println(buf)
	})
}

func TestCalcMaxPaddingLen(t *testing.T) {
	t.Run("only DNS names", func(t *testing.T) {
		cert := &x509.Certificate{
			DNSNames: []string{"foo"},
		}

		l := calcMaxPaddingLen(cert)
		require.Equal(t, len("DNS names"), l)
	})

	t.Run("only IP addresses", func(t *testing.T) {
		cert := &x509.Certificate{
			IPAddresses: []net.IP{nil},
		}

		l := calcMaxPaddingLen(cert)
		require.Equal(t, len("IP addresses"), l)
	})

	t.Run("only email addresses", func(t *testing.T) {
		cert := &x509.Certificate{
			EmailAddresses: []string{"foo"},
		}

		l := calcMaxPaddingLen(cert)
		require.Equal(t, len("Email addresses"), l)
	})

	t.Run("only URIs", func(t *testing.T) {
		cert := &x509.Certificate{
			URIs: []*url.URL{nil},
		}

		l := calcMaxPaddingLen(cert)
		require.Equal(t, len("URIs"), l)
	})
}
