package testrole

import (
	"sync"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/cert/certmgr"
	"project/internal/cert/certpool"
	"project/internal/dns"
	"project/internal/logger"
	"project/internal/proxy"
	"project/internal/proxy/http"
	"project/internal/proxy/socks"
	"project/internal/testsuite"
	"project/internal/timesync"
)

// CertPool is used to provide a certificate pool for test.
func CertPool(t require.TestingT) certmgr.NBCertPool {
	pool, err := certpool.System()
	require.NoError(t, err)
	certs := pool.Certs()
	cp := certmgr.NBCertPool{}
	for i := 0; i < len(certs); i++ {
		cp.PublicRootCACerts = append(cp.PublicRootCACerts, certs[i].Raw)
	}
	return cp
}

// proxy client tag
const (
	Socks5Tag    = "test_socks5"
	HTTPProxyTag = "test_http"
)

var (
	socks5Server    *socks.Server
	httpProxyServer *http.Server
	initProxyOnce   sync.Once
	wg              sync.WaitGroup
)

// ProxyClients is used to deploy proxy server and return corresponding proxy client.
func ProxyClients(t require.TestingT) []*proxy.Client {
	initProxyOnce.Do(func() {
		var err error
		// socks5 server
		socks5Server, err = socks.NewSocks5Server("test", logger.Test, nil)
		require.NoError(t, err)
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := socks5Server.ListenAndServe("tcp", "localhost:0")
			require.NoError(t, err)
		}()
		// http proxy server
		httpProxyServer, err = http.NewHTTPServer("test", logger.Test, nil)
		require.NoError(t, err)
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := httpProxyServer.ListenAndServe("tcp", "localhost:0")
			require.NoError(t, err)
		}()
		time.Sleep(250 * time.Millisecond)
	})
	return []*proxy.Client{
		{
			Tag:     Socks5Tag,
			Mode:    proxy.ModeSocks5,
			Network: "tcp",
			Address: socks5Server.Addresses()[0].String(),
		},
		{
			Tag:     HTTPProxyTag,
			Mode:    proxy.ModeHTTP,
			Network: "tcp",
			Address: httpProxyServer.Addresses()[0].String(),
		},
	}
}

// DNSServers is used to provide DNS servers for test.
func DNSServers() map[string]*dns.Server {
	servers := make(map[string]*dns.Server)
	if testsuite.IPv4Enabled {
		servers["test_udp_ipv4"] = &dns.Server{
			Method:  "udp",
			Address: "8.8.8.8:53",
		}
		servers["test_tcp_ipv4"] = &dns.Server{
			Method:  "tcp",
			Address: "8.8.8.8:53",
		}
		servers["test_dot_ipv4"] = &dns.Server{
			Method:  "dot",
			Address: "1.1.1.1:853",
		}
		servers["test_skip_ipv4"] = &dns.Server{
			Method:   "udp",
			Address:  "1.1.1.1:53",
			SkipTest: true,
		}
	}
	if testsuite.IPv6Enabled {
		servers["test_udp_ipv6"] = &dns.Server{
			Method:  "udp",
			Address: "[2606:4700:4700::1111]:53",
		}
		servers["test_tcp_ipv6"] = &dns.Server{
			Method:  "tcp",
			Address: "[2606:4700:4700::1111]:53",
		}
		servers["test_dot_ipv6"] = &dns.Server{
			Method:  "dot",
			Address: "[2606:4700:4700::1111]:853",
		}
		servers["test_skip_ipv6"] = &dns.Server{
			Method:   "udp",
			Address:  "[2606:4700:4700::1001]:53",
			SkipTest: true,
		}
	}
	servers["test_doh_ds"] = &dns.Server{
		Method:  "doh",
		Address: "https://cloudflare-dns.com/dns-query",
	}
	return servers
}

// TimeSyncerClients is used to provide time syncer clients for test.
func TimeSyncerClients() map[string]*timesync.Client {
	clients := make(map[string]*timesync.Client)
	config := `
timeout = "15s"

[request]
  url   = "https://www.cloudflare.com/"
  close = true

  [request.header]
    User-Agent      = ["Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:72.0) Gecko/20100101 Firefox/72.0"]
    Accept          = ["text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"]
    Accept-Language = ["zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2"]
    DNT             = ["1"]
    Pragma          = ["no-cache"]
    Cache-Control   = ["no-cache"]
`
	clients["test_http"] = &timesync.Client{
		Mode:   timesync.ModeHTTP,
		Config: config,
	}
	config = `address = "2.pool.ntp.org:123"`
	clients["test_ntp"] = &timesync.Client{
		Mode:   timesync.ModeNTP,
		Config: config,
	}
	return clients
}

// Clean is used to clean test data.
func Clean() {
	if socks5Server != nil {
		_ = socks5Server.Close()
	}
	if httpProxyServer != nil {
		_ = httpProxyServer.Close()
	}
	wg.Wait()
}
