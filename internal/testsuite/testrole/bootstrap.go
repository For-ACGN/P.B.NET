package testrole

import (
	"os"

	"github.com/stretchr/testify/require"

	"project/internal/bootstrap"
	"project/internal/messages"
)

// Bootstrap is used to provide test bootstrap
func Bootstrap(t require.TestingT) []*messages.Bootstrap {
	var bootstraps []*messages.Bootstrap
	// http
	config, err := os.ReadFile("../internal/bootstrap/testdata/http.toml")
	require.NoError(t, err)
	boot := &messages.Bootstrap{
		Tag:    "http",
		Mode:   bootstrap.ModeHTTP,
		Config: config,
	}
	bootstraps = append(bootstraps, boot)
	// dns
	config, err = os.ReadFile("../internal/bootstrap/testdata/dns.toml")
	require.NoError(t, err)
	boot = &messages.Bootstrap{
		Tag:    "dns",
		Mode:   bootstrap.ModeDNS,
		Config: config,
	}
	bootstraps = append(bootstraps, boot)
	// direct
	config, err = os.ReadFile("../internal/bootstrap/testdata/direct.toml")
	require.NoError(t, err)
	boot = &messages.Bootstrap{
		Tag:    "direct",
		Mode:   bootstrap.ModeDirect,
		Config: config,
	}
	bootstraps = append(bootstraps, boot)
	return bootstraps
}
