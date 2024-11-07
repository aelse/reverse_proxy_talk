package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type RProxy struct {
	proxy     *httputil.ReverseProxy
	targetURL url.URL
}

func (rp *RProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/proxy_target" {
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte(rp.targetURL.String() + "\n"))
		case http.MethodPost:
			target := r.URL.Query().Get("target")
			u, err := r.URL.Parse(target)
			if target == "" || err != nil {
				http.Error(w, "missing or bad target", http.StatusBadRequest)
				return
			}
			rp.proxy = httputil.NewSingleHostReverseProxy(u)
			rp.targetURL = *u
			fmt.Fprintf(w, "proxy target set to %s\n", target)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	// Proxy the request, but update the Host header to match the target
	r.Host = rp.targetURL.Host
	rp.proxy.ServeHTTP(w, r)
}

func main() {
	proxy := &RProxy{}
	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", proxy)
}
