package logger

import (
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

const (
	testPrefixF  = "test format %s %s"
	testPrefix   = "test print"
	testPrefixLn = "test println"
	testSrc      = "test src"
	testLog1     = "test"
	testLog2     = "log"
)

func TestParse(t *testing.T) {
	for _, testdata := range [...]*struct {
		name  string
		level Level
	}{
		{"all", All},
		{"trace", Trace},
		{"debug", Debug},
		{"info", Info},
		{"crucial", Crucial},
		{"warning", Warning},
		{"error", Error},
		{"exploit", Exploit},
		{"fatal", Fatal},
		{"off", Off},
	} {
		t.Run(testdata.name, func(t *testing.T) {
			l, err := Parse(testdata.name)
			require.NoError(t, err)
			require.Equal(t, l, testdata.level)
		})
	}

	t.Run("invalid level", func(t *testing.T) {
		lv, err := Parse("invalid level")
		require.Error(t, err)
		require.Equal(t, lv, Level(0))
	})
}

func TestPrefix(t *testing.T) {
	for lv := Level(0); lv < Off; lv++ {
		fmt.Println(Prefix(time.Now(), lv, testSrc).String())
	}
	// unknown level
	fmt.Println(Prefix(time.Now(), Level(153), testSrc).String())
}

func TestLogger(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		Common.Printf(Info, testSrc, testPrefixF, testLog1, testLog2)
		Common.Print(Info, testSrc, testPrefix, testLog1, testLog2)
		Common.Println(Info, testSrc, testPrefixLn, testLog1, testLog2)

		// will not display
		Common.Printf(Debug, testSrc, testPrefixF, testLog1, testLog2)
		Common.Print(Debug, testSrc, testPrefix, testLog1, testLog2)
		Common.Println(Debug, testSrc, testPrefixLn, testLog1, testLog2)
	})

	t.Run("test", func(t *testing.T) {
		Test.Printf(Debug, testSrc, testPrefixF, testLog1, testLog2)
		Test.Print(Debug, testSrc, testPrefix, testLog1, testLog2)
		Test.Println(Debug, testSrc, testPrefixLn, testLog1, testLog2)
	})

	t.Run("discard", func(t *testing.T) {
		Discard.Printf(Debug, testSrc, testPrefixF, testLog1, testLog2)
		Discard.Print(Debug, testSrc, testPrefix, testLog1, testLog2)
		Discard.Println(Debug, testSrc, testPrefixLn, testLog1, testLog2)
	})
}

func TestMultiLogger(t *testing.T) {
	logger := NewMultiLogger(Debug, os.Stdout)

	t.Run("common", func(t *testing.T) {
		logger.Printf(Debug, testSrc, testPrefixF, testLog1, testLog2)
		logger.Print(Debug, testSrc, testPrefix, testLog1, testLog2)
		logger.Println(Debug, testSrc, testPrefixLn, testLog1, testLog2)
	})

	t.Run("low level", func(t *testing.T) {
		err := logger.SetLevel(Info)
		require.NoError(t, err)

		logger.Printf(Debug, testSrc, testPrefixF, testLog1, testLog2)
		logger.Print(Debug, testSrc, testPrefix, testLog1, testLog2)
		logger.Println(Debug, testSrc, testPrefixLn, testLog1, testLog2)
	})

	t.Run("invalid level", func(t *testing.T) {
		err := logger.SetLevel(Level(123))
		require.EqualError(t, err, "invalid logger level: 123")
	})

	err := logger.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, logger)
}

func TestNewWriterWithPrefix(t *testing.T) {
	w := NewWriterWithPrefix(os.Stdout, "prefix")
	_, err := w.Write([]byte("test\n"))
	require.NoError(t, err)
}

func TestWrapLogger(t *testing.T) {
	w := WrapLogger(Debug, "test wrap", Test)
	_, err := w.Write([]byte("test data"))
	require.NoError(t, err)
}

func TestWrap(t *testing.T) {
	l := Wrap(Debug, "test wrap", Test)
	l.Printf("Printf")
	l.Print("Print")
	l.Println("Println")
}

func TestHijackLogWriter(t *testing.T) {
	HijackLogWriter(Error, "test", Test)
	log.Printf("Printf")
	log.Print("Print")
	log.Println("Println")
}

func TestSetErrorLogger(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		const name = "testdata/test.err"

		file, err := SetErrorLogger(name)
		require.NoError(t, err)

		log.Println("test log")

		err = file.Close()
		require.NoError(t, err)
		err = os.Remove(name)
		require.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		file, err := SetErrorLogger("testdata/<</file")
		require.Error(t, err)
		require.Nil(t, file)
	})
}

func TestConn(t *testing.T) {
	t.Run("local", func(t *testing.T) {
		listener, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)

		conn, err := net.Dial("tcp", listener.Addr().String())
		require.NoError(t, err)
		defer func() { _ = conn.Close() }()
		fmt.Println(Conn(conn))

		err = listener.Close()
		require.NoError(t, err)
	})

	t.Run("mock", func(t *testing.T) {
		conn := testsuite.NewMockConn()
		defer func() { _ = conn.Close() }()
		fmt.Println(Conn(conn))
	})
}
