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

// NewTransport returns a new http.Transport configured with explicit
// standard library defaults and no connection pooling: keep-alives are
// disabled and no idle connection is retained per host, so every request
// uses its own connection. Each call returns a fresh instance with a fresh
// net.Dialer, and the proxy is resolved from the process environment when a
// request is made. TLS verification keeps its zero value, HTTP/2 is
// attempted when the connection supports it, and callers own the returned
// instance and its lifecycle.
func NewTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: dialTimeout, KeepAlive: dialKeepAlive}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          maxIdleConns,
		MaxIdleConnsPerHost:   noIdleConnsPerHost,
		IdleConnTimeout:       idleConnTimeout,
		TLSHandshakeTimeout:   tlsHandshakeTimeout,
		ExpectContinueTimeout: expectContinueTimeout,
		DisableKeepAlives:     true,
	}
}
