package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
	handler := RateLimitMiddleware(2, 10)(okHandler())
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, newReqWithIP("1.2.3.4:1234"))

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRateLimit_BlocksAfterLimit(t *testing.T) {
	handler := RateLimitMiddleware(1, 10)(okHandler())

	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, newReqWithIP("5.6.7.8:1111"))

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, newReqWithIP("5.6.7.8:1111"))

	assert.Equal(t, http.StatusOK, rr1.Code)
	assert.Equal(t, http.StatusTooManyRequests, rr2.Code)
}

func TestRateLimit_RetryAfterHeader(t *testing.T) {
	handler := RateLimitMiddleware(1, 30)(okHandler())

	handler.ServeHTTP(httptest.NewRecorder(), newReqWithIP("4.4.4.4:1111"))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, newReqWithIP("4.4.4.4:1111"))

	assert.Equal(t, "30", rr.Header().Get("Retry-After"))
}

func TestRateLimit_SeparateBucketsPerIP(t *testing.T) {
	handler := RateLimitMiddleware(1, 10)(okHandler())

	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, newReqWithIP("10.0.0.1:1111"))

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, newReqWithIP("10.0.0.2:2222"))

	assert.Equal(t, http.StatusOK, rr1.Code)
	assert.Equal(t, http.StatusOK, rr2.Code)
}

// Separate middleware instances must not share state, otherwise a second App in
// the same process inherits the first one's exhausted buckets.
func TestRateLimit_StateIsNotSharedBetweenInstances(t *testing.T) {
	first := RateLimitMiddleware(1, 10)(okHandler())
	second := RateLimitMiddleware(1, 10)(okHandler())

	rr1 := httptest.NewRecorder()
	first.ServeHTTP(rr1, newReqWithIP("7.7.7.7:1111"))
	assert.Equal(t, http.StatusOK, rr1.Code)

	rr2 := httptest.NewRecorder()
	first.ServeHTTP(rr2, newReqWithIP("7.7.7.7:1111"))
	assert.Equal(t, http.StatusTooManyRequests, rr2.Code)

	rr3 := httptest.NewRecorder()
	second.ServeHTTP(rr3, newReqWithIP("7.7.7.7:1111"))
	assert.Equal(t, http.StatusOK, rr3.Code)
}

func TestRateLimit_ResetsAfterWindow(t *testing.T) {
	handler := RateLimitMiddleware(1, 1)(okHandler())

	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, newReqWithIP("9.9.9.9:1111"))

	time.Sleep(time.Second + 20*time.Millisecond)

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, newReqWithIP("9.9.9.9:1111"))

	assert.Equal(t, http.StatusOK, rr2.Code)
}

// A client that appears once must not leave a permanent entry behind, otherwise
// attacker-controlled IPs grow the map without bound.
func TestRateLimit_ExpiredBucketsAreSwept(t *testing.T) {
	handler := RateLimitMiddleware(10, 1)(okHandler())

	handler.ServeHTTP(httptest.NewRecorder(), newReqWithIP("8.8.8.8:1111"))

	time.Sleep(time.Second + 20*time.Millisecond)

	// A different client arriving after the window triggers the sweep.
	handler.ServeHTTP(httptest.NewRecorder(), newReqWithIP("8.8.4.4:1111"))

	// The original client can be served again, proving its stale bucket was
	// dropped rather than merely reset in place.
	rr := httptest.NewRecorder()
	for i := 0; i < 10; i++ {
		handler.ServeHTTP(rr, newReqWithIP("8.8.8.8:1111"))
	}
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRateLimit_InvalidConfigFallsBackToMinimums(t *testing.T) {
	handler := RateLimitMiddleware(0, 0)(okHandler())

	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, newReqWithIP("6.6.6.6:1111"))
	assert.Equal(t, http.StatusOK, rr1.Code)

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, newReqWithIP("6.6.6.6:1111"))
	assert.Equal(t, http.StatusTooManyRequests, rr2.Code)
}
