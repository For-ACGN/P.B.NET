package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"flag"
	"fmt"
	"log"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"

	s "project/module/auxiliary/server/ssh"
)

func main() {
	var (
		network string
		address string
	)
	flag.StringVar(&network, "net", "tcp", "ssh tcp listener network")
	flag.StringVar(&address, "addr", "127.0.0.1:1022", "ssh tcp listener address")
	flag.Parse()

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
	if err != nil {
		log.Fatalln(err)
	}
	signer, err := ssh.NewSignerFromSigner(pri)
	if err != nil {
		log.Fatalln(err)
	}
	cfg.AddHostKey(signer)

	server, err := s.New(network, address, &cfg)
	if err != nil {
		log.Fatalln(err)
	}
	server.Serve()
}
