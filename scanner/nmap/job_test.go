package nmap

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/toml"
	"project/internal/testsuite"
)

func TestJob_ToArgs(t *testing.T) {
	t.Run("tcp", func(t *testing.T) {
		job := Job{
			Protocol:   "tcp",
			ScanTech:   "sA",
			Target:     "127.0.0.1",
			Port:       "80-81",
			Extra:      "for test",
			outputPath: "output/1.xml",
			Options: &Options{
				NoPing:    true,
				Arguments: "--ttl 128 --badsum",
			},
		}
		const cmd = "-sA -p 80-81 -Pn --ttl 128 --badsum -oX output/1.xml -n 127.0.0.1"
		require.Equal(t, cmd, job.String())
	})

	t.Run("udp", func(t *testing.T) {
		job := Job{
			Protocol:   "udp",
			Target:     "127.0.0.1",
			Port:       "80-81",
			Extra:      "for test",
			outputPath: "output/1.xml",
			Options: &Options{
				NoPing:    true,
				Arguments: "--ttl 128 --badsum",
			},
		}
		const cmd = "-sU -p 80-81 -Pn --ttl 128 --badsum -oX output/1.xml -n 127.0.0.1"
		require.Equal(t, cmd, job.String())
	})
}

func TestJob_selectScanTech(t *testing.T) {
	t.Run("not set ScanTech", func(t *testing.T) {
		job := Job{
			Protocol:   "tcp",
			Target:     "127.0.0.1",
			Port:       "80-81",
			outputPath: "output/1.xml",
		}
		const cmd = "-sS -p 80-81 -oX output/1.xml -n 127.0.0.1"
		require.Equal(t, cmd, job.String())
	})

	t.Run("invalid TCP scan technique", func(t *testing.T) {
		job := Job{
			Protocol: "tcp",
			ScanTech: "sU",
		}
		args, err := job.ToArgs()
		require.EqualError(t, err, "invalid TCP scan technique: sU")
		require.Nil(t, args)
	})

	t.Run("invalid UDP scan technique", func(t *testing.T) {
		job := Job{
			Protocol: "udp",
			ScanTech: "sS",
		}
		args, err := job.ToArgs()
		require.EqualError(t, err, "UDP scan not support technique field except sU")
		require.Nil(t, args)
	})

	t.Run("empty protocol", func(t *testing.T) {
		job := Job{}
		args, err := job.ToArgs()
		require.EqualError(t, err, "protocol is empty")
		require.Nil(t, args)
	})

	t.Run("invalid protocol", func(t *testing.T) {
		job := Job{Protocol: "foo"}
		args, err := job.ToArgs()
		require.EqualError(t, err, "invalid protocol: \"foo\"")
		require.Nil(t, args)
	})
}

func TestJob_String(t *testing.T) {
	job := Job{Protocol: "foo"}
	require.Equal(t, "invalid protocol: \"foo\"", job.String())
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
	data, err := ioutil.ReadFile("testdata/job.toml")
	require.NoError(t, err)

	// check unnecessary field
	job := Job{}
	err = toml.Unmarshal(data, &job)
	require.NoError(t, err)

	// check zero value
	testsuite.ContainZeroValue(t, job)

	for _, testdata := range [...]*struct {
		expected interface{}
		actual   interface{}
	}{
		{expected: "tcp", actual: job.Protocol},
		{expected: "sS", actual: job.ScanTech},
		{expected: "127.0.0.1", actual: job.Target},
		{expected: "80-81", actual: job.Port},
		{expected: "for test", actual: job.Extra},
		{expected: true, actual: job.Options.NoPing},
	} {
		require.Equal(t, testdata.expected, testdata.actual)
	}
}
