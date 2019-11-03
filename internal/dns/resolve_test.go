package dns

import (
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/dns/dnsmessage"

	"project/internal/convert"
	"project/internal/testsuite"
)

const (
	// resolve target
	testDomain = "cloudflare-dns.com"
)

func TestSystemResolve(t *testing.T) {
	// ipv4
	ipList, err := systemResolve(TypeIPv4, testDomain)
	require.NoError(t, err)
	t.Log("system resolve ipv4:", ipList)
	// ipv6
	ipList, err = systemResolve(TypeIPv6, testDomain)
	require.NoError(t, err)
	t.Log("system resolve ipv6:", ipList)
	// invalid host
	ipList, err = systemResolve(TypeIPv4, "asd")
	require.Error(t, err)
	require.Equal(t, 0, len(ipList))
}

func TestCustomResolve(t *testing.T) {
	const domainPunycode = "错的是.世界"

	opts := &Options{
		dial:      net.DialTimeout,
		transport: &http.Transport{},
	}

	if testsuite.EnableIPv4() {
		const (
			udpServer = "1.1.1.1:53"
			tcpServer = "1.0.0.1:53"
			tlsIP     = "8.8.4.4:853"
			tlsDomain = "dns.google:853|8.8.8.8,8.8.4.4"
		)
		// udp
		ipList, err := customResolve(MethodUDP, udpServer, testDomain, TypeIPv4, opts)
		require.NoError(t, err)
		t.Log("UDP IPv4:", ipList)
		// tcp
		ipList, err = customResolve(MethodTCP, tcpServer, testDomain, TypeIPv4, opts)
		require.NoError(t, err)
		t.Log("TCP IPv4:", ipList)
		// dot ip mode
		ipList, err = customResolve(MethodDoT, tlsIP, testDomain, TypeIPv4, opts)
		require.NoError(t, err)
		t.Log("DOT-IP IPv4:", ipList)
		// dot domain mode
		ipList, err = customResolve(MethodDoT, tlsDomain, testDomain, TypeIPv4, opts)
		require.NoError(t, err)
		t.Log("DOT-Domain IPv4:", ipList)
		// punycode
		ipList, err = customResolve(MethodUDP, udpServer, domainPunycode, TypeIPv4, opts)
		require.NoError(t, err)
		t.Log("punycode:", ipList)
	}

	if testsuite.EnableIPv6() {
		const (
			udpServer = "[2606:4700:4700::1111]:53"
			tcpServer = "[2606:4700:4700::1001]:53"
			TLSIP     = "[2606:4700:4700::64]:853"
			TLSDomain = "cloudflare-dns.com:853|2606:4700:4700::1111,2606:4700:4700::1001"
		)
		// udp
		ipList, err := customResolve(MethodUDP, udpServer, testDomain, TypeIPv6, opts)
		require.NoError(t, err)
		t.Log("UDP IPv6:", ipList)
		// tcp
		ipList, err = customResolve(MethodTCP, tcpServer, testDomain, TypeIPv6, opts)
		require.NoError(t, err)
		t.Log("TCP IPv6:", ipList)
		// dot ip mode
		ipList, err = customResolve(MethodDoT, TLSIP, testDomain, TypeIPv6, opts)
		require.NoError(t, err)
		t.Log("DOT-IP IPv6:", ipList)
		// dot domain mode
		ipList, err = customResolve(MethodDoT, TLSDomain, testDomain, TypeIPv6, opts)
		require.NoError(t, err)
		t.Log("DOT-Domain IPv6:", ipList)
		// punycode
		ipList, err = customResolve(MethodUDP, udpServer, domainPunycode, TypeIPv6, opts)
		require.NoError(t, err)
		t.Log("punycode:", ipList)
	}

	// doh
	const dnsDOH = "https://cloudflare-dns.com/dns-query"
	ipList, err := customResolve(MethodDoH, dnsDOH, testDomain, TypeIPv4, opts)
	require.NoError(t, err)
	t.Log("DOH:", ipList)

	// resolve ip
	const dnsServer = "1.0.0.1:53"
	ipList, err = customResolve(MethodUDP, dnsServer, "1.1.1.1", TypeIPv4, opts)
	require.NoError(t, err)
	require.Equal(t, []string{"1.1.1.1"}, ipList)

	// empty domain
	ipList, err = customResolve(MethodUDP, dnsServer, "", TypeIPv4, opts)
	require.Error(t, err)
	require.Equal(t, 0, len(ipList))

	// resolve failed
	opts.Timeout = time.Second
	ipList, err = customResolve(MethodUDP, "0.0.0.0:1", domainPunycode, TypeIPv4, opts)
	require.Error(t, err)
	require.Equal(t, 0, len(ipList))
}

