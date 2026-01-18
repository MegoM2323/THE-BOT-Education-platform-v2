package concurrent

import (
	"sync"
	"testing"
	"time"
)

// TestSafeGo_NoPanic проверяет нормальное выполнение без паники
func TestSafeGo_NoPanic(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	executed := false

	SafeGo(func() {
		defer wg.Done()
		executed = true
	})

	wg.Wait()

	if !executed {
		t.Error("Expected function to execute normally")
	}
}

// TestSafeGo_WithPanic проверяет что panic перехватывается и goroutine не умирает
func TestSafeGo_WithPanic(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	recovered := false

	SafeGo(func() {
		defer func() {
			wg.Done()
		}()

		// Этот panic должен быть перехвачен внутри SafeGo
		panic("test panic")
	})

	// Даём время на выполнение и recovery
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Goroutine завершилась успешно, panic был перехвачен
		recovered = true
	case <-time.After(2 * time.Second):
		t.Fatal("Goroutine did not complete - panic was not recovered")
	}

	if !recovered {
		t.Error("Expected panic to be recovered")
	}
}

// TestSafeGo_MultiplePanics проверяет что несколько panic'ов обрабатываются независимо
func TestSafeGo_MultiplePanics(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 5
	wg.Add(numGoroutines)

	completed := make([]bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		idx := i
		SafeGo(func() {
			defer wg.Done()

			// Каждая goroutine делает panic
			if idx%2 == 0 {
				panic("even panic")
			}

			completed[idx] = true
		})
	}

	// Ждём завершения всех goroutines
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Все goroutines завершились
		for i := 0; i < numGoroutines; i++ {
			if i%2 == 1 && !completed[i] {
				t.Errorf("Expected goroutine %d to complete normally", i)
			}
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Not all goroutines completed")
	}
}
