package mysql

import (
	"fmt"
	"testing"
)

func TestLogin(t *testing.T) {
	fmt.Println(Login("127.0.0.1:3306", "pbnet", "pbnet"))
}

func TestLogin2(t *testing.T) {
	fmt.Println(Login("127.0.0.1:3406", "test", "test"))
}

func TestConnect(t *testing.T) {
	fmt.Println(connect("127.0.0.1:3306", "pbnet", "pbneasdt"))
}
