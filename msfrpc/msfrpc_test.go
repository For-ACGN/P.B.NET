package msfrpc

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
	"project/internal/patch/msgpack"
	"project/internal/testsuite"
)

const (
	testCommand  = "msfrpcd"
	testHost     = "127.0.0.1"
	testPort     = 55553
	testUsername = "msf"
	testPassword = "msf"

	testInvalidToken = "invalid token"
)

func TestMain(m *testing.M) {
	fmt.Println("[info] start Metasploit RPC service")
	cmd := exec.Command(testCommand, "-a", testHost, "-U", testUsername, "-P", testPassword)
	err := cmd.Start()
	testsuite.CheckErrorInTestMain(err)
	// if panic, kill it.
	defer func() {
		_ = cmd.Process.Kill()
		os.Exit(1)
	}()
	// wait some time for start Metasploit RPC service
	// stdout and stderr can't read any data, so use time.Sleep
	fmt.Println("[info] wait 10 seconds for wait Metasploit RPC service")
	// TODO remove comment
	// time.Sleep(10 * time.Second)
	exitCode := m.Run()
	// create msfrpc
	msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, nil)
	testsuite.CheckErrorInTestMain(err)
	err = msfrpc.AuthLogin()
	testsuite.CheckErrorInTestMain(err)
	// check leaks
	ctx := context.Background()
	for _, check := range []func(context.Context, *MSFRPC) bool{
		testMainCheckSession,
		testMainCheckJob,
		testMainCheckConsole,
		testMainCheckToken,
		testMainCheckThread,
	} {
		if check(ctx, msfrpc) {
			time.Sleep(time.Minute)
			return
		}
	}
	msfrpc.Kill()
	if !testsuite.Destroyed(msfrpc) {
		fmt.Println("[warning] msfrpc is not destroyed!")
		time.Sleep(time.Minute)
		return
	}
	// stop Metasploit RPC service
	_ = cmd.Process.Kill()
	os.Exit(exitCode)
}

