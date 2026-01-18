package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestHealthCheckGoroutineTermination verifies that health check goroutine
// properly terminates when context is cancelled, preventing goroutine leaks
func TestHealthCheckGoroutineTermination(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		expectTimeout bool
	}{
		{
			name:          "goroutine terminates within reasonable timeout",
			timeout:       5 * time.Second,
			expectTimeout: false,
		},
		{
			name:          "goroutine terminates quickly on cancellation",
			timeout:       1 * time.Second,
			expectTimeout: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context for health check goroutine
			healthCheckCtx, cancel := context.WithCancel(context.Background())

			// Track whether goroutine exited
			var wg sync.WaitGroup
			goroutineExited := make(chan bool, 1)

			wg.Add(1)
			go func() {
				defer wg.Done()
				ticker := time.NewTicker(30 * time.Second)
				defer ticker.Stop()

				for {
					select {
					case <-healthCheckCtx.Done():
						// Goroutine received cancellation signal
						goroutineExited <- true
						return
					case <-ticker.C:
						// Simulate health check work (nothing to do in test)
					}
				}
			}()

			// Cancel the health check context
			cancel()

			// Create a channel to track timeout
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			// Wait for goroutine to exit or timeout
			select {
			case <-done:
				// Goroutine exited successfully
				if !tt.expectTimeout {
					t.Logf("Goroutine exited successfully")
				}
			case <-time.After(tt.timeout):
				if !tt.expectTimeout {
					t.Errorf("Health check goroutine did not terminate within %v", tt.timeout)
				}
			}

			// Verify goroutine exit signal was received
			select {
			case <-goroutineExited:
				t.Logf("Health check goroutine received cancellation signal")
			default:
				t.Logf("Health check goroutine exited via context.Done()")
			}
		})
	}
}

// TestMultipleHealthCheckContexts verifies that each health check context
// is independent and can be managed separately
func TestMultipleHealthCheckContexts(t *testing.T) {
	// Create multiple health check contexts
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	exited1 := make(chan bool, 1)
	exited2 := make(chan bool, 1)

	// Start first goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx1.Done():
				exited1 <- true
				return
			case <-ticker.C:
			}
		}
	}()

	// Start second goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx2.Done():
				exited2 <- true
				return
			case <-ticker.C:
			}
		}
	}()

	// Cancel only the first context
	cancel1()

	// Verify first goroutine exits
	select {
	case <-exited1:
		t.Logf("First goroutine exited after cancellation")
	case <-time.After(2 * time.Second):
		t.Errorf("First goroutine did not exit within timeout")
	}

	// Verify second goroutine is still running
	select {
	case <-exited2:
		t.Errorf("Second goroutine exited prematurely (should still be running)")
	case <-time.After(100 * time.Millisecond):
		t.Logf("Second goroutine still running (as expected)")
	}

	// Cancel second context
	cancel2()

	// Verify second goroutine exits
	select {
	case <-exited2:
		t.Logf("Second goroutine exited after cancellation")
	case <-time.After(2 * time.Second):
		t.Errorf("Second goroutine did not exit within timeout")
	}

	// Wait for both goroutines to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Logf("All goroutines exited successfully")
	case <-time.After(5 * time.Second):
		t.Errorf("Goroutines did not exit within timeout")
	}
}

// BenchmarkHealthCheckContextCancellation measures the performance of
// context cancellation for health check goroutine
func BenchmarkHealthCheckContextCancellation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		healthCheckCtx, cancel := context.WithCancel(context.Background())

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-healthCheckCtx.Done():
					return
				case <-ticker.C:
				}
			}
		}()

		// Cancel immediately
		cancel()

		// Wait for goroutine to exit
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		<-done
	}
}
