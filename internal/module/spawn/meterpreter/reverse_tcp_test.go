package meterpreter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testNetwork = "tcp"
	testAddress = "127.0.0.1:8990"
	// testRC4Key  = "acg"
)

func TestReverseTCP(t *testing.T) {
	err := ReverseTCP(testNetwork, testAddress, "")
	require.Error(t, err)
}

// func TestReverseTCPRC4(t *testing.T) {
// 	err := ReverseTCPRC4(testNetwork, testAddress, "", []byte(testRC4Key))
// 	require.NoError(t, err)
// }
