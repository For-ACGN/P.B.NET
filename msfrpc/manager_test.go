package msfrpc

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/logger"
	"project/internal/testsuite"
)

func TestIOReader_Read(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	r, w := io.Pipe()
	var (
		read   bool
		closed bool
	)
	onRead := func() { read = true }
	onClose := func() { closed = true }
	reader := newIOReader(logger.Test, r, onRead, onClose)
	reader.ReadLoop()

	testdata := testsuite.Bytes()
	_, err := w.Write(testdata)
	require.NoError(t, err)

	// wait read
	time.Sleep(100 * time.Millisecond)

	t.Run("common", func(t *testing.T) {
		output := reader.Read(0)
		require.Equal(t, testdata, output)
	})

	t.Run("start < 0", func(t *testing.T) {
		output := reader.Read(-1)
		require.Equal(t, testdata, output)
	})

	t.Run("start != 0", func(t *testing.T) {
		output := reader.Read(10)
		require.Equal(t, testdata[10:], output)
	})

	t.Run("start > len", func(t *testing.T) {
		output := reader.Read(257)
		require.Nil(t, output)
	})

	err = reader.Close()
	require.NoError(t, err)

	require.True(t, read)
	require.True(t, closed)

	testsuite.IsDestroyed(t, reader)
}

func TestIOReader_Clean(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	r, w := io.Pipe()
	var (
		read   bool
		closed bool
	)
	onRead := func() { read = true }
	onClose := func() { closed = true }
	reader := newIOReader(logger.Test, r, onRead, onClose)
	reader.ReadLoop()

	testdata := testsuite.Bytes()
	_, err := w.Write(testdata)
	require.NoError(t, err)

	reader.Clean()

	output := reader.Read(257)
	require.Nil(t, output)

	err = reader.Close()
	require.NoError(t, err)

	require.True(t, read)
	require.True(t, closed)

	testsuite.IsDestroyed(t, reader)
}

func TestIOReader_Panic(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	conn := testsuite.NewMockConnWithReadPanic()
	var (
		read   bool
		closed bool
	)
	onRead := func() { read = true }
	onClose := func() { closed = true }
	reader := newIOReader(logger.Test, conn, onRead, onClose)
	reader.ReadLoop()

	time.Sleep(time.Second)

	err := reader.Close()
	require.NoError(t, err)

	require.False(t, read)
	require.True(t, closed)

	testsuite.IsDestroyed(t, reader)
}