var (
	testDNSMessage = packMessage(dnsmessage.TypeA, testDomain)
)

func TestDialUDP(t *testing.T) {
	opt := &Options{dial: net.DialTimeout}
	if testsuite.EnableIPv4() {
		msg, err := dialUDP("8.8.8.8:53", testDNSMessage, opt)
		require.NoError(t, err)
		ipList, err := unpackMessage(msg)
		require.NoError(t, err)
		t.Log("UDP (IPv4 DNS Server):", ipList)
	}
	if testsuite.EnableIPv6() {
		msg, err := dialUDP("[2606:4700:4700::1001]:53", testDNSMessage, opt)
		require.NoError(t, err)
		ipList, err := unpackMessage(msg)
		require.NoError(t, err)
		t.Log("UDP (IPv6 DNS Server):", ipList)
	}
	// unknown network
	opt.Network = "foo network"
	_, err := dialUDP("", nil, opt)
	require.Error(t, err)
	// no port
	opt.Network = "udp"
	_, err = dialUDP("1.2.3.4", nil, opt)
	require.Error(t, err)
	// no response
	opt.Timeout = time.Second
	if testsuite.EnableIPv4() {
		_, err = dialUDP("1.2.3.4:23421", nil, opt)
		require.Equal(t, ErrNoConnection, err)
	}
	if testsuite.EnableIPv6() {
		_, err = dialUDP("[::1]:23421", nil, opt)
		require.Equal(t, ErrNoConnection, err)
	}
}

func TestDialTCP(t *testing.T) {
	opt := &Options{dial: net.DialTimeout}
	if testsuite.EnableIPv4() {
		msg, err := dialTCP("8.8.8.8:53", testDNSMessage, opt)
		require.NoError(t, err)
		ipList, err := unpackMessage(msg)
		require.NoError(t, err)
		t.Log("TCP (IPv4 DNS Server):", ipList)
	}
	if testsuite.EnableIPv6() {
		msg, err := dialTCP("[2606:4700:4700::1001]:53", testDNSMessage, opt)
		require.NoError(t, err)
		ipList, err := unpackMessage(msg)
		require.NoError(t, err)
		t.Log("TCP (IPv6 DNS Server):", ipList)
	}
	// unknown network
	opt.Network = "foo network"
	_, err := dialTCP("", nil, opt)
	require.Error(t, err)
	// no port
	opt.Network = "tcp"
	_, err = dialTCP("1.2.3.4", nil, opt)
	require.Error(t, err)
}

