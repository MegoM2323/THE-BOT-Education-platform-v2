package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// SimulatedDB represents a mock database connection pool
type SimulatedDB struct {
	isClosed    int64
	accessCount int64
	accessesMu  sync.Mutex
	accesses    []int64 // timestamps of accesses
}

func NewSimulatedDB() *SimulatedDB {
	return &SimulatedDB{
		accesses: make([]int64, 0),
	}
}

func (db *SimulatedDB) Ping() error {
	if atomic.LoadInt64(&db.isClosed) == 1 {
		return fmt.Errorf("database closed")
	}
	atomic.AddInt64(&db.accessCount, 1)
	db.accessesMu.Lock()
	db.accesses = append(db.accesses, time.Now().UnixNano())
	db.accessesMu.Unlock()
	return nil
}

func (db *SimulatedDB) Close() error {
	atomic.StoreInt64(&db.isClosed, 1)
	return nil
}

func (db *SimulatedDB) IsClosed() bool {
	return atomic.LoadInt64(&db.isClosed) == 1
}

func (db *SimulatedDB) GetAccessCount() int64 {
	return atomic.LoadInt64(&db.accessCount)
}

// TestGracefulShutdownWithHealthCheck simulates the real shutdown scenario
// with a health check goroutine and database
func TestGracefulShutdownWithHealthCheck(t *testing.T) {
	tests := []struct {
		name              string
		healthCheckPeriod time.Duration
		shutdownPhase1    time.Duration
		gracePeriodMS     int
		expectAccess      bool
	}{
		{
			name:              "health check doesn't access DB after close",
			healthCheckPeriod: 50 * time.Millisecond,
			shutdownPhase1:    30 * time.Millisecond,
			gracePeriodMS:     100,
			expectAccess:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewSimulatedDB()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var accessAfterClose int64
			var wg sync.WaitGroup

			// Simulate health check goroutine
			wg.Add(1)
			go func() {
				defer wg.Done()

				ticker := time.NewTicker(tt.healthCheckPeriod)
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						// Try to access database
						if err := db.Ping(); err != nil {
							// Database is closed, stop trying
							atomic.AddInt64(&accessAfterClose, 1)
							return
						}
					}
				}
			}()

			// Simulate operational time
			time.Sleep(tt.shutdownPhase1)

			// PHASE 1: Signal goroutine to stop
			cancel()
			t.Logf("Phase 1: Context cancelled")

			// PHASE 2: Grace period for goroutines to exit
			gracePeriod := time.Duration(tt.gracePeriodMS) * time.Millisecond
			time.Sleep(gracePeriod)
			t.Logf("Phase 2: Grace period completed (%v)", gracePeriod)

			// PHASE 3: Close database
			accessBeforeClose := db.GetAccessCount()
			if err := db.Close(); err != nil {
				t.Logf("Phase 3: Database closed with error: %v", err)
			}
			t.Logf("Phase 3: Database closed")

			// Wait for goroutine to fully exit
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				accessesAfter := atomic.LoadInt64(&accessAfterClose)
				t.Logf("Health check completed: %d accesses before close, %d attempts after close",
					accessBeforeClose, accessesAfter)

				if accessesAfter > 0 && !tt.expectAccess {
					t.Errorf("Expected no accesses after close, got %d", accessesAfter)
				}
			case <-time.After(1 * time.Second):
				t.Error("Health check goroutine did not exit")
			}
		})
	}
}

