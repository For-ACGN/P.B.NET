package mysql

import (
	"fmt"
	"testing"
)

func TestLogin(t *testing.T) {
	fmt.Println(connect("127.0.0.1:3306", "test", "test"))
}