func testMainCheckSession(ctx context.Context, msfrpc *MSFRPC) bool {
	var (
		list map[uint64]*SessionInfo
		err  error
	)
	for i := 0; i < 30; i++ {
		list, err = msfrpc.SessionList(ctx)
		testsuite.CheckErrorInTestMain(err)
		if len(list) == 0 {
			return false
		}
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("[warning] msfrpcd session leaks!")
	const format = "id: %d type: %s remote: %s\n"
	for id, session := range list {
		fmt.Printf(format, id, session.Type, session.TunnelPeer)
	}
	return true
}

func testMainCheckJob(ctx context.Context, msfrpc *MSFRPC) bool {
	var (
		list map[string]string
		err  error
	)
	for i := 0; i < 30; i++ {
		list, err = msfrpc.JobList(ctx)
		testsuite.CheckErrorInTestMain(err)
		if len(list) == 0 {
			return false
		}
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("[warning] msfrpcd job leaks!")
	const format = "id: %s name: %s\n"
	for id, name := range list {
		fmt.Printf(format, id, name)
	}
	return true
}

func testMainCheckConsole(ctx context.Context, msfrpc *MSFRPC) bool {
	var (
		list []*ConsoleInfo
		err  error
	)
	for i := 0; i < 30; i++ {
		list, err = msfrpc.ConsoleList(ctx)
		testsuite.CheckErrorInTestMain(err)
		if len(list) == 0 {
			return false
		}
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("[warning] msfrpcd console leaks!")
	const format = "id: %s prompt: %s\n"
	for i := 0; i < len(list); i++ {
		fmt.Printf(format, list[i].ID, list[i].Prompt)
	}
	return true
}

func testMainCheckToken(ctx context.Context, msfrpc *MSFRPC) bool {
	var (
		list []string
		err  error
	)
	for i := 0; i < 30; i++ {
		list, err = msfrpc.AuthTokenList(ctx)
		testsuite.CheckErrorInTestMain(err)
		// include self token
		if len(list) == 1 {
			return false
		}
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("[warning] msfrpcd token leaks!")
	for i := 0; i < len(list); i++ {
		fmt.Println(list[i])
	}
	return true
}

func testMainCheckThread(ctx context.Context, msfrpc *MSFRPC) bool {
	var (
		list map[uint64]*CoreThreadInfo
		err  error
	)
	for i := 0; i < 30; i++ {
		list, err = msfrpc.CoreThreadList(ctx)
		testsuite.CheckErrorInTestMain(err)
		// TODO [external] msfrpcd thread leaks
		// if you call SessionMeterpreterRead() or SessionMeterpreterWrite()
		// when you exit meterpreter shell. this thread is always sleep.
		// so deceive ourselves now.
		for id, thread := range list {
			if thread.Name == "StreamMonitorRemote" {
				delete(list, id)
			}
		}
		// 3 = internal(do noting)
		// 9 = start sessions scheduler(5) and session manager(1)
		l := len(list)
		if l == 3 || l == 9 {
			return false
		}
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("[warning] msfrpcd thread leaks!")
	const format = "id: %d\nname: %s\ncritical: %t\nstatus: %s\nstarted: %s\n\n"
	for i, t := range list {
		fmt.Printf(format, i, t.Name, t.Critical, t.Status, t.Started)
	}
	return true
}

func TestNewMSFRPC(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("ok", func(t *testing.T) {
		msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, nil)
		require.NoError(t, err)

		msfrpc.Kill()
		testsuite.IsDestroyed(t, msfrpc)
	})

	t.Run("invalid transport option", func(t *testing.T) {
		opts := Options{}
		opts.Transport.TLSClientConfig.RootCAs = []string{"foo ca"}
		msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, &opts)
		require.Error(t, err)
		require.Nil(t, msfrpc)
	})

	t.Run("disable TLS", func(t *testing.T) {
		opts := Options{DisableTLS: true}
		msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, &opts)
		require.NoError(t, err)
		require.NotNil(t, msfrpc)

		msfrpc.Kill()
		testsuite.IsDestroyed(t, msfrpc)
	})

	t.Run("custom handler", func(t *testing.T) {
		opts := Options{Handler: "hello"}
		msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, &opts)
		require.NoError(t, err)
		require.NotNil(t, msfrpc)

		msfrpc.Kill()
		testsuite.IsDestroyed(t, msfrpc)
	})
}

func TestMSFRPC_sendWithReplace(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, nil)
	require.NoError(t, err)
	err = msfrpc.AuthLogin()
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("failed to read from", func(t *testing.T) {
		// patch
		client := new(http.Client)
		patchFunc := func(interface{}, *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       testsuite.NewMockReadCloserWithReadError(),
			}, nil
		}
		pg := monkey.PatchInstanceMethod(client, "Do", patchFunc)
		defer pg.Unpatch()

		err = msfrpc.sendWithReplace(ctx, nil, nil, nil)
		require.EqualError(t, testsuite.ErrMockReadCloser, err.Error())
	})

	padding := func() {}

	t.Run("ok", func(t *testing.T) {
		request := AuthTokenListRequest{
			Method: MethodAuthTokenList,
			Token:  msfrpc.GetToken(),
		}
		var result AuthTokenListResult
		err = msfrpc.sendWithReplace(ctx, request, &result, padding)
		require.NoError(t, err)
	})

	t.Run("replace", func(t *testing.T) {
		request := AuthTokenListRequest{
			Method: MethodAuthTokenList,
			Token:  msfrpc.GetToken(),
		}
		var result AuthTokenListResult
		err = msfrpc.sendWithReplace(ctx, request, padding, &result)
		require.NoError(t, err)
	})

	msfrpc.Kill()
	testsuite.IsDestroyed(t, msfrpc)
}

