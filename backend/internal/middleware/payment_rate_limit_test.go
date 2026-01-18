package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
	"tutoring-platform/internal/models"
)

// TestPaymentRateLimiter_BasicFunctionality tests basic rate limiter operations
func TestPaymentRateLimiter_BasicFunctionality(t *testing.T) {
	limiter := PaymentRateLimiter()
	defer limiter.Stop()

	userID := uuid.New().String()

	// Test 10 requests succeed, 11th fails
	for i := 0; i < 10; i++ {
		l := limiter.GetLimiter(userID)
		if !l.Allow() {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 11th request should fail
	l := limiter.GetLimiter(userID)
	if l.Allow() {
		t.Error("11th request should be denied (rate limit exceeded)")
	}
}

// TestPaymentRateLimiter_PerUserLimiting tests that each user has separate limits
func TestPaymentRateLimiter_PerUserLimiting(t *testing.T) {
	limiter := PaymentRateLimiter()
	defer limiter.Stop()

	user1 := uuid.New().String()
	user2 := uuid.New().String()

	// Exhaust user1's limit
	limiter1 := limiter.GetLimiter(user1)
	for i := 0; i < 10; i++ {
		if !limiter1.Allow() {
			t.Fatalf("Failed to use user1's quota at request %d", i+1)
		}
	}

	// User1 should be rate limited
	if limiter1.Allow() {
		t.Error("User1 should be rate limited after 10 requests")
	}

	// User2 should still have quota
	limiter2 := limiter.GetLimiter(user2)
	for i := 0; i < 10; i++ {
		if !limiter2.Allow() {
			t.Errorf("User2 should not be rate limited at request %d", i+1)
		}
	}

	// User2 should also be rate limited now
	if limiter2.Allow() {
		t.Error("User2 should be rate limited after 10 requests")
	}
}

// TestPaymentRateLimiter_Cleanup tests that cleanup removes expired entries
func TestPaymentRateLimiter_Cleanup(t *testing.T) {
	// Create limiter with short TTL for testing
	limiter := NewUserRateLimiter(rate.Every(6*time.Second), 10)
	// Override TTL to 10ms for testing
	limiter.ttl = 10 * time.Millisecond
	defer limiter.Stop()

	userID := uuid.New().String()

	// Access a user to create an entry
	limiter.GetLimiter(userID)

	// Verify entry exists
	limiter.mu.RLock()
	initialCount := len(limiter.users)
	limiter.mu.RUnlock()

	if initialCount != 1 {
		t.Errorf("Expected 1 user entry, got %d", initialCount)
	}

	// Wait for entry to expire
	time.Sleep(50 * time.Millisecond)

	// Run cleanup
	limiter.CleanupExpired()

	// Verify entry was removed
	limiter.mu.RLock()
	finalCount := len(limiter.users)
	limiter.mu.RUnlock()

	if finalCount != 0 {
		t.Errorf("Expected 0 user entries after cleanup, got %d", finalCount)
	}
}

// TestPaymentRateLimiter_Middleware tests UserRateLimitMiddleware
func TestPaymentRateLimiter_Middleware(t *testing.T) {
	limiter := PaymentRateLimiter()
	defer limiter.Stop()

	userID := uuid.New()
	user := &models.User{ID: userID}

	middleware := UserRateLimitMiddleware(limiter)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test first 10 requests succeed
	for i := 0; i < 10; i++ {
		ctx := context.WithValue(context.Background(), UserContextKey, user)
		req := httptest.NewRequest("POST", "/api/v1/payments/create", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected %d, got %d",
				i+1, http.StatusOK, w.Code)
		}
	}

	// Test 11th request is rate limited
	ctx := context.WithValue(context.Background(), UserContextKey, user)
	req := httptest.NewRequest("POST", "/api/v1/payments/create", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("11th request: expected %d, got %d",
			http.StatusTooManyRequests, w.Code)
	}
}

// TestPaymentRateLimiter_Middleware_NoAuth tests that unauthenticated requests are rejected
func TestPaymentRateLimiter_Middleware_NoAuth(t *testing.T) {
	limiter := PaymentRateLimiter()
	defer limiter.Stop()

	middleware := UserRateLimitMiddleware(limiter)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request without user in context
	req := httptest.NewRequest("POST", "/api/v1/payments/create", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Unauthenticated request: expected %d, got %d",
			http.StatusUnauthorized, w.Code)
	}
}

// BenchmarkPaymentRateLimiter_Allow benchmarks the Allow operation
func BenchmarkPaymentRateLimiter_Allow(b *testing.B) {
	limiter := PaymentRateLimiter()
	defer limiter.Stop()

	userID := uuid.New().String()
	l := limiter.GetLimiter(userID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Allow()
	}
}

// BenchmarkPaymentRateLimiter_GetLimiter benchmarks GetLimiter operation
func BenchmarkPaymentRateLimiter_GetLimiter(b *testing.B) {
	limiter := PaymentRateLimiter()
	defer limiter.Stop()

	userID := uuid.New().String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.GetLimiter(userID)
	}
}

// TestPaymentRateLimiter_MultipleUsers tests concurrent access from multiple users
func TestPaymentRateLimiter_MultipleUsers(t *testing.T) {
	limiter := PaymentRateLimiter()
	defer limiter.Stop()

	// Create 3 concurrent users
	users := []string{
		uuid.New().String(),
		uuid.New().String(),
		uuid.New().String(),
	}

	// Test that each user can make 10 requests
	for _, userID := range users {
		userLimiter := limiter.GetLimiter(userID)
		for i := 0; i < 10; i++ {
			if !userLimiter.Allow() {
				t.Errorf("User %s: request %d should be allowed", userID, i+1)
			}
		}
		// 11th should fail
		if userLimiter.Allow() {
			t.Errorf("User %s: 11th request should be denied", userID)
		}
	}
}

// TestPaymentRateLimiter_StopGoroutine verifies cleanup goroutine stops properly
func TestPaymentRateLimiter_StopGoroutine(t *testing.T) {
	limiter := PaymentRateLimiter()

	// Give goroutine a moment to start
	time.Sleep(10 * time.Millisecond)

	// Stop should not panic
	limiter.Stop()

	// Closing stopChan again should be safe (it's read-only in cleanup goroutine)
	time.Sleep(10 * time.Millisecond)
}

// TestUserRateLimiter_Isolation tests that different limiters don't interfere
func TestUserRateLimiter_Isolation(t *testing.T) {
	limiter1 := PaymentRateLimiter()
	defer limiter1.Stop()

	limiter2 := PaymentRateLimiter()
	defer limiter2.Stop()

	userID := uuid.New().String()

	// Exhaust quota in limiter1
	l1 := limiter1.GetLimiter(userID)
	for i := 0; i < 10; i++ {
		if !l1.Allow() {
			t.Fatalf("limiter1: request %d should be allowed", i+1)
		}
	}

	// limiter2 should have independent quota
	l2 := limiter2.GetLimiter(userID)
	for i := 0; i < 10; i++ {
		if !l2.Allow() {
			t.Errorf("limiter2: request %d should be allowed", i+1)
		}
	}
}
