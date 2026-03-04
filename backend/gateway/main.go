package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	// Cache proxy handlers per backend
	proxyCache := make(map[string]http.Handler)
	for _, backend := range ServiceBackend {
		if _, ok := proxyCache[backend]; ok {
			continue
		}
		h, err := proxyHandlerFor(backend)
		if err != nil {
			log.Printf("gateway: invalid backend %q: %v", backend, err)
			continue
		}
		proxyCache[backend] = h
	}

	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "*"
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS headers cho mọi response
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		path := r.URL.Path
		backend := longestPrefixMatch(path)
		if backend == "" {
			http.NotFound(w, r)
			return
		}
		h := proxyCache[backend]
		if h == nil {
			http.Error(w, "gateway: no proxy for backend", http.StatusBadGateway)
			return
		}
		h.ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("gateway: listening on :%s (CORS: %s)", port, corsOrigin)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
