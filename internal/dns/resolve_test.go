package dns

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/dns/dnsmessage"
)

const (
	dnsServer        = "8.8.8.8:53"
	dnsTLSDomainMode = "dns.google:853|8.8.8.8,8.8.4.4"
	dnsDOH           = "https://cloudflare-dns.com/dns-query"
	domain           = "cloudflare-dns.com"
	domainPunycode   = "münchen.com"

	// dnsDOH           = "https://cloudflare-dns.com/dns-query"
	// dnsDOH           = "https://mozilla.cloudflare-dns.com/dns-query"
)

func TestResolve(t *testing.T) {
	opt := Options{}
	// udp
	ipList, err := resolve(dnsServer, domain, &opt)
	require.NoError(t, err)
	t.Log("UDP IPv4:", ipList)
	// punycode
	ipList, err = resolve(dnsServer, domainPunycode, &opt)
	require.NoError(t, err)
	t.Log("UDP IPv4 punycode:", ipList)
	// tcp
	opt.Method = TCP
	opt.Type = IPv6
	ipList, err = resolve(dnsServer, domain, &opt)
	require.NoError(t, err)
	t.Log("TCP IPv6:", ipList)
	// tls
	opt.Method = TLS
	opt.Type = IPv4
	ipList, err = resolve(dnsTLSDomainMode, domain, &opt)
	require.NoError(t, err)
	t.Log("TLS IPv4:", ipList)
	// doh
	opt.Method = DOH
	ipList, err = resolve(dnsDOH, domain, &opt)
	require.NoError(t, err)
	t.Log("DOH IPv4:", ipList)
	// is ip
	ipList, err = resolve(dnsServer, "8.8.8.8", &opt)
	require.NoError(t, err)
	require.Equal(t, "8.8.8.8", ipList[0])
	ipList, err = resolve(dnsServer, "::1", &opt)
	require.NoError(t, err)
	require.Equal(t, "::1", ipList[0])
	// not domain
	_, err = resolve(dnsServer, "xxx-", &opt)
	require.Error(t, err)
	require.Equal(t, "invalid domain name: xxx-", err.Error())
	// invalid Type
	opt.Type = "foo"
	_, err = resolve(dnsServer, domain, &opt)
	require.Error(t, err)
	require.Equal(t, "unknown type: foo", err.Error())
	// invalid method
	opt.Type = IPv4
	opt.Method = "foo"
	_, err = resolve(dnsServer, domain, &opt)
	require.Error(t, err)
	require.Equal(t, "unknown method: foo", err.Error())
	// dial failed
	opt.Network = "udp"
	opt.Method = UDP
	opt.Timeout = time.Millisecond * 500
	_, err = resolve("8.8.8.8:153", domain, &opt)
	require.Equal(t, ErrNoConnection, err)
}

func TestIsDomainName(t *testing.T) {
	require.True(t, IsDomainName("asd.com"))
	require.True(t, IsDomainName("asd-asd.com"))
	// invalid domain
	require.False(t, IsDomainName(""))
	require.False(t, IsDomainName(string([]byte{255, 254, 12, 35})))
	require.False(t, IsDomainName("asd-"))
	require.False(t, IsDomainName("asd.-"))
	require.False(t, IsDomainName("asd.."))
	require.False(t, IsDomainName(strings.Repeat("a", 64)+".com"))
}

func TestDialUDP(t *testing.T) {
	opt := Options{
		Network: "udp",
		dial:    net.Dial,
	}
	msg := packMessage(dnsmessage.TypeA, domain)
	msg, err := dialUDP(dnsServer, msg, &opt)
	require.NoError(t, err)
	ipList, err := unpackMessage(msg)
	require.NoError(t, err)
	t.Log("UDP IPv4:", ipList)
	// no port
	_, err = dialUDP("1.2.3.4", msg, &opt)
	require.Error(t, err)
	// no response
	_, err = dialUDP("1.2.3.4:23421", msg, &opt)
	require.Equal(t, ErrNoConnection, err)
}

func TestDialTCP(t *testing.T) {
	opt := Options{
		Network: "tcp",
		dial:    net.Dial,
	}
	msg := packMessage(dnsmessage.TypeA, domain)
	msg, err := dialTCP(dnsServer, msg, &opt)
	require.NoError(t, err)
	ipList, err := unpackMessage(msg)
	require.NoError(t, err)
	t.Log("TCP IPv4:", ipList)
	// no port
	_, err = dialTCP("8.8.8.8", msg, &opt)
	require.Error(t, err)
}

func TestDialTLS(t *testing.T) {
	opt := Options{
		Network: "tcp",
		dial:    net.Dial,
	}
	msg := packMessage(dnsmessage.TypeA, domain)
	// domain name mode
	resp, err := dialTLS(dnsTLSDomainMode, msg, &opt)
	require.NoError(t, err)
	ipList, err := unpackMessage(resp)
	require.NoError(t, err)
	t.Log("TLS domain IPv4:", ipList)
	// ip mode
	resp, err = dialTLS("1.1.1.1:853", msg, &opt)
	require.NoError(t, err)
	ipList, err = unpackMessage(resp)
	require.NoError(t, err)
	t.Log("TLS ip IPv4:", ipList)
	// no port(ip mode)
	_, err = dialTLS("1.2.3.4", msg, &opt)
	require.Error(t, err)
	// dial failed
	_, err = dialTLS("127.0.0.1:888", msg, &opt)
	require.Error(t, err)
	// error ip(domain mode)
	_, err = dialTLS("dns.google:853|127.0.0.1", msg, &opt)
	require.Equal(t, ErrNoConnection, err)
	// no port(domain mode)
	_, err = dialTLS("dns.google|1.2.3.235", msg, &opt)
	require.Error(t, err)
	// invalid config
	_, err = dialTLS("asd:153|xxx|xxx", msg, &opt)
	require.Error(t, err)
	require.Equal(t, "invalid address: asd:153|xxx|xxx", err.Error())
}

func TestDialDOH(t *testing.T) {
	opt := Options{}
	msg := packMessage(dnsmessage.TypeA, domain)
	// get
	resp, err := dialDOH(dnsDOH, msg, &opt)
	require.NoError(t, err)
	ipList, err := unpackMessage(resp)
	require.NoError(t, err)
	t.Log("DOH get IPv4:", ipList)
	// post
	resp, err = dialDOH(dnsDOH+"#"+strings.Repeat("a", 2048), msg, &opt)
	require.NoError(t, err)
	ipList, err = unpackMessage(resp)
	require.NoError(t, err)
	t.Log("DOH post IPv4:", ipList)
	// invalid doh server
	_, err = dialDOH("foo\n", msg, &opt)
	require.Error(t, err)
	_, err = dialDOH("foo\n"+"#"+strings.Repeat("a", 2048), msg, &opt)
	require.Error(t, err)
	// Do failed
	_, err = dialDOH("http://asd.1dsa.asd", msg, &opt)
	require.Error(t, err)
}
