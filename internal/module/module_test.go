package module

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMethod_String(t *testing.T) {
	t.Run("full", func(t *testing.T) {
		const expected = `
--------------------------------
method: Scan
--------------------------------
parameter:
  host string
  port uint16
--------------------------------
return value:
  open bool
  err  error
--------------------------------
`
		method := &Method{
			Name: "Scan",
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
--------------------------------
method: Kill
--------------------------------
return value:
  ok  bool
  err error
--------------------------------
`
		method := &Method{
			Name: "Kill",
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
--------------------------------
method: Kill
--------------------------------
parameter:
  pid   uint32
  force bool
--------------------------------
`
		method := &Method{
			Name: "Kill",
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
--------------------------------
method: Stop
--------------------------------
`
		method := &Method{
			Name: "Stop",
		}

		str := method.String()
		fmt.Println(str)
		require.Equal(t, expected[1:len(expected)-1], str)
	})
}
