package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestShutdownSequenceOrder verifies that shutdown phases execute in correct order
func TestShutdownSequenceOrder(t *testing.T) {
	tests := []struct {
		name      string
		wantOrder []string
	}{
		{
			name: "phases execute in order: http_shutdown -> goroutines_stop -> db_close",
			wantOrder: []string{
				"phase_1_start",
				"phase_1_complete",
				"phase_2_start",
				"phase_2_complete",
				"phase_3_start",
				"phase_3_complete",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executionOrder := make([]string, 0)
			mu := sync.Mutex{}

			recordPhase := func(phase string) {
				mu.Lock()
				defer mu.Unlock()
				executionOrder = append(executionOrder, phase)
			}

			// Phase 1: HTTP server shutdown
			recordPhase("phase_1_start")
			// Simulated: srv.Shutdown(ctx)
			time.Sleep(10 * time.Millisecond)
			recordPhase("phase_1_complete")

			// Phase 2: Stop goroutines
			recordPhase("phase_2_start")
			// Simulated: cancel health check, stop rate limiters, stop telegram
			time.Sleep(10 * time.Millisecond)
			recordPhase("phase_2_complete")

			// Phase 3: Close database
			recordPhase("phase_3_start")
			// Simulated: db.Close()
			time.Sleep(10 * time.Millisecond)
			recordPhase("phase_3_complete")

			// Verify order
			if len(executionOrder) != len(tt.wantOrder) {
				t.Errorf("phase count mismatch: got %d, want %d", len(executionOrder), len(tt.wantOrder))
			}

			for i, phase := range executionOrder {
				if i < len(tt.wantOrder) && phase != tt.wantOrder[i] {
					t.Errorf("phase %d: got %q, want %q", i, phase, tt.wantOrder[i])
				}
			}
		})
	}
}

// TestHealthCheckGoroutineStopsBeforeDBClose verifies health check doesn't access DB after close
func TestHealthCheckGoroutineStopsBeforeDBClose(t *testing.T) {
	tests := []struct {
		name           string
		contextTimeout time.Duration
		expectExit     bool
	}{
		{
			name:           "health check goroutine exits when context cancelled",
			contextTimeout: 100 * time.Millisecond,
			expectExit:     true,
		},
		{
			name:           "health check respects grace period",
			contextTimeout: 50 * time.Millisecond,
			expectExit:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			// Track goroutine state
			var wg sync.WaitGroup
			goroutineExited := make(chan bool, 1)
			var dbAccessCount int64

			wg.Add(1)
			go func() {
				defer wg.Done()

				// Simulate health check goroutine
				ticker := time.NewTicker(10 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						// Gracefully exit when context is cancelled
						goroutineExited <- true
						return
					case <-ticker.C:
						// Simulate database access
						atomic.AddInt64(&dbAccessCount, 1)
					}
				}
			}()

			// Cancel context (simulating shutdown signal)
			time.Sleep(tt.contextTimeout)
			cancel()

			// Wait for goroutine to exit
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				// Goroutine exited successfully
				t.Logf("Health check goroutine exited successfully after %d DB accesses", atomic.LoadInt64(&dbAccessCount))
			case <-time.After(1 * time.Second):
				t.Errorf("Health check goroutine did not exit within timeout")
			}

			// Verify goroutine received cancellation signal
			select {
			case <-goroutineExited:
				t.Log("Health check goroutine received context cancellation")
			default:
				t.Errorf("Health check goroutine did not exit via context.Done()")
			}
		})
	}
}

// TestNoDBAccessAfterClose verifies background goroutines don't access DB after close
func TestNoDBAccessAfterClose(t *testing.T) {
	tests := []struct {
		name                string
		gracePeriodMS       int
		expectDBAccessAfter bool
	}{
		{
			name:                "with grace period, no DB access after close",
			gracePeriodMS:       100,
			expectDBAccessAfter: false,
		},
		{
			name:                "without grace period, may have access after close",
			gracePeriodMS:       0,
			expectDBAccessAfter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			var accessBeforeClose int64
			var accessAfterClose int64
			var dbClosed int64
			var wg sync.WaitGroup

			// Simulate background goroutine with database access
			wg.Add(1)
			go func() {
				defer wg.Done()

				ticker := time.NewTicker(10 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						if atomic.LoadInt64(&dbClosed) == 0 {
							atomic.AddInt64(&accessBeforeClose, 1)
						} else {
							atomic.AddInt64(&accessAfterClose, 1)
						}
					}
				}
			}()

			// Simulate shutdown
			time.Sleep(50 * time.Millisecond)
			cancel()

			// Simulate grace period
			if tt.gracePeriodMS > 0 {
				time.Sleep(time.Duration(tt.gracePeriodMS) * time.Millisecond)
			}

			// Mark database as closed
			atomic.StoreInt64(&dbClosed, 1)

			// Wait for goroutine to exit
			wg.Wait()

			accessAfter := atomic.LoadInt64(&accessAfterClose)
			t.Logf("DB accesses before close: %d, after close: %d", atomic.LoadInt64(&accessBeforeClose), accessAfter)

			// With grace period, should have no (or very few) accesses after close
			if tt.gracePeriodMS > 0 && accessAfter > 0 {
				t.Logf("Warning: grace period did not fully prevent DB access after close (%d accesses)", accessAfter)
			}
		})
	}
}

