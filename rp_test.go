package reverse_proxy_talk

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
)

func TestReverseProxy(t *testing.T) {
	// Create a test http server for the proxy to call
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Make any assertions we like about the incoming request based on what we expect the proxy to do
		if r.Method != http.MethodGet {
			t.Errorf("expected method GET, got: %s", r.Method)
		}
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("hello world"))
	}))
	defer target.Close()

	// Configure the proxy and start a test server with the proxy as the handler
	u, _ := url.Parse(target.URL)
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxyServer := httptest.NewServer(proxy)
	defer proxyServer.Close()

	// We can call the proxy with our http client and make assertions about the response
	resp, err := http.DefaultClient.Get(proxyServer.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	t.Logf("response: %s", string(b))
	if resp.StatusCode != http.StatusTeapot {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}
}

func TestRewrite(t *testing.T) {
	// Set up a rewrite function to send requests to a different target
	target, _ := url.Parse("http://localhost:8081/baseurl/")
	rewrite := func(pr *httputil.ProxyRequest) {
		pr.SetURL(target)
		pr.Out.Host = target.Host
	}

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080", nil)
	pr := &httputil.ProxyRequest{
		In:  req,
		Out: req.Clone(req.Context()),
	}
	rewrite(pr)

	if pr.Out.URL.String() != target.String() {
		t.Errorf("unexpected target: %s", pr.Out.URL.String())
	}
}
