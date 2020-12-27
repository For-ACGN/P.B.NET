package module

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMethod_String(t *testing.T) {
	t.Run("full", func(t *testing.T) {
		const expected = `
----------------------------------------------------------------
Method: Scan
----------------------------------------------------------------
Description:
  Scan is used to scan a host with port, it will return the port
  status about this host.
----------------------------------------------------------------
Parameter:
  host string
  port uint16
----------------------------------------------------------------
Return Value:
  open bool
  err  error
----------------------------------------------------------------
`
		method := &Method{
			Name: "Scan",
			Desc: "Scan is used to scan a host with port," +
				" it will return the port status about this host.",
			Args: []*Value{
				{"host", "string"},
				{"port", "uint16"},
			},
			Rets: []*Value{
				{"open", "bool"},
				{"err", "error"},
			},
		}

		str := method.String()
		fmt.Println(str)
		require.Equal(t, expected[1:len(expected)-1], str)
	})

	t.Run("no parameter", func(t *testing.T) {
		const expected = `
----------------------------------------------------------------
Method: Kill
----------------------------------------------------------------
Description:
  Kill is used to kill current scan task.
----------------------------------------------------------------
Return Value:
  ok  bool
  err error
----------------------------------------------------------------
`
		method := &Method{
			Name: "Kill",
			Desc: "Kill is used to kill current scan task.",
			Rets: []*Value{
				{"ok", "bool"},
				{"err", "error"},
			},
		}

		str := method.String()
		fmt.Println(str)
		require.Equal(t, expected[1:len(expected)-1], str)
	})

	t.Run("no return value", func(t *testing.T) {
		const expected = `
----------------------------------------------------------------
Method: Kill
----------------------------------------------------------------
Description:
  Kill is used to kill current scan task.
----------------------------------------------------------------
Parameter:
  pid   uint32
  force bool
----------------------------------------------------------------
`
		method := &Method{
			Name: "Kill",
			Desc: "Kill is used to kill current scan task.",
			Args: []*Value{
				{"pid", "uint32"},
				{"force", "bool"},
			},
		}

		str := method.String()
		fmt.Println(str)
		require.Equal(t, expected[1:len(expected)-1], str)
	})

	t.Run("only name", func(t *testing.T) {
		const expected = `
----------------------------------------------------------------
Method: Stop
----------------------------------------------------------------
Description:
  Stop is used to stop all scan tasks.
----------------------------------------------------------------
`
		method := &Method{
			Name: "Stop",
			Desc: "Stop is used to stop all scan tasks.",
		}

		str := method.String()
		fmt.Println(str)
		require.Equal(t, expected[1:len(expected)-1], str)
	})
}

func TestMethod_printDescription(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		const expected = `
----------------------------------------------------------------
Method: Stop
----------------------------------------------------------------
Description:
----------------------------------------------------------------
`
		method := &Method{
			Name: "Stop",
		}

		str := method.String()
		fmt.Println(str)
		require.Equal(t, expected[1:len(expected)-1], str)
	})

	t.Run("equal one line", func(t *testing.T) {
		const expected = `
----------------------------------------------------------------
Method: Stop
----------------------------------------------------------------
Description:
  a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a 
----------------------------------------------------------------
`
		method := &Method{
			Name: "Stop",
			Desc: strings.Repeat("a ", (maxLineSize-2)/2),
		}

		str := method.String()
		fmt.Println(str)
		require.Equal(t, expected[1:len(expected)-1], str)
	})

	t.Run("long", func(t *testing.T) {
		const expected = `
----------------------------------------------------------------
Method: Stop
----------------------------------------------------------------
Description:
  a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a 
  a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a 
----------------------------------------------------------------
`
		method := &Method{
			Name: "Stop",
			Desc: strings.Repeat("a ", maxLineSize-2),
		}

		str := method.String()
		fmt.Println(str)
		require.Equal(t, expected[1:len(expected)-1], str)
	})
}
