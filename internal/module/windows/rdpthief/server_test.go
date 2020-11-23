package rdpthief

import (
	"os/exec"
	"testing"
	"time"

	"github.com/Microsoft/go-winio"
	"github.com/stretchr/testify/require"

	"project/internal/logger"
	"project/internal/testsuite"
)

// create "leaks" goroutine first in github.com/Microsoft/go-winio
func init() {
	_, _ = winio.MakeOpenFile(0)
}

var (
	testPipeName = "test_rdpthief"
	testPassword = "test"
	testConfig   = &Config{
		PipeName: testPipeName,
		Password: testPassword,
	}
)

// simulate
func testInjector(uint32, []byte) error {
	return nil
}

func TestServer(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	var received bool
	cb := func(cred *Credential) {
		require.Equal(t, testCredHostname, cred.Hostname)
		require.Equal(t, testCredUsername, cred.Username)
		require.Equal(t, testCredPassword, cred.Password)
		received = true
	}
	server, err := NewServer(logger.Test, testInjector, cb, testConfig)
	require.NoError(t, err)

	time.Sleep(time.Second)

	client, err := NewClient(testPipeName, testPassword)
	require.NoError(t, err)

	cmd := exec.Command("mstsc.exe")
	err = cmd.Start()
	require.NoError(t, err)

	time.Sleep(time.Second)

	// simulate steal credential
	testCreateCredential(t)
	time.Sleep(time.Second)

	err = client.Close()
	require.NoError(t, err)

	err = server.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, client)
	testsuite.IsDestroyed(t, server)

	require.True(t, received)

	err = cmd.Process.Kill()
	require.NoError(t, err)
	// exit status 1
	err = cmd.Wait()
	require.Error(t, err)
}
