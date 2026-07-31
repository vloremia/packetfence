package api

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type rateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	clients map[string]bucket
}

type bucket struct {
	started time.Time
	count   int
}

func newRateLimiter(limit int) *rateLimiter {
	if limit < 1 {
		limit = 1
	}
	return &rateLimiter{limit: limit, window: time.Minute, clients: make(map[string]bucket)}
}

func (l *rateLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	b := l.clients[key]
	if b.started.IsZero() || now.Sub(b.started) >= l.window {
		l.clients[key] = bucket{started: now, count: 1}
		return true
	}
	if b.count >= l.limit {
		return false
	}
	b.count++
	l.clients[key] = b
	return true
}

func clientKey(r *http.Request) string {
	if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); forwarded != "" {
		return forwarded
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}
