package curve25519

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestX25519Base(t *testing.T) {
	x := make([]byte, ScalarSize)
	x[0] = 1
	for i := 0; i < 200; i++ {
		var err error
		x, err = X25519Base(x)
		if err != nil {
			t.Fatal(err)
		}
	}
	result := hex.EncodeToString(x)
	const expectedHex = "89161fde887b2b53de549af483940106ecc114d6982daa98256de23bdf77661a"
	require.Equal(t, expectedHex, result)
}

func TestKeyExchange(t *testing.T) {
	// client side
	clientPri := make([]byte, ScalarSize)
	clientPri[0] = 199
	clientPub, err := X25519Base(clientPri)
	require.NoError(t, err)

	// server side
	serverPri := make([]byte, ScalarSize)
	serverPri[0] = 2
	serverPub, err := X25519Base(serverPri)
	require.NoError(t, err)

	// start exchange
	clientKey, err := X25519(clientPri, serverPub)
	require.NoError(t, err)
	serverKey, err := X25519(serverPri, clientPub)
	require.NoError(t, err)

	// check result key
	require.NotEqual(t, bytes.Repeat([]byte{0}, 32), clientKey)
	require.Equal(t, clientKey, serverKey)
	t.Log(clientKey)

	// invalid in data size
	clientPub, err = X25519Base(nil)
	require.Error(t, err)
	require.Nil(t, clientPub)
}
