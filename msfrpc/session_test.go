package msfrpc

import (
	"context"
	"encoding/hex"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/module/shellcode"
	"project/internal/patch/monkey"
	"project/internal/testsuite"
)

func testCreateSession(t *testing.T, msfrpc *MSFRPC, port string) {
	ctx := context.Background()

	// select payload
	const exploit = "multi/handler"
	opts := make(map[string]interface{})
	switch runtime.GOOS {
	case "windows":
		switch runtime.GOARCH {
		case "386":
			opts["PAYLOAD"] = "windows/meterpreter/reverse_tcp"
		case "amd64":
			opts["PAYLOAD"] = "windows/x64/meterpreter/reverse_tcp"
		default:
			t.Skip("only support 386 and amd64")
		}
	case "linux":
		switch runtime.GOARCH {
		case "386":
			opts["PAYLOAD"] = "linux/meterpreter/reverse_tcp"
		case "amd64":
			opts["PAYLOAD"] = "linux/x64/meterpreter/reverse_tcp"
		default:
			t.Skip("only support 386 and amd64")
		}
	default:
		t.Skip("only support windows and linux")
	}
	opts["EXITFUNC"] = "thread"
	opts["TARGET"] = 0
	opts["LHOST"] = "127.0.0.1"
	opts["LPORT"] = port

	// start handler
	result, err := msfrpc.ModuleExecute(ctx, "exploit", exploit, opts)
	require.NoError(t, err)
	var ok bool
	defer func() {
		if ok {
			return
		}
		jobID := strconv.FormatUint(result.JobID, 10)
		err = msfrpc.JobStop(jobID)
		require.NoError(t, err)
	}()

	// generate payload
	payload := opts["PAYLOAD"].(string)
	payloadOpts := NewModuleExecuteOptions()
	payloadOpts.Format = "raw"
	payloadOpts.DataStore["EXITFUNC"] = "thread"
	payloadOpts.DataStore["LHOST"] = "127.0.0.1"
	payloadOpts.DataStore["LPORT"] = port
	result, err = msfrpc.ModuleExecute(ctx, "payload", payload, payloadOpts)
	require.NoError(t, err)
	sc := []byte(result.Payload)
	t.Log("raw payload:", hex.EncodeToString(sc))

	// execute shellcode and wait
	go func() { _ = shellcode.Execute("", sc) }()
	time.Sleep(5 * time.Second)

	sessions, err := msfrpc.SessionList(ctx)
	require.NoError(t, err)
	if len(sessions) == 1 {
		ok = true
	}
}

func TestMSFRPC_SessionList(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	msfrpc, err := NewMSFRPC(testHost, testPort, testUsername, testPassword, nil)
	require.NoError(t, err)
	err = msfrpc.Login()
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		testCreateSession(t, msfrpc, "55001")

		sessions, err := msfrpc.SessionList(ctx)
		require.NoError(t, err)
		for id, session := range sessions {
			t.Logf("id: %d type: %s\n", id, session.Type)
		}

	})

	t.Run("invalid authentication token", func(t *testing.T) {
		token := msfrpc.GetToken()
		defer msfrpc.SetToken(token)
		msfrpc.SetToken(testInvalidToken)

		sessions, err := msfrpc.SessionList(ctx)
		require.EqualError(t, err, testErrInvalidToken)
		require.Nil(t, sessions)
	})

	t.Run("send failed", func(t *testing.T) {
		testPatchSend(func() {
			sessions, err := msfrpc.SessionList(ctx)
			monkey.IsMonkeyError(t, err)
			require.Nil(t, sessions)
		})
	})

	msfrpc.Kill()
	testsuite.IsDestroyed(t, msfrpc)
}
