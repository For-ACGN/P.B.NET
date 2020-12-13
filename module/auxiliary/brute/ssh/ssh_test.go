package ssh

import (
	"fmt"
	"testing"
)

func TestBrute(t *testing.T) {
	address := "127.0.0.1:1022"
	usernames := []string{"root"}
	passwords := []string{"root", "admin", "123456", "12345678"}
	username, password, ok := Brute(address, usernames, passwords)
	if ok {
		fmt.Println(username, password)
	}
}
