package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

// This example demonstrates a way to respond to requests made to ReverseProxy
// with a custom response without forwarding the request to the target server.
// Generally it is probably better to handle requests you don't want to proxy at all
// in the ServeHTTP method of a custom http.Handler and not call the proxy.ServeHTTP method.
// It is possible achieve the same effect by hooking into the Transport.

func main() {
	target := &url.URL{Scheme: "http", Host: "localhost:8081"}
	proxy := newReverseProxy(target, []string{"/forbidden", "/secret"})
	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", proxy)
}

func newReverseProxy(target *url.URL, forbiddenPaths []string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite:   intercept(target, forbiddenPaths),
		Transport: transportDenyForbidden(http.DefaultTransport),
	}
}

func intercept(forwardURL *url.URL, forbiddenPaths []string) func(*httputil.ProxyRequest) {
	isForbiddenRequest := func(req *http.Request) bool {
		return containsPrefix(req.URL.Path, forbiddenPaths)
	}
	return func(pr *httputil.ProxyRequest) {
		deleteInterceptHeaders(pr.Out)
		if isForbiddenRequest(pr.In) {
			setInterceptHeaders(pr.Out, http.StatusForbidden, "Forbidden path\n")
			return
		}
		pr.Out.URL.Scheme = forwardURL.Scheme
		pr.Out.URL.Host = forwardURL.Host
		pr.Out.Host = forwardURL.Host
	}
}

func containsPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// Adaptor mirroring the http.HandlerFunc signature
type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func transportDenyForbidden(rt http.RoundTripper) http.RoundTripper {
	return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		resp := getInterceptResponse(req)
		if resp != nil {
			return resp, nil
		}
		return rt.RoundTrip(req)
	})
}

func setInterceptHeaders(r *http.Request, statusCode int, statusText string) {
	r.Header.Set("interceptor-status-code", fmt.Sprint(statusCode))
	r.Header.Set("interceptor-status-text", statusText)
}

func getInterceptResponse(r *http.Request) *http.Response {
	if r.Header.Get("interceptor-status-code") == "" {
		return nil
	}
	code, _ := strconv.Atoi(r.Header.Get("interceptor-status-code"))
	text := r.Header.Get("interceptor-status-text")
	return &http.Response{
		StatusCode: code,
		Body:       io.NopCloser(strings.NewReader(text)),
	}
}

func deleteInterceptHeaders(r *http.Request) {
	r.Header.Del("interceptor-status-code")
	r.Header.Del("interceptor-status-text")
}
