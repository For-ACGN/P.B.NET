// +build windows

package rdpthief

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestClient(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	client, err := NewClient("test", "pass")
	require.NoError(t, err)

	testCreateCredential(t)

	// wait sendCred
	time.Sleep(time.Second)

	err = client.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, client)
}
