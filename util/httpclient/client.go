// Package httpclient provides constructors for private, non-pooled HTTP
// clients and transports built from explicit standard library defaults.
package httpclient

import "net/http"

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