func TestMSFRPC_send(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	ctx := context.Background()

	t.Run("invalid request", func(t *testing.T) {
		msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, nil)
		require.NoError(t, err)

		err = msfrpc.send(ctx, func() {}, nil)
		require.Error(t, err)

		msfrpc.Kill()
		testsuite.IsDestroyed(t, msfrpc)
	})

	// start mock server(like msfrpcd)
	const testError = "test error"

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/200", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = msgpack.NewEncoder(w).Encode([]byte("ok"))
	})
	serverMux.HandleFunc("/500_ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		var msfErr MSFError
		msfErr.ErrorMessage = testError
		msfErr.ErrorCode = 500
		_ = msgpack.NewEncoder(w).Encode(msfErr)
	})
	serverMux.HandleFunc("/500_failed", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("invalid data"))
	})
	serverMux.HandleFunc("/401", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	serverMux.HandleFunc("/403", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	serverMux.HandleFunc("/unknown", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
	server := http.Server{
		Addr:    "127.0.0.1:0",
		Handler: serverMux,
	}
	port := testsuite.RunHTTPServer(t, "tcp", &server)
	defer func() { _ = server.Close() }()
	portNum, err := strconv.Atoi(port)
	require.NoError(t, err)
	portNumber := uint16(portNum)

	t.Run("internal server error_ok", func(t *testing.T) {
		opts := Options{
			DisableTLS: true,
			Handler:    "500_ok",
		}
		msfrpc, err := NewMSFRPC(testHost, portNumber, testUsername, testPassword, &opts)
		require.NoError(t, err)

		err = msfrpc.send(ctx, nil, nil)
		require.EqualError(t, err, testError)

		msfrpc.Kill()
		testsuite.IsDestroyed(t, msfrpc)
	})

	t.Run("internal server error_failed", func(t *testing.T) {
		opts := Options{
			DisableTLS: true,
			Handler:    "500_failed",
		}
		msfrpc, err := NewMSFRPC(testHost, portNumber, testUsername, testPassword, &opts)
		require.NoError(t, err)

		err = msfrpc.send(ctx, nil, nil)
		require.Error(t, err)

		msfrpc.Kill()
		testsuite.IsDestroyed(t, msfrpc)
	})

	t.Run("unauthorized", func(t *testing.T) {
		opts := Options{
			DisableTLS: true,
			Handler:    "401",
		}
		msfrpc, err := NewMSFRPC(testHost, portNumber, testUsername, testPassword, &opts)
		require.NoError(t, err)

		err = msfrpc.send(ctx, nil, nil)
		require.EqualError(t, err, "token is invalid")

		msfrpc.Kill()
		testsuite.IsDestroyed(t, msfrpc)
	})

	t.Run("forbidden", func(t *testing.T) {
		opts := Options{
			DisableTLS: true,
			Handler:    "403",
		}
		msfrpc, err := NewMSFRPC(testHost, portNumber, testUsername, testPassword, &opts)
		require.NoError(t, err)

		err = msfrpc.send(ctx, nil, nil)
		require.EqualError(t, err, "token is not granted access to the resource")

		msfrpc.Kill()
		testsuite.IsDestroyed(t, msfrpc)
	})

	t.Run("not found", func(t *testing.T) {
		opts := Options{
			DisableTLS: true,
			Handler:    "not_found",
		}
		msfrpc, err := NewMSFRPC(testHost, portNumber, testUsername, testPassword, &opts)
		require.NoError(t, err)

		err = msfrpc.send(ctx, nil, nil)
		require.EqualError(t, err, "the request was sent to an invalid URL")

		msfrpc.Kill()
		testsuite.IsDestroyed(t, msfrpc)

	})

	t.Run("other status code", func(t *testing.T) {
		opts := Options{
			DisableTLS: true,
			Handler:    "unknown",
		}
		msfrpc, err := NewMSFRPC(testHost, portNumber, testUsername, testPassword, &opts)
		require.NoError(t, err)

		err = msfrpc.send(ctx, nil, nil)
		require.EqualError(t, err, "202 Accepted")

		msfrpc.Kill()
		testsuite.IsDestroyed(t, msfrpc)
	})

	t.Run("parallel", func(t *testing.T) {
		opts := Options{
			DisableTLS: true,
			Handler:    "200",
		}
		msfrpc, err := NewMSFRPC(testHost, portNumber, testUsername, testPassword, &opts)
		require.NoError(t, err)

		f1 := func() {
			testdata := []byte{0x00, 0x01}
			var result []byte
			err := msfrpc.send(ctx, &testdata, &result)
			require.NoError(t, err)
			require.Equal(t, []byte("ok"), result)
		}
		f2 := func() {
			testdata := []byte{0x02, 0x03}
			var result []byte
			err := msfrpc.send(ctx, &testdata, &result)
			require.NoError(t, err)
			require.Equal(t, []byte("ok"), result)
		}
		testsuite.RunParallel(f1, f2)

		msfrpc.Kill()
		testsuite.IsDestroyed(t, msfrpc)
	})
}

func testPatchSend(f func()) {
	patch := func(context.Context, string, string, io.Reader) (*http.Request, error) {
		return nil, monkey.ErrMonkey
	}
	pg := monkey.Patch(http.NewRequestWithContext, patch)
	defer pg.Unpatch()
	f()
}

func TestMSFRPC_AuthLogin(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, nil)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		err = msfrpc.AuthLogin()
		require.NoError(t, err)
	})

	t.Run("failed to login", func(t *testing.T) {
		msfrpc.password = "foo"
		err = msfrpc.AuthLogin()
		require.EqualError(t, err, "Login Failed")

		msfrpc.password = testUsername
	})

	t.Run("failed to send", func(t *testing.T) {
		testPatchSend(func() {
			err = msfrpc.AuthLogin()
			monkey.IsMonkeyError(t, err)
		})
	})

	msfrpc.Kill()
	testsuite.IsDestroyed(t, msfrpc)
}

