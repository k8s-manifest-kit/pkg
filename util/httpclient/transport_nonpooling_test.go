package httpclient_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptrace"
	"testing"
	"time"

	"github.com/k8s-manifest-kit/pkg/util/httpclient"

	. "github.com/onsi/gomega"
)

// nonPoolingRequestCount is the number of sequential requests each
// non-pooling test performs.
const nonPoolingRequestCount = 3

// serverReadHeaderTimeout bounds request header reads on test servers.
const serverReadHeaderTimeout = 5 * time.Second

func TestTransportDoesNotReuseHTTP1Connections(t *testing.T) {
	g := NewWithT(t)

	listener, err := newCountingListener(t.Context())
	g.Expect(err).NotTo(HaveOccurred())

	handler := &recordingHandler{}
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: serverReadHeaderTimeout,
	}
	go func() {
		_ = server.Serve(listener)
	}()
	t.Cleanup(func() {
		_ = server.Close()
	})

	client := httpclient.NewClient()
	t.Cleanup(client.CloseIdleConnections)
	recorder := &gotConnRecorder{}

	for range nonPoolingRequestCount {
		request, err := http.NewRequestWithContext(
			t.Context(),
			http.MethodGet,
			"http://"+listener.Addr().String()+"/manifests",
			nil,
		)
		g.Expect(err).NotTo(HaveOccurred())
		request = request.WithContext(httptrace.WithClientTrace(request.Context(), &httptrace.ClientTrace{
			GotConn: recorder.record,
		}))

		response, err := client.Do(request)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(response.Proto).To(Equal("HTTP/1.1"))

		body, err := io.ReadAll(response.Body)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(string(body)).To(Equal(servedBody))
		g.Expect(response.Body.Close()).To(Succeed())
	}

	observations := handler.snapshot()
	g.Expect(observations).To(HaveLen(nonPoolingRequestCount))
	for _, observation := range observations {
		g.Expect(observation.closeRequested).To(BeTrue(),
			"the client should ask the server to close each HTTP/1.1 connection")
	}

	conns := recorder.snapshot()
	g.Expect(conns).To(HaveLen(nonPoolingRequestCount))
	for _, conn := range conns {
		g.Expect(conn.Reused).To(BeFalse(), "no HTTP/1.1 connection should be reused")
	}

	g.Expect(listener.accepted.Load()).To(Equal(int64(nonPoolingRequestCount)),
		"every sequential HTTP/1.1 request should open its own TCP connection")
}

func TestTransportDoesNotReuseHTTP2Connections(t *testing.T) {
	g := NewWithT(t)

	certificate, rootPool, err := newTestCertificate()
	g.Expect(err).NotTo(HaveOccurred())

	listener, err := newCountingListener(t.Context())
	g.Expect(err).NotTo(HaveOccurred())

	handler := &recordingHandler{}
	server := &http.Server{
		Handler: handler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{certificate},
			NextProtos:   []string{"h2", "http/1.1"},
		},
		ReadHeaderTimeout: serverReadHeaderTimeout,
	}
	go func() {
		_ = server.ServeTLS(listener, "", "")
	}()
	t.Cleanup(func() {
		_ = server.Close()
	})

	client := httpclient.NewClient()
	t.Cleanup(client.CloseIdleConnections)
	// Trust only the test-owned CA; verification stays fully enabled.
	client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
		RootCAs: rootPool,
	}
	recorder := &gotConnRecorder{}

	for range nonPoolingRequestCount {
		request, err := http.NewRequestWithContext(
			t.Context(),
			http.MethodGet,
			"https://"+listener.Addr().String()+"/manifests",
			nil,
		)
		g.Expect(err).NotTo(HaveOccurred())
		request = request.WithContext(httptrace.WithClientTrace(request.Context(), &httptrace.ClientTrace{
			GotConn: recorder.record,
		}))

		response, err := client.Do(request)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(response.Proto).To(Equal("HTTP/2.0"))
		g.Expect(response.TLS.NegotiatedProtocol).To(Equal("h2"))

		body, err := io.ReadAll(response.Body)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(string(body)).To(Equal(servedBody))
		g.Expect(response.Body.Close()).To(Succeed())
	}

	observations := handler.snapshot()
	g.Expect(observations).To(HaveLen(nonPoolingRequestCount))

	conns := recorder.snapshot()
	g.Expect(conns).To(HaveLen(nonPoolingRequestCount))
	for _, conn := range conns {
		g.Expect(conn.Reused).To(BeFalse(), "no HTTP/2 connection should be reused")
	}

	g.Expect(listener.accepted.Load()).To(Equal(int64(nonPoolingRequestCount)),
		"every sequential HTTP/2 request should open its own TLS connection")
}

// newTestCertificate creates a throwaway CA and a server certificate for
// loopback TLS serving, returning the certificate together with a root pool
// that trusts only the test-owned CA.
func newTestCertificate() (tls.Certificate, *x509.CertPool, error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("generate CA key: %w", err)
	}

	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "httpclient test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caCertificate, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("create CA certificate: %w", err)
	}
	ca, err := x509.ParseCertificate(caCertificate)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("parse CA certificate: %w", err)
	}

	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("generate server key: %w", err)
	}
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	serverCertificate, err := x509.CreateCertificate(rand.Reader, serverTemplate, ca, &serverKey.PublicKey, caKey)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("create server certificate: %w", err)
	}

	rootPool := x509.NewCertPool()
	rootPool.AddCert(ca)

	return tls.Certificate{
		Certificate: [][]byte{serverCertificate},
		PrivateKey:  serverKey,
	}, rootPool, nil
}
