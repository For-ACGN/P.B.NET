// +build windows

package kiwimon

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/logger"
	"project/internal/module/windows/kiwi"
	"project/internal/testsuite"
)

func TestMonitor(t *testing.T) {
	monitor, err := New(logger.Test, func(local, remote string, pid int64, cred *kiwi.Credential) {
		fmt.Println(local, remote, pid, cred)
	}, nil)
	require.NoError(t, err)

	time.Sleep(3 * time.Second)

	err = monitor.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, monitor)
}
