package ecdsa

import (
	"bytes"
	"crypto/elliptic"
	"crypto/x509"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/crypto/rsa"
)

func Test_ECDSA(t *testing.T) {
	privatekey, err := Generate_Key(elliptic.P256())
	require.Nil(t, err, err)
	file, err := ioutil.ReadFile("ecdsa.key")
	require.Nil(t, err, err)
	_, err = Import_PrivateKey_PEM(file)
	require.Nil(t, err, err)
	_, err = Import_PrivateKey_PEM(nil)
	require.Equal(t, err, ERR_INVALID_PEM_BLOCK, err)
	privatekey_bytes, err := Export_PrivateKey(privatekey)
	require.Nil(t, err, err)
	_, err = Import_PrivateKey(privatekey_bytes)
	require.Nil(t, err, err)
	publickey := &privatekey.PublicKey
	publickey_bytes := Export_PublicKey(publickey)
	_, err = Import_PublicKey(publickey_bytes)
	require.Nil(t, err, err)
	// invalid publickey
	_, err = Import_PublicKey(nil)
	require.NotNil(t, err)
	// rsa publickey
	rsa_pri, _ := rsa.Generate_Key(1024)
	rsa_pri_b, _ := x509.MarshalPKIXPublicKey(&rsa_pri.PublicKey)
	_, err = Import_PublicKey(rsa_pri_b)
	require.Equal(t, ERR_NOT_PUBLIC_KEY, err, err)
	signature, err := Sign(privatekey, file)
	require.Nil(t, err, err)
	require.True(t, Verify(publickey, file, signature), "invalid data")
	require.False(t, Verify(publickey, file, nil), "error verify")
	// error sign
	privatekey.PublicKey.Curve.Params().N.SetBytes(nil)
	_, err = Sign(privatekey, file)
	require.NotNil(t, err)
	// error verify
	msg := "error verify"
	require.False(t, Verify(publickey, file, []byte{0}), msg)
	require.False(t, Verify(publickey, file, []byte{0, 3, 22, 22}), msg)
	require.False(t, Verify(publickey, file, []byte{0, 2, 22, 22}), msg)
	require.False(t, Verify(publickey, file, []byte{0, 2, 22, 22, 0, 1}), msg)
}

func Benchmark_Sign(b *testing.B) {
	privatekey, err := Generate_Key(elliptic.P256())
	require.Nil(b, err, err)
	data := bytes.Repeat([]byte{0}, 4096)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Sign(privatekey, data)
	}
}

func Benchmark_Verify(b *testing.B) {
	privatekey, err := Generate_Key(elliptic.P256())
	require.Nil(b, err, err)
	data := bytes.Repeat([]byte{0}, 4096)
	signature, err := Sign(privatekey, data)
	require.Nil(b, err, err)
	publickey := &privatekey.PublicKey
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Verify(publickey, data, signature)
		// require.True(b, Verify(publickey, data, signature), "verify failed")
	}
	b.StopTimer()
}

func Benchmark_Sign_Verify(b *testing.B) {
	privatekey, err := Generate_Key(elliptic.P256())
	require.Nil(b, err, err)
	data := bytes.Repeat([]byte{0}, 4096)
	publickey := &privatekey.PublicKey
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		signature, _ := Sign(privatekey, data)
		Verify(publickey, data, signature)
		// signature, err := Sign(privatekey, data)
		// require.Nil(b, err, err)
		// require.True(b, Verify(publickey, data, signature), "verify failed")
	}
	b.StopTimer()
}
