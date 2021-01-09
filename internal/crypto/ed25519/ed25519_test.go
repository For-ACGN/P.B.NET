package ed25519

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/crypto/rand"
	"project/internal/patch/monkey"
)

func TestEd25519(t *testing.T) {
	pri, err := GenerateKey()
	require.NoError(t, err)

	message := []byte("test message")
	signature := Sign(pri, message)
	require.Len(t, signature, SignatureSize)

	valid := Verify(GetPublicKey(pri), message, signature)
	require.True(t, valid)
}

func TestGenerateKey(t *testing.T) {
	patch := func([]byte) (int, error) {
		return 0, monkey.Error
	}
	pg := monkey.Patch(rand.Read, patch)
	defer pg.Unpatch()

	_, err := GenerateKey()
	monkey.IsExistMonkeyError(t, err)
}

func TestImportPrivateKey(t *testing.T) {
	pri, err := ImportPrivateKey(bytes.Repeat([]byte{0, 1}, 32))
	require.NoError(t, err)
	require.NotNil(t, pri)

	pri, err = ImportPrivateKey(bytes.Repeat([]byte{0, 1}, 161))
	require.Equal(t, ErrInvalidPrivateKeySize, err)
	require.Nil(t, pri)
}

func TestImportPublicKey(t *testing.T) {
	pub, err := ImportPublicKey(bytes.Repeat([]byte{0, 1}, 16))
	require.NoError(t, err)
	require.NotNil(t, pub)

	pub, err = ImportPublicKey(bytes.Repeat([]byte{0, 1}, 161))
	require.Equal(t, ErrInvalidPublicKeySize, err)
	require.Nil(t, pub)
}

func BenchmarkSign(b *testing.B) {
	pri, err := GenerateKey()
	require.NoError(b, err)
	msg := bytes.Repeat([]byte{0}, 256)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Sign(pri, msg)
	}

	b.StopTimer()
}

func BenchmarkVerify(b *testing.B) {
	pri, err := GenerateKey()
	require.NoError(b, err)
	msg := bytes.Repeat([]byte{0}, 256)
	signature := Sign(pri, msg)
	pub := GetPublicKey(pri)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Verify(pub, msg, signature)
	}

	b.StopTimer()
}