func TestIOReader_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	testdata := testsuite.Bytes()

	t.Run("without close", func(t *testing.T) {
		t.Run("part", func(t *testing.T) {
			r, w := io.Pipe()
			var (
				bRead  bool
				closed bool
			)
			onRead := func() { bRead = true }
			onClose := func() { closed = true }
			reader := newIOReader(logger.Test, r, onRead, onClose)
			reader.ReadLoop()

			write := func() {
				_, err := w.Write(testdata)
				require.NoError(t, err)
			}
			read := func() {
				for i := 0; i < 100; i++ {
					output := reader.Read(i)
					if len(output) == 0 {
						return
					}
					require.Equal(t, testdata[i:], output)
				}
			}
			clean := func() {
				reader.Clean()
			}
			cleanup := func() {
				// maybe not read finish
				time.Sleep(25 * time.Millisecond)
				reader.Clean()
			}
			testsuite.RunParallel(100, nil, cleanup, write, read, clean)

			err := reader.Close()
			require.NoError(t, err)

			require.True(t, bRead)
			require.True(t, closed)

			testsuite.IsDestroyed(t, reader)
		})

		t.Run("whole", func(t *testing.T) {
			var (
				r      *io.PipeReader
				w      *io.PipeWriter
				reader *ioReader
			)
			var (
				bRead  bool
				closed bool
			)
			onRead := func() { bRead = true }
			onClose := func() { closed = true }
			init := func() {
				r, w = io.Pipe()
				reader = newIOReader(logger.Test, r, onRead, onClose)
				reader.ReadLoop()
			}
			write := func() {
				_, err := w.Write(testdata)
				require.NoError(t, err)
			}
			read := func() {
				for i := 0; i < 100; i++ {
					output := reader.Read(i)
					if len(output) == 0 {
						return
					}
					require.Equal(t, testdata[i:], output)
				}
			}
			clean := func() {
				reader.Clean()
			}
			cleanup := func() {
				err := reader.Close()
				require.NoError(t, err)

				require.True(t, bRead)
				require.True(t, closed)
				bRead = false
				closed = false
			}
			testsuite.RunParallel(100, init, cleanup, write, read, clean)

			testsuite.IsDestroyed(t, reader)
		})
	})

	t.Run("with close", func(t *testing.T) {
		t.Run("part", func(t *testing.T) {
			r, w := io.Pipe()
			var (
				bRead  bool
				closed bool
			)
			onRead := func() { bRead = true }
			onClose := func() { closed = true }
			reader := newIOReader(logger.Test, r, onRead, onClose)
			reader.ReadLoop()

			write := func() {
				// writer pipe maybe closed
				_, _ = w.Write(testdata)
			}
			read := func() {
				for i := 0; i < 100; i++ {
					output := reader.Read(i)
					if len(output) == 0 {
						return
					}
					require.Equal(t, testdata[i:], output)
				}
			}
			clean := func() {
				reader.Clean()
			}
			close1 := func() {
				err := reader.Close()
				require.NoError(t, err)
			}
			cleanup := func() {
				reader.Clean()
			}
			testsuite.RunParallel(100, nil, cleanup, write, read, clean, close1)

			require.True(t, bRead)
			require.True(t, closed)

			testsuite.IsDestroyed(t, reader)
		})

		t.Run("whole", func(t *testing.T) {
			var (
				r      *io.PipeReader
				w      *io.PipeWriter
				reader *ioReader
			)
			var (
				bRead  bool
				closed bool
			)
			onRead := func() { bRead = true }
			onClose := func() { closed = true }

			init := func() {
				r, w = io.Pipe()
				reader = newIOReader(logger.Test, r, onRead, onClose)
				reader.ReadLoop()
			}
			write := func() {
				// writer pipe maybe closed
				_, _ = w.Write(testdata)
			}
			read := func() {
				for i := 0; i < 100; i++ {
					output := reader.Read(i)
					if len(output) == 0 {
						return
					}
					require.Equal(t, testdata[i:], output)
				}
			}
			clean := func() {
				reader.Clean()
			}
			close1 := func() {
				// maybe not read
				time.Sleep(25 * time.Millisecond)

				err := reader.Close()
				require.NoError(t, err)
			}
			cleanup := func() {
				reader.Clean()

				require.True(t, bRead)
				require.True(t, closed)
				bRead = false
				closed = false
			}
			testsuite.RunParallel(100, init, cleanup, write, read, clean, close1)

			testsuite.IsDestroyed(t, reader)
		})
	})
}

var (
	testIOManagerOptions = &IOManagerOptions{
		Interval: 25 * time.Millisecond,
	}
	testUserToken      = "test-user-token"
	testAnotherToken   = "test-another-token"
	testAdminToken     = "test-admin-token"
	testConsoleCommand = []byte("version\r\n")
)

func testReadDataFromIOObject(obj *IOObject) {
	start := 0
	for i := 0; i < 10; i++ {
		data := obj.Read(start)
		l := len(data)
		if l != 0 {
			fmt.Println(string(data))
			start += l
		}
		time.Sleep(minReadInterval)
	}
}