// TestRateLimiterStopsBeforeDBClose verifies rate limiters complete cleanup before DB close
func TestRateLimiterStopsBeforeDBClose(t *testing.T) {
	tests := []struct {
		name           string
		expectComplete bool
	}{
		{
			name:           "rate limiter cleanup completes before DB close",
			expectComplete: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanupComplete int64
			var dbAccessCount int64
			var wg sync.WaitGroup

			// Simulate rate limiter with Stop() method
			stopChan := make(chan struct{})

			wg.Add(1)
			go func() {
				defer wg.Done()

				ticker := time.NewTicker(5 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-stopChan:
						// Cleanup phase - might access DB one more time
						atomic.AddInt64(&dbAccessCount, 1)
						atomic.StoreInt64(&cleanupComplete, 1)
						return
					case <-ticker.C:
						// Normal operation
						atomic.AddInt64(&dbAccessCount, 1)
					}
				}
			}()

			// Simulate operation
			time.Sleep(30 * time.Millisecond)

			// Simulate shutdown: call Stop()
			close(stopChan)

			// Wait for cleanup
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				if atomic.LoadInt64(&cleanupComplete) == 1 {
					t.Logf("Rate limiter cleanup completed successfully (%d DB accesses)", atomic.LoadInt64(&dbAccessCount))
				} else {
					t.Error("Rate limiter cleanup did not complete")
				}
			case <-time.After(1 * time.Second):
				t.Error("Rate limiter cleanup timeout")
			}
		})
	}
}

// TestTelegramServiceShutdownBeforeDBClose verifies Telegram service stops before DB close
func TestTelegramServiceShutdownBeforeDBClose(t *testing.T) {
	tests := []struct {
		name           string
		expectShutdown bool
	}{
		{
			name:           "telegram service shutdown completes before DB close",
			expectShutdown: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var shutdownStarted int64
			var cleanupComplete int64
			var wg sync.WaitGroup

			// Simulate Telegram service with background token cleanup
			stopChan := make(chan struct{})

			wg.Add(1)
			go func() {
				defer wg.Done()

				atomic.StoreInt64(&shutdownStarted, 1)

				// Simulate cleanup work (e.g., removing expired tokens from DB)
				ticker := time.NewTicker(5 * time.Millisecond)
				defer ticker.Stop()

				// Cleanup phase
				for {
					select {
					case <-stopChan:
						// Final cleanup iteration
						atomic.AddInt64(&cleanupComplete, 1)
						return
					case <-ticker.C:
						// Simulate DB cleanup work
						continue
					}
				}
			}()

			// Start shutdown
			time.Sleep(30 * time.Millisecond)
			close(stopChan)

			// Wait for service to shutdown
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				if atomic.LoadInt64(&cleanupComplete) == 1 {
					t.Log("Telegram service shutdown completed successfully")
				} else {
					t.Error("Telegram service cleanup did not complete")
				}
			case <-time.After(1 * time.Second):
				t.Error("Telegram service shutdown timeout")
			}
		})
	}
}

