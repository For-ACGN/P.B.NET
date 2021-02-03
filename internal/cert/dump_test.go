package cert

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

func TestDump(t *testing.T) {
	opts := Options{
		DNSNames:       []string{"test.com", "foo.com"},
		IPAddresses:    []string{"1.1.1.1", "1234::1234"},
		EmailAddresses: []string{"admin@test.com", "user@test.com"},
		URLs:           []string{"https://1.1.1.1/", "http://example.com/"},
	}
	opts.Subject.Organization = []string{"org a", "org b"}

	t.Run("rsa", func(t *testing.T) {
		opts.Algorithm = "rsa|2048"

		ca, err := GenerateCA(&opts)
		require.NoError(t, err)

		Dump(ca.Certificate)
	})

	t.Run("ecdsa", func(t *testing.T) {
		opts.Algorithm = "ecdsa|p256"

		ca, err := GenerateCA(&opts)
		require.NoError(t, err)

		Dump(ca.Certificate)
	})

	t.Run("ed25519", func(t *testing.T) {
		opts.Algorithm = "ed25519"

		ca, err := GenerateCA(&opts)
		require.NoError(t, err)

		Dump(ca.Certificate)
	})
}

func TestSdump(t *testing.T) {
	opts := Options{
		DNSNames:       []string{"test.com", "foo.com"},
		IPAddresses:    []string{"1.1.1.1", "1234::1234"},
		EmailAddresses: []string{"admin@test.com", "user@test.com"},
		URLs:           []string{"https://1.1.1.1/", "http://example.com/"},
	}
	opts.Subject.Organization = []string{"org a", "org b"}

	ca, err := GenerateCA(&opts)
	require.NoError(t, err)

	output := Sdump(ca.Certificate)
	fmt.Println(output)
}

func TestFdump(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 512))

	t.Run("empty info", func(t *testing.T) {
		ca, err := GenerateCA(nil)
		require.NoError(t, err)

		buf.Reset()

		n, err := Fdump(buf, ca.Certificate)
		require.NoError(t, err)
		require.Equal(t, n, buf.Len())

		fmt.Println(buf)
	})

	opts := Options{
		DNSNames:       []string{"test.com", "foo.com"},
		IPAddresses:    []string{"1.1.1.1", "1234::1234"},
		EmailAddresses: []string{"admin@test.com", "user@test.com"},
		URLs:           []string{"https://1.1.1.1/", "http://example.com/"},
	}
	opts.Subject.Organization = []string{"org a", "org b"}

	ca, err := GenerateCA(&opts)
	require.NoError(t, err)

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