func TestIOObject(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	client := testGenerateClientAndLogin(t)

	ctx := context.Background()
	var (
		consoleID string
		read      bool
		closed    bool
		locked    bool
		unlocked  bool
	)
	handlers := IOEventHandlers{
		OnConsoleRead: func(id string) {
			require.Equal(t, consoleID, id)
			read = true
		},
		OnConsoleClosed: func(id string) {
			require.Equal(t, consoleID, id)
			closed = true
		},
		OnConsoleLocked: func(id, token string) {
			require.Equal(t, consoleID, id)
			require.Equal(t, testUserToken, token)
			locked = true
		},
		OnConsoleUnlocked: func(id, token string) {
			require.Equal(t, consoleID, id)
			if token != testUserToken && token != testAdminToken {
				t.Fatal("unexpected token:", token)
			}
			unlocked = true
		},
	}
	manager := NewIOManager(client, &handlers, testIOManagerOptions)

	console, err := manager.NewConsole(ctx, defaultWorkspace)
	require.NoError(t, err)
	consoleID = console.ToConsole().id

	t.Run("not locked", func(t *testing.T) {
		err := console.Write(testUserToken, testConsoleCommand)
		require.NoError(t, err)

		testReadDataFromIOObject(console)
	})

	t.Run("lock", func(t *testing.T) {
		ok := console.Lock(testUserToken)
		require.True(t, ok)

		fmt.Println("lock at:", console.LockAt())

		err := console.Write(testUserToken, testConsoleCommand)
		require.NoError(t, err)

		testReadDataFromIOObject(console)

		ok = console.Unlock(testUserToken)
		require.True(t, ok)
	})

	t.Run("write after another lock", func(t *testing.T) {
		ok := console.Lock(testUserToken)
		require.True(t, ok)

		err := console.Write(testUserToken, testConsoleCommand)
		require.NoError(t, err)

		testReadDataFromIOObject(console)

		err = console.Write(testAnotherToken, nil)
		require.Error(t, err)

		ok = console.Unlock(testUserToken)
		require.True(t, ok)

		err = console.Write(testAnotherToken, testConsoleCommand)
		require.NoError(t, err)

		testReadDataFromIOObject(console)
	})

	t.Run("lock after another lock", func(t *testing.T) {
		ok := console.Lock(testUserToken)
		require.True(t, ok)

		ok = console.Lock(testAnotherToken)
		require.False(t, ok)

		ok = console.Unlock(testUserToken)
		require.True(t, ok)
	})

	t.Run("unlock without lock", func(t *testing.T) {
		ok := console.Unlock(testUserToken)
		require.True(t, ok)
	})

	t.Run("unlock with another token", func(t *testing.T) {
		ok := console.Lock(testUserToken)
		require.True(t, ok)

		ok = console.Unlock(testAnotherToken)
		require.False(t, ok)

		ok = console.Unlock(testUserToken)
		require.True(t, ok)
	})

	t.Run("force unlock", func(t *testing.T) {
		ok := console.Lock(testUserToken)
		require.True(t, ok)

		console.ForceUnlock(testAdminToken)

		err = console.Write(testAnotherToken, testConsoleCommand)
		require.NoError(t, err)

		testReadDataFromIOObject(console)
	})

	t.Run("Clean", func(t *testing.T) {
		err := console.Write(testUserToken, testConsoleCommand)
		require.NoError(t, err)

		err = console.Clean(testUserToken)
		require.NoError(t, err)

		data := console.Read(0)
		require.Nil(t, data)
	})

	t.Run("clean after lock", func(t *testing.T) {
		ok := console.Lock(testUserToken)
		require.True(t, ok)

		err = console.Clean(testAnotherToken)
		require.Error(t, err)

		ok = console.Unlock(testUserToken)
		require.True(t, ok)
	})

	t.Run("Close", func(t *testing.T) {
		err := console.Close()
		require.NoError(t, err)
	})

	err = manager.Close()
	require.NoError(t, err)

	require.True(t, read)
	require.True(t, closed)
	require.True(t, locked)
	require.True(t, unlocked)

	err = client.ConsoleDestroy(ctx, consoleID)
	require.NoError(t, err)

	testsuite.IsDestroyed(t, console)
	testsuite.IsDestroyed(t, manager)

	err = client.Close()
	require.NoError(t, err)

	testsuite.IsDestroyed(t, client)
}

func TestIOObject_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

}
