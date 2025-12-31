// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

package middleware

import (
	"net/http"
	"sync"
	"time"
)

type visitor struct {
	limiter  *rateLimiter
	lastSeen time.Time
}

type rateLimiter struct {
	rate   int
	per    time.Duration
	tokens int
	last   time.Time
	mu     sync.Mutex
}

func newRateLimiter(r int, d time.Duration) *rateLimiter {
	return &rateLimiter{
		rate:   r,
		per:    d,
		tokens: r,
		last:   time.Now(),
	}
}

func (l *rateLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	elapsed := now.Sub(l.last)
	if elapsed > l.per {
		l.tokens = l.rate
		l.last = now
	}

	if l.tokens > 0 {
		l.tokens--
		return true
	}
	return false
}

var (
	visitors = make(map[string]*rateLimiter)
	mu       sync.Mutex
)

func RateLimit(requests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			mu.Lock()
			limiter, exists := visitors[ip]
			if !exists {
				limiter = newRateLimiter(requests, window)
				visitors[ip] = limiter
			}
			mu.Unlock()

			if !limiter.Allow() {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