func TestMSFRPC_AuthLogout(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, nil)
	require.NoError(t, err)

	t.Run("logout self", func(t *testing.T) {
		err = msfrpc.AuthLogin()
		require.NoError(t, err)

		err = msfrpc.AuthLogout(msfrpc.GetToken())
		require.NoError(t, err)
	})

	t.Run("logout invalid token", func(t *testing.T) {
		err = msfrpc.AuthLogin()
		require.NoError(t, err)

		err = msfrpc.AuthLogout(testInvalidToken)
		require.EqualError(t, err, ErrInvalidTokenFriendly)
	})

	t.Run("failed to send", func(t *testing.T) {
		testPatchSend(func() {
			err = msfrpc.AuthLogout(msfrpc.GetToken())
			monkey.IsMonkeyError(t, err)
		})
	})

	msfrpc.Kill()
	testsuite.IsDestroyed(t, msfrpc)
}

func TestMSFRPC_AuthTokenList(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, nil)
	require.NoError(t, err)
	err = msfrpc.AuthLogin()
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		token := msfrpc.GetToken()
		list, err := msfrpc.AuthTokenList(ctx)
		require.NoError(t, err)
		var exist bool
		for i := 0; i < len(list); i++ {
			t.Log(list[i])
			if token == list[i] {
				exist = true
			}
		}
		require.True(t, exist)
	})

	t.Run("invalid authentication token", func(t *testing.T) {
		token := msfrpc.GetToken()
		defer msfrpc.SetToken(token)
		msfrpc.SetToken(testInvalidToken)

		list, err := msfrpc.AuthTokenList(ctx)
		require.EqualError(t, err, ErrInvalidTokenFriendly)
		require.Nil(t, list)
	})

	t.Run("failed to send", func(t *testing.T) {
		testPatchSend(func() {
			list, err := msfrpc.AuthTokenList(ctx)
			monkey.IsMonkeyError(t, err)
			require.Nil(t, list)
		})
	})

	msfrpc.Kill()
	testsuite.IsDestroyed(t, msfrpc)
}

func TestMSFRPC_AuthTokenGenerate(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, nil)
	require.NoError(t, err)
	err = msfrpc.AuthLogin()
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		token, err := msfrpc.AuthTokenGenerate(ctx)
		require.NoError(t, err)
		t.Log(token)

		tokens, err := msfrpc.AuthTokenList(ctx)
		require.NoError(t, err)
		require.Contains(t, tokens, token)

		err = msfrpc.AuthTokenRemove(ctx, token)
		require.NoError(t, err)
	})

	t.Run("invalid authentication token", func(t *testing.T) {
		token := msfrpc.GetToken()
		defer msfrpc.SetToken(token)
		msfrpc.SetToken(testInvalidToken)

		token, err := msfrpc.AuthTokenGenerate(ctx)
		require.EqualError(t, err, ErrInvalidTokenFriendly)
		require.Zero(t, token)
	})

	t.Run("failed to send", func(t *testing.T) {
		testPatchSend(func() {
			token, err := msfrpc.AuthTokenGenerate(ctx)
			monkey.IsMonkeyError(t, err)
			require.Zero(t, token)
		})
	})

	msfrpc.Kill()
	testsuite.IsDestroyed(t, msfrpc)
}

