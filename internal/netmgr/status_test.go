package netmgr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/convert"
	"project/internal/logger"
)

func TestListenerStatus_String(t *testing.T) {
	ls := ListenerStatus{
		Network:  "tcp",
		Address:  "127.0.0.1:1234",
		EstConns: 123,
		MaxConns: 10000,
	}
	listened, err := time.Parse(logger.TimeLayout, "2018-11-27 00:00:00 +08:00")
	require.NoError(t, err)
	lastAccept, err := time.Parse(logger.TimeLayout, "2018-11-27 00:01:00 +08:00")
	require.NoError(t, err)
	ls.Listened = listened
	ls.LastAccept = lastAccept

	const expected = `
------------------------listener status-------------------------
address:     tcp 127.0.0.1:1234
connections: 123/10000 (est/max)
listened:    2018-11-27 00:00:00 +08:00
last accept: 2018-11-27 00:01:00 +08:00
----------------------------------------------------------------`
	require.Equal(t, expected[1:], ls.String())
}

func TestConnStatus_String(t *testing.T) {
	cs := ConnStatus{
		LocalNetwork:   "tcp",
		LocalAddress:   "127.0.0.1:1234",
		RemoteNetwork:  "tcp4",
		RemoteAddress:  "127.0.0.1:5678",
		ReadLimitRate:  32 * convert.MiB,
		WriteLimitRate: 16 * convert.MiB,
		Read:           1127,
		Written:        123,
	}
	established, err := time.Parse(logger.TimeLayout, "2018-11-27 00:00:00 +08:00")
	require.NoError(t, err)
	lastRead, err := time.Parse(logger.TimeLayout, "2018-11-27 00:02:00 +08:00")
	require.NoError(t, err)
	lastWrite, err := time.Parse(logger.TimeLayout, "2018-11-27 00:01:00 +08:00")
	require.NoError(t, err)
	cs.Established = established
	cs.LastRead = lastRead
	cs.LastWrite = lastWrite

	t.Run("limit", func(t *testing.T) {
		const expected = `
-----------------------connection status------------------------
local addr:  tcp 127.0.0.1:1234
remote addr: tcp4 127.0.0.1:5678
rate:        16 MiB/32 MiB (send/recv)
traffic:     123 Byte/1.1 KiB (sent/recv)
established: 2018-11-27 00:00:00 +08:00
last send:   2018-11-27 00:01:00 +08:00
last recv:   2018-11-27 00:02:00 +08:00
----------------------------------------------------------------`
		require.Equal(t, expected[1:], cs.String())
	})

	t.Run("no read limit", func(t *testing.T) {
		cs.ReadLimitRate = 0
		cs.WriteLimitRate = 16 * convert.MiB

		const expected = `
-----------------------connection status------------------------
local addr:  tcp 127.0.0.1:1234
remote addr: tcp4 127.0.0.1:5678
rate:        16 MiB/[no limit] (send/recv)
traffic:     123 Byte/1.1 KiB (sent/recv)
established: 2018-11-27 00:00:00 +08:00
last send:   2018-11-27 00:01:00 +08:00
last recv:   2018-11-27 00:02:00 +08:00
----------------------------------------------------------------`
		require.Equal(t, expected[1:], cs.String())
	})

	t.Run("no write limit", func(t *testing.T) {
		cs.ReadLimitRate = 32 * convert.MiB
		cs.WriteLimitRate = 0

		const expected = `
-----------------------connection status------------------------
local addr:  tcp 127.0.0.1:1234
remote addr: tcp4 127.0.0.1:5678
rate:        [no limit]/32 MiB (send/recv)
traffic:     123 Byte/1.1 KiB (sent/recv)
established: 2018-11-27 00:00:00 +08:00
last send:   2018-11-27 00:01:00 +08:00
last recv:   2018-11-27 00:02:00 +08:00
----------------------------------------------------------------`
		require.Equal(t, expected[1:], cs.String())
	})

	t.Run("no limit", func(t *testing.T) {
		cs.ReadLimitRate = 0
		cs.WriteLimitRate = 0

		const expected = `
-----------------------connection status------------------------
local addr:  tcp 127.0.0.1:1234
remote addr: tcp4 127.0.0.1:5678
rate:        [no limit]/[no limit] (send/recv)
traffic:     123 Byte/1.1 KiB (sent/recv)
established: 2018-11-27 00:00:00 +08:00
last send:   2018-11-27 00:01:00 +08:00
last recv:   2018-11-27 00:02:00 +08:00
----------------------------------------------------------------`
		require.Equal(t, expected[1:], cs.String())
	})
}
