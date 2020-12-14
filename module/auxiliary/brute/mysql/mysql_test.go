package mysql

import (
	"fmt"
	"testing"
)


func TestLogin(t *testing.T) {
	fmt.Println(Login("127.0.0.1:3306", "root", ""))
}


func TestConnect(t *testing.T) {
	fmt.Println(connect("127.0.0.1:3406", "test", "test"))
}