func TestMSFRPC_AuthTokenAdd(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, nil)
	require.NoError(t, err)
	err = msfrpc.AuthLogin()
	require.NoError(t, err)

	ctx := context.Background()
	const token = "TEST0123456789012345678901234567"

	t.Run("success", func(t *testing.T) {
		err := msfrpc.AuthTokenAdd(ctx, token)
		require.NoError(t, err)

		tokens, err := msfrpc.AuthTokenList(ctx)
		require.NoError(t, err)
		require.Contains(t, tokens, token)

		err = msfrpc.AuthTokenRemove(ctx, token)
		require.NoError(t, err)
	})

	t.Run("add invalid token", func(t *testing.T) {
		err := msfrpc.AuthTokenAdd(ctx, testInvalidToken)
		require.NoError(t, err)

		tokens, err := msfrpc.AuthTokenList(ctx)
		require.NoError(t, err)
		require.Contains(t, tokens, testInvalidToken)

		err = msfrpc.AuthTokenRemove(ctx, testInvalidToken)
		require.NoError(t, err)
	})

	t.Run("invalid authentication token", func(t *testing.T) {
		// due to the last sub test added testInvalidToken,
		// so must change the token that will be set
		former := msfrpc.GetToken()
		defer msfrpc.SetToken(former)
		msfrpc.SetToken(testInvalidToken + "foo")
		err := msfrpc.AuthTokenAdd(ctx, token)
		require.EqualError(t, err, ErrInvalidTokenFriendly)
	})

	t.Run("failed to send", func(t *testing.T) {
		testPatchSend(func() {
			err := msfrpc.AuthTokenAdd(ctx, token)
			monkey.IsMonkeyError(t, err)
		})
	})

	msfrpc.Kill()
	testsuite.IsDestroyed(t, msfrpc)
}

func TestMSFRPC_AuthTokenRemove(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, nil)
	require.NoError(t, err)
	err = msfrpc.AuthLogin()
	require.NoError(t, err)

	ctx := context.Background()
	const token = "TEST0123456789012345678901234567"

	t.Run("success", func(t *testing.T) {
		err := msfrpc.AuthTokenRemove(ctx, token)
		require.NoError(t, err)

		tokens, err := msfrpc.AuthTokenList(ctx)
		require.NoError(t, err)
		require.NotContains(t, tokens, token)
	})

	t.Run("remove invalid token", func(t *testing.T) {
		err := msfrpc.AuthTokenAdd(ctx, testInvalidToken)
		require.NoError(t, err)

		err = msfrpc.AuthTokenRemove(ctx, testInvalidToken)
		require.NoError(t, err)

		// doesn't exists
		err = msfrpc.AuthTokenRemove(ctx, testInvalidToken)
		require.NoError(t, err)

		tokens, err := msfrpc.AuthTokenList(ctx)
		require.NoError(t, err)
		require.NotContains(t, tokens, testInvalidToken)
	})

	t.Run("invalid authentication token", func(t *testing.T) {
		former := msfrpc.GetToken()
		defer msfrpc.SetToken(former)
		msfrpc.SetToken(testInvalidToken)

		err := msfrpc.AuthTokenRemove(ctx, token)
		require.EqualError(t, err, ErrInvalidTokenFriendly)
	})

	t.Run("failed to send", func(t *testing.T) {
		testPatchSend(func() {
			err := msfrpc.AuthTokenRemove(ctx, token)
			monkey.IsMonkeyError(t, err)
		})
	})

	msfrpc.Kill()
	testsuite.IsDestroyed(t, msfrpc)
}

func TestMSFRPC_Close(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, nil)
	require.NoError(t, err)

	t.Run("ok", func(t *testing.T) {
		err = msfrpc.AuthLogin()
		require.NoError(t, err)
		err = msfrpc.Close()
		require.NoError(t, err)
	})

	t.Run("failed", func(t *testing.T) {
		err = msfrpc.Close()
		require.Error(t, err)
	})

	msfrpc.Kill()
	testsuite.IsDestroyed(t, msfrpc)
}
