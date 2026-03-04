// Package gateway: router chung chứa tất cả service URL + CORS.
package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// ServiceBackend mapping path prefix -> backend URL (service đích).
var ServiceBackend = map[string]string{
	"/api/auth":         "http://localhost:8080",
	"/api/orders":       "http://localhost:8081",
	"/api/products":     "http://localhost:8082",
	"/api/inventory":    "http://localhost:8083",
	"/api/cart":         "http://localhost:8084",
	"/api/payment":      "http://localhost:8085",
	"/api/shipping":     "http://localhost:8086",
	"/api/promotion":    "http://localhost:8087",
	"/api/notification": "http://localhost:8088",
	"/api/images":       "http://localhost:8089", // chạy image-service :8089 (tránh trùng payment :8085)
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