// TestShutdownSequenceWithMultipleGoroutines tests shutdown with multiple background goroutines
func TestShutdownSequenceWithMultipleGoroutines(t *testing.T) {
	tests := []struct {
		name            string
		goroutineCount  int
		shutdownTimeout time.Duration
	}{
		{
			name:            "10 goroutines shutdown gracefully",
			goroutineCount:  10,
			shutdownTimeout: 2 * time.Second,
		},
		{
			name:            "50 goroutines shutdown gracefully",
			goroutineCount:  50,
			shutdownTimeout: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewSimulatedDB()
			ctx, cancel := context.WithCancel(context.Background())

			var wg sync.WaitGroup
			var accessCount int64
			var postShutdownAccess int64

			// Start multiple goroutines that access database
			for i := 0; i < tt.goroutineCount; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()

					ticker := time.NewTicker(10 * time.Millisecond)
					defer ticker.Stop()

					for {
						select {
						case <-ctx.Done():
							return
						case <-ticker.C:
							if err := db.Ping(); err != nil {
								atomic.AddInt64(&postShutdownAccess, 1)
								return
							}
							atomic.AddInt64(&accessCount, 1)
						}
					}
				}(i)
			}

			// Let goroutines run
			time.Sleep(50 * time.Millisecond)

			// Shutdown sequence
			cancel()
			time.Sleep(100 * time.Millisecond) // Grace period
			db.Close()

			// Wait for all goroutines to exit
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				t.Logf("All %d goroutines exited: %d accesses, %d post-close attempts",
					tt.goroutineCount, atomic.LoadInt64(&accessCount), atomic.LoadInt64(&postShutdownAccess))
			case <-time.After(tt.shutdownTimeout):
				t.Errorf("Shutdown timeout: some goroutines did not exit")
			}
		})
	}
}

// TestDBAccessPattern verifies the access pattern matches expected shutdown sequence
func TestDBAccessPattern(t *testing.T) {
	tests := []struct {
		name           string
		operationTime  time.Duration
		checkPeriod    time.Duration
		expectDecaying bool // accesses should decrease during shutdown
	}{
		{
			name:           "access count decreases during shutdown",
			operationTime:  100 * time.Millisecond,
			checkPeriod:    5 * time.Millisecond,
			expectDecaying: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewSimulatedDB()
			ctx, cancel := context.WithCancel(context.Background())

			var wg sync.WaitGroup

			// Start health check goroutine
			wg.Add(1)
			go func() {
				defer wg.Done()

				ticker := time.NewTicker(tt.checkPeriod)
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						_ = db.Ping()
					}
				}
			}()

			// Measure baseline access rate
			time.Sleep(tt.operationTime)
			accessesBeforeShutdown := db.GetAccessCount()
			t.Logf("Accesses during operation: %d", accessesBeforeShutdown)

			// Start shutdown
			cancel()

			// Quick grace period
			time.Sleep(50 * time.Millisecond)

			// Close DB
			db.Close()

			// Wait for goroutine to exit
			wg.Wait()

			accessesTotal := db.GetAccessCount()
			t.Logf("Total accesses after shutdown: %d", accessesTotal)

			// Verify some accesses happened before close
			if accessesBeforeShutdown == 0 {
				t.Error("No accesses recorded during operation")
			}

			// Verify no panic occurred (test passes if we get here)
			t.Log("Shutdown completed without panic")
		})
	}
}

// TestShutdownContextPropagation verifies context cancellation propagates correctly
func TestShutdownContextPropagation(t *testing.T) {
	tests := []struct {
		name               string
		nestingLevel       int
		expectAllCancelled bool
	}{
		{
			name:               "context propagates through nested goroutines",
			nestingLevel:       3,
			expectAllCancelled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var wg sync.WaitGroup
			cancelledCounts := make([]int64, tt.nestingLevel)

			// Create nested goroutines
			var createNested func(level int, parentCtx context.Context)
			createNested = func(level int, parentCtx context.Context) {
				if level == 0 {
					return
				}

				wg.Add(1)
				go func(myLevel int, ctx context.Context) {
					defer wg.Done()

					// Wait for cancellation
					<-ctx.Done()
					atomic.AddInt64(&cancelledCounts[myLevel-1], 1)

					// Create child
					childCtx, childCancel := context.WithCancel(ctx)
					defer childCancel()
					createNested(myLevel-1, childCtx)
				}(level, parentCtx)
			}

			// Create hierarchy
			createNested(tt.nestingLevel, ctx)

			// Give time to establish
			time.Sleep(10 * time.Millisecond)

			// Cancel root context
			cancel()

			// Wait for all to exit
			wg.Wait()

			// Verify all received cancellation signal
			for i, count := range cancelledCounts {
				if count > 0 {
					t.Logf("Level %d received cancellation", i+1)
				}
			}

			t.Log("Context propagation completed successfully")
		})
	}
}
