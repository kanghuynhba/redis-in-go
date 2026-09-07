package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestConcurrentSetGet tests multiple goroutines concurrently reading and writing their own keys.
func TestConcurrentSetGet(t *testing.T) {
	store := NewStore()
	numGoroutines := 50
	opsPerGoroutine := 100

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := fmt.Sprintf("key-%d-%d", workerID, j)
				val := fmt.Sprintf("val-%d-%d", workerID, j)

				// Concurrent SET
				store.Set(key, val)

				// Concurrent GET
				got, exists := store.Get(key)
				if !exists || got != val {
					t.Errorf("expected %s, got %s (exists=%v)", val, got, exists)
				}
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentSharedKeys tests multiple reader and writer goroutines hammering the same set of keys.
func TestConcurrentSharedKeys(t *testing.T) {
	store := NewStore()
	numWorkers := 20
	numKeys := 5
	ops := 200

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(2)

		// Writer goroutines
		go func(id int) {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				key := fmt.Sprintf("shared-key-%d", j%numKeys)
				val := fmt.Sprintf("worker-%d-op-%d", id, j)
				store.Set(key, val)
			}
		}(i)

		// Reader goroutines
		go func() {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				key := fmt.Sprintf("shared-key-%d", j%numKeys)
				store.Get(key)
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentSetGetWithExpiry tests concurrent writes and reads with TTL expiry.
func TestConcurrentSetGetWithExpiry(t *testing.T) {
	store := NewStore()
	numGoroutines := 30
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("exp-key-%d", id)
			val := fmt.Sprintf("exp-val-%d", id)

			// Set with 50ms expiry
			store.SetWithExpiry(key, val, 50*time.Millisecond)

			// Immediate GET should succeed
			if got, ok := store.Get(key); !ok || got != val {
				t.Errorf("immediate get failed for %s: got %s, ok=%v", key, got, ok)
			}

			// Wait for expiration
			time.Sleep(70 * time.Millisecond)

			// GET after expiration should return false
			if _, ok := store.Get(key); ok {
				t.Errorf("expected key %s to be expired, but it was found", key)
			}
		}(i)
	}

	wg.Wait()
}
