package httpclient_test

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/k8s-manifest-kit/pkg/util/httpclient"

	. "github.com/onsi/gomega"
)

func TestNewClientReturnsIndependentClients(t *testing.T) {
	g := NewWithT(t)

	first := httpclient.NewClient()
	second := httpclient.NewClient()
	t.Cleanup(first.CloseIdleConnections)
	t.Cleanup(second.CloseIdleConnections)

	g.Expect(first).NotTo(BeNil())
	g.Expect(second).NotTo(BeNil())
	g.Expect(first).NotTo(BeIdenticalTo(second))
	g.Expect(first).NotTo(BeIdenticalTo(http.DefaultClient))
	g.Expect(first.Transport).NotTo(BeNil())
	g.Expect(second.Transport).NotTo(BeNil())
	g.Expect(first.Transport).NotTo(BeIdenticalTo(second.Transport))
	g.Expect(first.Transport).NotTo(BeIdenticalTo(http.DefaultTransport))
}

func TestNewTransportReturnsIndependentTransports(t *testing.T) {
	g := NewWithT(t)

	first := httpclient.NewTransport()
	second := httpclient.NewTransport()
	t.Cleanup(first.CloseIdleConnections)
	t.Cleanup(second.CloseIdleConnections)

	g.Expect(first).NotTo(BeNil())
	g.Expect(second).NotTo(BeNil())
	g.Expect(first).NotTo(BeIdenticalTo(second))
	g.Expect(first).NotTo(BeIdenticalTo(http.DefaultTransport))
}

func TestNewClientDefaults(t *testing.T) {
	g := NewWithT(t)

	client := httpclient.NewClient()
	t.Cleanup(client.CloseIdleConnections)

	g.Expect(client.Timeout).To(Equal(time.Duration(0)))
	g.Expect(client.CheckRedirect).To(BeNil())
	g.Expect(client.Jar).To(BeNil())

	transport := client.Transport.(*http.Transport)

	g.Expect(transport.Proxy).NotTo(BeNil())
	g.Expect(transport.DialContext).NotTo(BeNil())
	g.Expect(transport.DisableKeepAlives).To(BeTrue())
	g.Expect(transport.ForceAttemptHTTP2).To(BeTrue())
	g.Expect(transport.MaxIdleConns).To(Equal(100))
	g.Expect(transport.MaxIdleConnsPerHost).To(Equal(-1))
	g.Expect(transport.IdleConnTimeout).To(Equal(90 * time.Second))
	g.Expect(transport.TLSHandshakeTimeout).To(Equal(10 * time.Second))
	g.Expect(transport.ExpectContinueTimeout).To(Equal(1 * time.Second))
	g.Expect(transport.TLSClientConfig).To(BeNil())
	g.Expect(transport.TLSNextProto).To(BeNil())
	g.Expect(transport.ResponseHeaderTimeout).To(Equal(time.Duration(0)))
	g.Expect(transport.MaxConnsPerHost).To(Equal(0))
}

func TestMutatingOneInstanceLeavesOtherInstancesUnchanged(t *testing.T) {
	g := NewWithT(t)

	first := httpclient.NewClient()
	second := httpclient.NewClient()
	t.Cleanup(first.CloseIdleConnections)
	t.Cleanup(second.CloseIdleConnections)
	firstTransport := first.Transport.(*http.Transport)
	secondTransport := second.Transport.(*http.Transport)
	third := httpclient.NewTransport()
	t.Cleanup(third.CloseIdleConnections)

	first.Timeout = time.Minute
	first.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	firstTransport.DisableKeepAlives = false
	firstTransport.MaxIdleConns = 0
	firstTransport.MaxIdleConnsPerHost = 100
	firstTransport.TLSHandshakeTimeout = time.Minute
	firstTransport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS13}

	third.DisableKeepAlives = false

	g.Expect(second.Timeout).To(Equal(time.Duration(0)))
	g.Expect(second.CheckRedirect).To(BeNil())
	g.Expect(secondTransport.DisableKeepAlives).To(BeTrue())
	g.Expect(secondTransport.MaxIdleConns).To(Equal(100))
	g.Expect(secondTransport.MaxIdleConnsPerHost).To(Equal(-1))
	g.Expect(secondTransport.TLSHandshakeTimeout).To(Equal(10 * time.Second))
	g.Expect(secondTransport.TLSClientConfig).To(BeNil())

	fourth := httpclient.NewTransport()
	t.Cleanup(fourth.CloseIdleConnections)
	g.Expect(fourth.DisableKeepAlives).To(BeTrue())
	g.Expect(fourth.MaxIdleConnsPerHost).To(Equal(-1))

	// Construction must never touch the standard library globals.
	g.Expect(http.DefaultTransport.(*http.Transport).DisableKeepAlives).To(BeFalse())
	g.Expect(http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost).To(Equal(0))
	g.Expect(http.DefaultClient.Timeout).To(Equal(time.Duration(0)))
}

const proxyTestSubprocessEnv = "PKG_UTIL_HTTPCLIENT_PROXY_TEST_CHILD"

func TestTransportResolvesProxyFromEnvironment(t *testing.T) {
	g := NewWithT(t)
	if os.Getenv(proxyTestSubprocessEnv) != "1" {
		executable, err := os.Executable()
		g.Expect(err).NotTo(HaveOccurred())

		//nolint:gosec // Re-exec this test binary with a fixed test selector.
		command := exec.CommandContext(
			t.Context(),
			executable,
			"-test.run=^TestTransportResolvesProxyFromEnvironment$",
		)
		command.Env = proxyTestEnvironment()
		output, err := command.CombinedOutput()
		g.Expect(err).NotTo(HaveOccurred(), "proxy subprocess output:\n%s", output)

		return
	}

	transport := httpclient.NewTransport()
	g.Expect(transport.Proxy).NotTo(BeNil())

	request := &http.Request{
		URL: &url.URL{Scheme: "http", Host: "origin.example.test", Path: "/manifests"},
	}
	proxyURL, err := transport.Proxy(request)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(proxyURL).NotTo(BeNil())
	g.Expect(proxyURL.String()).To(Equal("http://proxy.example.test:3128"))
}

func proxyTestEnvironment() []string {
	env := make([]string, 0, len(os.Environ())+5)
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		switch name {
		case "HTTP_PROXY", "http_proxy", "NO_PROXY", "no_proxy", "REQUEST_METHOD":
			continue
		default:
			env = append(env, entry)
		}
	}

	return append(env,
		proxyTestSubprocessEnv+"=1",
		"HTTP_PROXY=http://proxy.example.test:3128",
		"http_proxy=http://proxy.example.test:3128",
		"NO_PROXY=",
		"no_proxy=",
	)
}
