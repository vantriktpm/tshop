// Package middleware provides HTTP/gRPC middleware (rate limit, auth).
package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter in-memory rate limiter (production: use Redis).
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // requests per window
	window   time.Duration
}

type visitor struct {
	count    int
	lastSeen time.Time
}

// NewRateLimiter creates a rate limiter: rate requests per window per key.
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	r := &RateLimiter{visitors: make(map[string]*visitor), rate: rate, window: window}
	go r.cleanup()
	return r
}

func (r *RateLimiter) cleanup() {
	for range time.Tick(time.Minute) {
		r.mu.Lock()
		for k, v := range r.visitors {
			if time.Since(v.lastSeen) > r.window {
				delete(r.visitors, k)
			}
		}
		r.mu.Unlock()
	}
}

// Allow returns true if the key (e.g. IP or user ID) is within rate limit.
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	v, ok := r.visitors[key]
	if !ok {
		r.visitors[key] = &visitor{count: 1, lastSeen: now}
		return true
	}
	if now.Sub(v.lastSeen) > r.window {
		v.count = 1
		v.lastSeen = now
		return true
	}
	v.count++
	v.lastSeen = now
	return v.count <= r.rate
}

// KeyFunc extracts rate-limit key from request (e.g. IP). Used by Gin handler.
func (r *RateLimiter) KeyFromRequest(req *http.Request) string {
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	return req.RemoteAddr
}
