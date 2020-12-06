// +build go1.10,!go1.11

package http

import (
	"crypto/tls"
)

// Clone returns a deep copy of t's exported fields.
//
// From go1.13(changed)
func (t *Transport) Clone() *Transport {
	t.nextProtoOnce.Do(t.onceSetNextProtoDefaults)
	t2 := &Transport{
		Proxy:       t.Proxy,
		DialContext: t.DialContext,
		Dial:        t.Dial,
		DialTLS:     t.DialTLS,
		// DialTLSContext:      t.DialTLSContext,
		TLSHandshakeTimeout: t.TLSHandshakeTimeout,
		DisableKeepAlives:   t.DisableKeepAlives,
		DisableCompression:  t.DisableCompression,
		MaxIdleConns:        t.MaxIdleConns,
		MaxIdleConnsPerHost: t.MaxIdleConnsPerHost,
		// MaxConnsPerHost:        t.MaxConnsPerHost,
		IdleConnTimeout:       t.IdleConnTimeout,
		ResponseHeaderTimeout: t.ResponseHeaderTimeout,
		ExpectContinueTimeout: t.ExpectContinueTimeout,
		ProxyConnectHeader:    t.ProxyConnectHeader.Clone(),
		// GetProxyConnectHeader:  t.GetProxyConnectHeader,
		MaxResponseHeaderBytes: t.MaxResponseHeaderBytes,
		// ForceAttemptHTTP2:      t.ForceAttemptHTTP2,
		// WriteBufferSize:        t.WriteBufferSize,
		// ReadBufferSize:         t.ReadBufferSize,
	}
	if t.TLSClientConfig != nil {
		t2.TLSClientConfig = t.TLSClientConfig.Clone()
	}
	if len(t.TLSNextProto) != 0 { // fix missed tlsNextProtoWasNil structure field.
		npm := map[string]func(authority string, c *tls.Conn) RoundTripper{}
		for k, v := range t.TLSNextProto {
			npm[k] = v
		}
		t2.TLSNextProto = npm
	}
	return t2
}
