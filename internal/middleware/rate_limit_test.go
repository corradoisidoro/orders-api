package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func resetBuckets(t *testing.T) {
	t.Helper()
	mu.Lock()
	buckets = make(map[string]*clientBucket)
	mu.Unlock()
}

func newReqWithIP(ip string) *http.Request {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = ip
	return req
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRateLimit_AllowsFirstRequest(t *testing.T) {
	resetBuckets(t)

	handler := RateLimitMiddleware(2, 10)(okHandler())
	rr := httptest.NewRecorder()
	req := newReqWithIP("1.2.3.4:1234")

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRateLimit_BlocksAfterLimit(t *testing.T) {
	resetBuckets(t)

	handler := RateLimitMiddleware(1, 10)(okHandler())

	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, newReqWithIP("5.6.7.8:1111"))

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, newReqWithIP("5.6.7.8:1111"))

	assert.Equal(t, http.StatusTooManyRequests, rr2.Code)
}

func TestRateLimit_SeparateBucketsPerIP(t *testing.T) {
	resetBuckets(t)

	handler := RateLimitMiddleware(1, 10)(okHandler())

	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, newReqWithIP("10.0.0.1:1111"))

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, newReqWithIP("10.0.0.2:2222"))

	assert.Equal(t, http.StatusOK, rr1.Code)
	assert.Equal(t, http.StatusOK, rr2.Code)
}

func TestRateLimit_ResetsAfterWindow(t *testing.T) {
	resetBuckets(t)

	handler := RateLimitMiddleware(1, 1)(okHandler())

	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, newReqWithIP("9.9.9.9:1111"))

	time.Sleep(time.Second + 20*time.Millisecond)

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, newReqWithIP("9.9.9.9:1111"))

	assert.Equal(t, http.StatusOK, rr2.Code)
}
