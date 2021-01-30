package cert

import (
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

	ca, err := GenerateCA(&opts)
	require.NoError(t, err)

	Dump(ca.Certificate)
}
