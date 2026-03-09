// Package gateway: router chung chứa tất cả service URL + CORS.
package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// ServiceBackend mapping path prefix -> backend URL (service đích).
// Hostname matches the Docker Compose service name; port matches the internal
// port each service listens on (also exposed on the host at the same number).
var ServiceBackend = map[string]string{
	"/api/auth":         "http://user-service:5001",
	"/api/orders":       "http://order-service:5002",
	"/api/products":     "http://product-service:5003",
	"/api/inventory":    "http://inventory-service:5004",
	"/api/cart":         "http://cart-service:5005",
	"/api/payment":      "http://payment-service:5006",
	"/api/shipping":     "http://shipping-service:5007",
	"/api/promotion":    "http://promotion-service:5008",
	"/api/notification": "http://notification-service:5009",
	"/api/images":       "http://image-service:5010",
}

func newReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
	}
	return proxy
}

func proxyHandlerFor(backendURL string) (http.Handler, error) {
	u, err := url.Parse(backendURL)
	if err != nil {
		return nil, err
	}
	return newReverseProxy(u), nil
}

// longestPrefixMatch trả về backend URL cho path (ưu tiên prefix dài nhất).
func longestPrefixMatch(path string) string {
	var bestPrefix string
	var bestBackend string
	for prefix, backend := range ServiceBackend {
		if strings.HasPrefix(path, prefix) && len(prefix) > len(bestPrefix) {
			bestPrefix = prefix
			bestBackend = backend
		}
	}
	return bestBackend
}
