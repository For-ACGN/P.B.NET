package testproxy

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestPoolAndManager(t *testing.T) {
	testsuite.InitHTTPServers(t)

	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	proxyPool, proxyMgr, certPool := PoolAndManager(t)

	// test balance
	balance, err := proxyPool.Get(TagBalance)
	require.NoError(t, err)

	// test http client
	transport := new(http.Transport)
	balance.HTTP(transport)
	testsuite.HTTPClient(t, transport, "localhost")

	testsuite.IsDestroyed(t, proxyPool)
	require.NoError(t, proxyMgr.Close())
	testsuite.IsDestroyed(t, proxyMgr)
	testsuite.IsDestroyed(t, certPool)
}
