package ssh

import (
	"golang.org/x/crypto/ssh"
)

// Brute is
func Brute(address string, usernames, passwords []string) (string, string, bool) {
	for _, username := range usernames {
		for _, password := range passwords {
			if Login(address, username, password) {
				return username, password, true
			}
		}
	}
	return "", "", false
}

// Login is
func Login(address string, username, password string) bool {
	clientCfg := ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", address, &clientCfg)
	if err != nil {
		return false
	}
	defer func() { _ = client.Close() }()
	return true
}
