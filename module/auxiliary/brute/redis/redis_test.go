package redis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnonymousLogin(t *testing.T) {
	ok, err := anonymousLogin("192.168.255.186:6379")
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = anonymousLogin("192.168.255.186:6380")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestLogin(t *testing.T) {
	ok, err := login("192.168.255.186:6379", "", "test1")
	require.Error(t, err)
	require.False(t, ok)

	ok, err = login("192.168.255.186:6379", "", "test")
	require.NoError(t, err)
	require.True(t, ok)
}
