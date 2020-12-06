package nmap

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/toml"
	"project/internal/testsuite"
)

func TestOptions_ToArgs(t *testing.T) {
	opts := Options{
		NoPing:        true,
		Service:       true,
		OS:            true,
		Device:        "eth0",
		LocalIP:       []string{"192.168.1.2"},
		HostTimeout:   "10m",
		MaxRTTTimeout: "50ms",
		MinRate:       1000,
		MaxRate:       10000,
		Arguments:     "--ttl 128 --badsum",
	}
	const cmd = "-Pn -sV -O -e eth0 -S 192.168.1.2 " +
		"--host-timeout 10m --max-rtt-timeout 50ms " +
		"--min-rate 1000 --max-rate 10000 --ttl 128 --badsum"
	require.Equal(t, cmd, opts.String())
}

func TestOptions(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/options.toml")
	require.NoError(t, err)

	// check unnecessary field
	opts := Options{}
	err = toml.Unmarshal(data, &opts)
	require.NoError(t, err)

	// check zero value
	testsuite.ContainZeroValue(t, opts)

	for _, testdata := range [...]*struct {
		expected interface{}
		actual   interface{}
	}{
		{expected: true, actual: opts.NoPing},
		{expected: true, actual: opts.Service},
		{expected: true, actual: opts.OS},
		{expected: "eth0", actual: opts.Device},
		{expected: []string{"192.168.1.2"}, actual: opts.LocalIP},
		{expected: "10m", actual: opts.HostTimeout},
		{expected: "50ms", actual: opts.MaxRTTTimeout},
		{expected: 1000, actual: opts.MinRate},
		{expected: 10000, actual: opts.MaxRate},
		{expected: "--ttl 128 --badsum", actual: opts.Arguments},
	} {
		require.Equal(t, testdata.expected, testdata.actual)
	}
}
