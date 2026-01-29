package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type clientBucket struct {
	Requests int
	ResetAt  time.Time
}

var (
	buckets = make(map[string]*clientBucket)
	mu      sync.RWMutex
)

func RateLimitMiddleware(maxRequests int, windowSecs int) func(http.Handler) http.Handler {
	if windowSecs < 1 {
		windowSecs = 1
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			now := time.Now()

			mu.RLock()
			bucket, exists := buckets[ip]
			mu.RUnlock()

			if !exists || now.After(bucket.ResetAt) {
				mu.Lock()
				bucket = &clientBucket{
					Requests: 0,
					ResetAt:  now.Add(time.Duration(windowSecs) * time.Second),
				}
				buckets[ip] = bucket
				mu.Unlock()
			}

			mu.Lock()
			if bucket.Requests >= maxRequests {
				mu.Unlock()
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			bucket.Requests++
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
