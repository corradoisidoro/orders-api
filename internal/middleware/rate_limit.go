package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// clientBucket tracks request usage for a single client within a window.
type clientBucket struct {
	Requests int
	ResetAt  time.Time
}

// RateLimitMiddleware limits each client to maxRequests within windowSecs.
//
// Buckets are per-middleware-instance and in-memory, so the effective limit is
// per process. Expired buckets are swept whenever a new client appears, which
// bounds memory to the set of clients seen within roughly one window.
func RateLimitMiddleware(maxRequests int, windowSecs int) func(http.Handler) http.Handler {
	if maxRequests < 1 {
		maxRequests = 1
	}
	if windowSecs < 1 {
		windowSecs = 1
	}

	var (
		mu      sync.Mutex
		buckets = make(map[string]*clientBucket)
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)

			now := time.Now()
			window := time.Duration(windowSecs) * time.Second

			mu.Lock()
			// Sweep expired buckets so the map cannot grow without bound when
			// clients only ever appear once.
			for key, b := range buckets {
				if now.After(b.ResetAt) {
					delete(buckets, key)
				}
			}

			bucket, ok := buckets[ip]
			if !ok {
				bucket = &clientBucket{ResetAt: now.Add(window)}
				buckets[ip] = bucket
			}
			limited := bucket.Requests >= maxRequests
			if !limited {
				bucket.Requests++
			}
			mu.Unlock()

			if limited {
				w.Header().Set("Retry-After", strconv.Itoa(windowSecs))
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
