package nmap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJob_ToArgs(t *testing.T) {
	t.Run("tcp", func(t *testing.T) {
		job := Job{
			Protocol: "tcp",
			ScanTech: "sA",
			Target:   "127.0.0.1",
			Port:     "80-81",
			Extra:    "for test",
			Options: &Options{
				NoPing:    true,
				Arguments: "--ttl 128 --badsum",
			},
		}
		const cmd = "-sA -p 80-81 -Pn --ttl 128 --badsum -n 127.0.0.1"
		require.Equal(t, cmd, job.String())
	})

	t.Run("udp", func(t *testing.T) {
		job := Job{
			Protocol: "udp",
			Target:   "127.0.0.1",
			Port:     "80-81",
			Extra:    "for test",
			Options: &Options{
				NoPing:    true,
				Arguments: "--ttl 128 --badsum",
			},
		}
		const cmd = "-sU -p 80-81 -Pn --ttl 128 --badsum -n 127.0.0.1"
		require.Equal(t, cmd, job.String())
	})
}

func TestJob_selectScanTech(t *testing.T) {

}

func TestJob_String(t *testing.T) {

}

func TestIsDomainName(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, domain := range [...]string{
			"test.com",
			"Test-sub.com",
			"test-sub2.com",
		} {
			require.True(t, isDomainName(domain))
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for _, domain := range [...]string{
			"",
			string([]byte{255, 254, 12, 35}),
			"test-",
			"Test.-",
			"test..",
			strings.Repeat("a", 64) + ".com",
		} {
			require.False(t, isDomainName(domain))
		}
	})
}

func TestJob(t *testing.T) {

}
