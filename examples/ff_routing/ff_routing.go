package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			callingService := pr.In.Header.Get("Calling-Service")
			if shouldApplyNewRouteLogic(callingService) {
				log.Printf("Applying new route logic for service %s", callingService)
				agg := selectAggregator(callingService)
				pr.Out.URL.Scheme = agg.URL.Scheme
				pr.Out.URL.Host = agg.URL.Host
				pr.Out.Host = agg.URL.Host
				return
			}
			log.Printf("Applying old route logic for service %s", callingService)
			pr.Out.URL.Scheme = "http"
			pr.Out.URL.Host = "localhost:8081"
			pr.Out.Host = "localhost:8081"
		},
	}
	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", proxy)
}

func shouldApplyNewRouteLogic(callingService string) bool {
	// This is where you would consult your feature flag.
	// Lets just say any service beginning with letter 'a' should use the new route logic.
	return strings.HasPrefix(callingService, "a")
}

type Aggregator struct {
	URL *url.URL
}

var aggregators = []*Aggregator{
	{URL: &url.URL{Scheme: "https", Host: "aggregator1.example.com"}},
	{URL: &url.URL{Scheme: "https", Host: "aggregator2.example.com"}},
}

func selectAggregator(callingService string) *Aggregator {
	idx := len(callingService) % len(aggregators)
	return aggregators[idx]
}
