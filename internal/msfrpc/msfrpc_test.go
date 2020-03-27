package msfrpc

import (
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

const (
	testHost     = "127.0.0.1"
	testPort     = 55553
	testUsername = "msf"
	testPassword = "msf"
)

func TestNewMSFRPC(t *testing.T) {
	msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, nil)
	require.NoError(t, err)

	testsuite.IsDestroyed(t, msfrpc)
}

func TestMSFRPC_Login(t *testing.T) {
	msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, nil)
	require.NoError(t, err)

	t.Run("login success", func(t *testing.T) {
		err = msfrpc.Login()
		require.NoError(t, err)
	})

	t.Run("login failed", func(t *testing.T) {
		msfrpc.password = "foo"
		err = msfrpc.Login()
		require.Error(t, err)

		msfrpc.password = testUsername
	})

	testsuite.IsDestroyed(t, msfrpc)
}
