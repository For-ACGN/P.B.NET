package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"net"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestNew(t *testing.T) {
	cfg := ssh.ServerConfig{
		NoClientAuth: false,
		MaxAuthTries: 100,
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			user := conn.User()
			pwd := string(password)
			fmt.Println("address:", conn.RemoteAddr())
			fmt.Println("username:", user)
			fmt.Println("password:", pwd)
			fmt.Println()

			if user != "root" || pwd != "123456" {
				return nil, errors.New("invalid username/password")
			}
			return nil, nil
		},
	}
	_, pri, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	signer, err := ssh.NewSignerFromSigner(pri)
	require.NoError(t, err)

	cfg.AddHostKey(signer)

	server, err := New("tcp", "127.0.0.1:1022", &cfg)
	require.NoError(t, err)
	go server.Serve()

	clientCfg := ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{ssh.Password("123456")},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	client, err := ssh.Dial("tcp", "127.0.0.1:1022", &clientCfg)
	require.NoError(t, err)
	client.Close()
}
