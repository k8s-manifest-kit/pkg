// Package httpclient provides constructors for private, non-pooled HTTP
// clients and transports built from explicit standard library defaults.
package httpclient

import (
	"net"
	"net/http"
	"time"
)

const (
	// dialTimeout bounds connection establishment.
	dialTimeout = 30 * time.Second
	// dialKeepAlive is the TCP keep-alive probe interval for fresh dials.
	dialKeepAlive = 30 * time.Second
	// maxIdleConns is the transport-wide idle connection bound retained from
	// the standard library defaults even though no idle connection is kept.
	maxIdleConns = 100
	// idleConnTimeout is the idle connection lifetime retained from the
	// standard library defaults even though no idle connection is kept.
	idleConnTimeout = 90 * time.Second
	// tlsHandshakeTimeout bounds the TLS handshake.
	tlsHandshakeTimeout = 10 * time.Second
	// expectContinueTimeout bounds waiting for a server's first response
	// byte after an "Expect: 100-continue" request header.
	expectContinueTimeout = 1 * time.Second
	// noIdleConnsPerHost disables retention of idle connections per host.
	noIdleConnsPerHost = -1
)

// NewClient returns a new http.Client that owns a private transport created
// by NewTransport. The returned client is independent of every other client
// and of the http package globals: it has no timeout, no cookie jar, and no
// custom redirect policy. Callers own the returned instance and its
// lifecycle.
func NewClient() *http.Client {
	return &http.Client{
		Transport: NewTransport(),
	}
}

// NewTransport returns a new http.Transport configured with explicit
// standard library defaults and no connection pooling: keep-alives are
// disabled and no idle connection is retained per host, so every request
// uses its own connection. Each call returns a fresh instance with a fresh
// net.Dialer, and the proxy is resolved from the process environment when a
// request is made. TLS verification keeps its zero value, HTTP/2 is
// attempted when the connection supports it, and callers own the returned
// instance and its lifecycle.
func NewTransport() *http.Transport {
	d := &net.Dialer{
		Timeout:   dialTimeout,
		KeepAlive: dialKeepAlive,
	}

	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           d.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          maxIdleConns,
		MaxIdleConnsPerHost:   noIdleConnsPerHost,
		IdleConnTimeout:       idleConnTimeout,
		TLSHandshakeTimeout:   tlsHandshakeTimeout,
		ExpectContinueTimeout: expectContinueTimeout,
		DisableKeepAlives:     true,
	}
}
