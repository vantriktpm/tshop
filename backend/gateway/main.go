package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
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

	wsHub := newHub()
	go wsHub.run()

	// Redis: subscribe avatar.saved and push to WebSocket clients
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		rdb := redis.NewClient(&redis.Options{Addr: addr})
		if err := rdb.Ping(context.Background()).Err(); err != nil {
			log.Printf("gateway: redis ping: %v (ws push disabled)", err)
		} else {
			go runRedisSub(rdb, wsHub)
		}
	}

	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "*"
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.URL.Path == "/ws" {
			wsHub.handleWS(w, r)
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
		port = "5000"
	}
	log.Printf("gateway: listening on :%s (CORS: %s)", port, corsOrigin)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
