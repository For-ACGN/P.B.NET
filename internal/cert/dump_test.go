package cert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
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

}