func TestDialDoT(t *testing.T) {
	opt := &Options{dial: net.DialTimeout}
	if testsuite.EnableIPv4() {
		const (
			dnsServerIPV4 = "8.8.8.8:853"
			dnsDomainIPv4 = "dns.google:853|8.8.8.8,8.8.4.4"
		)
		// IP mode
		msg, err := dialDoT(dnsServerIPV4, testDNSMessage, opt)
		require.NoError(t, err)
		ipList, err := unpackMessage(msg)
		require.NoError(t, err)
		t.Log("DoT-IP (IPv4 DNS Server):", ipList)
		// domain mode
		msg, err = dialDoT(dnsDomainIPv4, testDNSMessage, opt)
		require.NoError(t, err)
		ipList, err = unpackMessage(msg)
		require.NoError(t, err)
		t.Log("DoT-Domain (IPv4 DNS Server):", ipList)
	}
	if testsuite.EnableIPv6() {
		const (
			dnsServerIPv6 = "[2606:4700:4700::64]:853"
			dnsDomainIPv6 = "cloudflare-dns.com:853|2606:4700:4700::1111,2606:4700:4700::1001"
		)
		// IP mode
		msg, err := dialDoT(dnsServerIPv6, testDNSMessage, opt)
		require.NoError(t, err)
		ipList, err := unpackMessage(msg)
		require.NoError(t, err)
		t.Log("DoT-IP (IPv6 DNS Server):", ipList)
		// domain mode
		msg, err = dialDoT(dnsDomainIPv6, testDNSMessage, opt)
		require.NoError(t, err)
		ipList, err = unpackMessage(msg)
		require.NoError(t, err)
		t.Log("DoT-Domain (IPv6 DNS Server):", ipList)
	}
	// unknown network
	opt.Network = "foo network"
	_, err := dialDoT("", nil, opt)
	require.Error(t, err)
	// no port(ip mode)
	opt.Network = "tcp"
	_, err = dialDoT("1.2.3.4", nil, opt)
	require.Error(t, err)
	// dial failed
	_, err = dialDoT("127.0.0.1:888", nil, opt)
	require.Error(t, err)
	// error ip(domain mode)
	_, err = dialDoT("dns.google:853|127.0.0.1", nil, opt)
	require.Equal(t, ErrNoConnection, err)
	// no port(domain mode)
	_, err = dialDoT("dns.google|1.2.3.235", nil, opt)
	require.Error(t, err)
	// invalid config
	cfg := "asd:153|xxx|xxx"
	_, err = dialDoT(cfg, nil, opt)
	require.Errorf(t, err, "invalid address: %s", cfg)
}

func TestDialDoH(t *testing.T) {
	const dnsServer = "https://cloudflare-dns.com/dns-query"
	opt := &Options{transport: new(http.Transport)}
	// get
	resp, err := dialDoH(dnsServer, testDNSMessage, opt)
	require.NoError(t, err)
	ipList, err := unpackMessage(resp)
	require.NoError(t, err)
	t.Log("DoH GET:", ipList)
	// post
	url := dnsServer + "#" + strings.Repeat("a", 2048)
	resp, err = dialDoH(url, testDNSMessage, opt)
	require.NoError(t, err)
	ipList, err = unpackMessage(resp)
	require.NoError(t, err)
	t.Log("DoH POST:", ipList)
	// invalid doh server
	_, err = dialDoH("foo\n", testDNSMessage, opt)
	require.Error(t, err)
	url = "foo\n" + "#" + strings.Repeat("a", 2048)
	_, err = dialDoH(url, testDNSMessage, opt)
	require.Error(t, err)
	// unreachable doh server
	_, err = dialDoH("https://1.2.3.4/", testDNSMessage, opt)
	require.Error(t, err)
}

func TestFailedToSendMessage(t *testing.T) {
	// failed to write message
	server, client := net.Pipe()
	_ = server.Close()
	_, err := sendMessage(client, testDNSMessage, time.Second)
	require.Error(t, err)

	// failed to read response size
	server, client = net.Pipe()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := io.ReadFull(server, make([]byte, headerSize+len(testDNSMessage)))
		require.NoError(t, err)
		_ = server.Close()
	}()
	_, err = sendMessage(client, testDNSMessage, time.Second)
	require.Error(t, err)
	wg.Wait()

	// failed to read response
	server, client = net.Pipe()
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := io.ReadFull(server, make([]byte, headerSize+len(testDNSMessage)))
		require.NoError(t, err)
		_, _ = server.Write(convert.Uint16ToBytes(4))
		_ = server.Close()
	}()
	_, err = sendMessage(client, testDNSMessage, time.Second)
	require.Error(t, err)
	wg.Wait()
}
