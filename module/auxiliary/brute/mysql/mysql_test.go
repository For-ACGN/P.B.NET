package mysql

import (
	"fmt"
	"testing"
)

func TestConnect(t *testing.T) {
	fmt.Println(connect("127.0.0.1:3306", "pbnet", "pbnet"))
	fmt.Println(connect("127.0.0.1:3306", "pbnet", "1234"))
	fmt.Println(connect("127.0.0.1:3306", "root", ""))

	fmt.Println(connect("127.0.0.1:3406", "test", "test"))
	fmt.Println(connect("127.0.0.1:3406", "test", "1234"))
	fmt.Println(connect("127.0.0.1:3406", "root", ""))
}
