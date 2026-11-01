package httpclient_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/k8s-manifest-kit/pkg/util/httpclient"

	. "github.com/onsi/gomega"
)

func TestNewClientReturnsIndependentClients(t *testing.T) {
	g := NewWithT(t)

	first := httpclient.NewClient()
	second := httpclient.NewClient()

	g.Expect(first).NotTo(BeIdenticalTo(second))
	g.Expect(first).NotTo(BeIdenticalTo(http.DefaultClient))
	g.Expect(first.Transport).NotTo(BeIdenticalTo(second.Transport))
	g.Expect(first.Transport).NotTo(BeIdenticalTo(http.DefaultTransport))
	g.Expect(first.Timeout).To(Equal(time.Duration(0)))
}

func TestNewTransportReturnsNonPooledTransports(t *testing.T) {
	g := NewWithT(t)

	first := httpclient.NewTransport()
	second := httpclient.NewTransport()

	g.Expect(first).NotTo(BeIdenticalTo(second))
	g.Expect(first).NotTo(BeIdenticalTo(http.DefaultTransport))
	g.Expect(first.Proxy).NotTo(BeNil())
	g.Expect(first.DisableKeepAlives).To(BeTrue())
	g.Expect(first.MaxIdleConnsPerHost).To(Equal(-1))
}
