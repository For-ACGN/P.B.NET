package shell

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/axgle/mahonia"
	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestTrimPrefixSpace(t *testing.T) {
	for _, testdata := range [...]*struct {
		input    string
		expected string
	}{
		{input: "a  ", expected: "a  "},
		{input: "  a", expected: "a"},
		{input: "   ", expected: ""},
		{input: "  a ", expected: "a "},
	} {
		require.Equal(t, testdata.expected, trimPrefixSpace(testdata.input))
	}
}

func TestTerminal(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	terminal, err := NewTerminal(true)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		reader := mahonia.NewDecoder("GBK").NewReader(terminal)
		buf := make([]byte, 512)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				return
			}
			data := make([]byte, n)
			copy(data, buf[:n])
			_, err = os.Stdout.Write(data)
			require.NoError(t, err)
		}
	}()

	// wait print welcome information
	time.Sleep(1 * time.Second)

	command := []string{
		// about empty command
		"\n",
		"  \n",

		// about change directory
		"  cd\n",
		"  cd ..\n",
		"cd shell\n",
		`cd "is not exist"` + "\n",

		// about environment variable
		"  set\n",
		"  set   \n",
		"set  p\n",
		"set  pt\n",
		"set  =  \n",
		"set  test=value 1\n",
		"set test\n",
		"set test=value 2\n",
		"set test\n",
		"set test=\n",
		"set test\n",
		"set\n",

		// about dir
		"dir\n",
		"ls\n",
		"cd ..\n",
		"dir\n",
		"ls\n",
		"dir shell\n",
		"ls shell\n",
		"cd ../..\n",
		"dir\n",
		"ls\n",
		`dir "is not exist"` + "\n",

		// about execute
		"ping",
		"interrupt",
		"\"does not exist\"\n",

		// last
		"dir C:\\windows\n",
		"dir /\n",
		"cd C:\\windows\n",
		"cd /\n",

		"exit\n",
	}

	for _, cmd := range command {
		switch cmd {
		case "interrupt":
			err = terminal.Interrupt()
			require.NoError(t, err)
			time.Sleep(1 * time.Second)
		case "ping":
			_, err = terminal.Write([]byte("ping 8.8.8.8 -t\n"))
			require.NoError(t, err)
			time.Sleep(4 * time.Second)
		default:
			_, err = terminal.Write([]byte(cmd))
			require.NoError(t, err)
			time.Sleep(250 * time.Millisecond)
		}
	}

	err = terminal.Close()
	require.NoError(t, err)

	wg.Wait()

	// for test output
	fmt.Println()
}