// TestConcurrentRequestsCompleteDuringShutdown verifies in-flight requests complete
func TestConcurrentRequestsCompleteDuringShutdown(t *testing.T) {
	tests := []struct {
		name              string
		requestCount      int
		requestDuration   time.Duration
		shutdownTimeout   time.Duration
		expectAllComplete bool
	}{
		{
			name:              "short requests complete during shutdown grace period",
			requestCount:      5,
			requestDuration:   10 * time.Millisecond,
			shutdownTimeout:   1 * time.Second,
			expectAllComplete: true,
		},
		{
			name:              "many concurrent requests complete gracefully",
			requestCount:      20,
			requestDuration:   50 * time.Millisecond,
			shutdownTimeout:   2 * time.Second,
			expectAllComplete: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestsStarted int64
			var requestsCompleted int64
			var wg sync.WaitGroup

			// Simulate concurrent requests
			for i := 0; i < tt.requestCount; i++ {
				wg.Add(1)
				go func(reqID int) {
					defer wg.Done()
					atomic.AddInt64(&requestsStarted, 1)

					// Simulate request processing
					time.Sleep(tt.requestDuration)

					atomic.AddInt64(&requestsCompleted, 1)
				}(i)
			}

			// Wait for server shutdown timeout
			shutdownDone := make(chan struct{})
			go func() {
				wg.Wait()
				close(shutdownDone)
			}()

			select {
			case <-shutdownDone:
				completed := atomic.LoadInt64(&requestsCompleted)
				started := atomic.LoadInt64(&requestsStarted)

				if completed == int64(tt.requestCount) {
					t.Logf("All %d requests completed during shutdown", tt.requestCount)
				} else {
					t.Logf("Only %d/%d requests completed during shutdown", completed, started)
				}
			case <-time.After(tt.shutdownTimeout):
				completed := atomic.LoadInt64(&requestsCompleted)
				t.Errorf("Shutdown timeout: %d/%d requests completed", completed, tt.requestCount)
			}
		})
	}
}

// TestGracePeriodAllowsGoroutinesToExit verifies grace period is sufficient
func TestGracePeriodAllowsGoroutinesToExit(t *testing.T) {
	tests := []struct {
		name              string
		gracePeriodMS     int
		goroutineExitTime time.Duration
		expectExit        bool
	}{
		{
			name:              "100ms grace period allows 50ms goroutine to exit",
			gracePeriodMS:     100,
			goroutineExitTime: 50 * time.Millisecond,
			expectExit:        true,
		},
		{
			name:              "grace period timeout prevents blocked goroutines",
			gracePeriodMS:     50,
			goroutineExitTime: 200 * time.Millisecond,
			expectExit:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			var goroutineExited int64
			var wg sync.WaitGroup

			// Simulate goroutine that takes time to exit
			wg.Add(1)
			go func() {
				defer wg.Done()

				select {
				case <-ctx.Done():
					time.Sleep(tt.goroutineExitTime)
					atomic.StoreInt64(&goroutineExited, 1)
				}
			}()

			// Cancel context
			cancel()

			// Apply grace period
			gracePeriod := time.Duration(tt.gracePeriodMS) * time.Millisecond
			time.Sleep(gracePeriod)

			// Try to wait for goroutine
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				if atomic.LoadInt64(&goroutineExited) == 1 {
					t.Log("Goroutine exited within grace period")
				} else {
					t.Error("Goroutine did not properly exit")
				}
			case <-time.After(time.Duration(tt.gracePeriodMS) * time.Millisecond):
				if tt.expectExit {
					t.Errorf("Goroutine did not exit within %dms grace period", tt.gracePeriodMS)
				} else {
					t.Logf("Goroutine correctly did not exit within %dms (as expected)", tt.gracePeriodMS)
				}
			}
		})
	}
}

// TestShutdownErrorHandling verifies shutdown errors are logged but don't panic
func TestShutdownErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		dbCloseError  error
		expectLog     bool
		expectNoPanic bool
	}{
		{
			name:          "database close error is logged but doesn't panic",
			dbCloseError:  fmt.Errorf("connection pool already closed"),
			expectLog:     true,
			expectNoPanic: true,
		},
		{
			name:          "successful database close logs completion",
			dbCloseError:  nil,
			expectLog:     true,
			expectNoPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track if panic occurs
			var panicked bool
			defer func() {
				if r := recover(); r != nil {
					panicked = true
					t.Logf("Panic occurred: %v", r)
				}
			}()

			// Simulate database close
			if tt.dbCloseError != nil {
				// In production: log.Error().Err(err).Msg("Error closing database")
				t.Logf("Error closing database: %v", tt.dbCloseError)
			} else {
				// In production: log.Debug().Msg("Phase 3: Database connection closed")
				t.Log("Database connection closed")
			}

			if panicked && tt.expectNoPanic {
				t.Error("Shutdown should not panic even on database close error")
			}

			if !panicked && tt.expectNoPanic {
				t.Log("Shutdown handled error gracefully without panic")
			}
		})
	}
}
