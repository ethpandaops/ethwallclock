package ethwallclock

import (
	"sync"
	"testing"
	"time"
)

func TestBeaconChainEventCallbacks(t *testing.T) {
	beacon := NewEthereumBeaconChain(time.Now(), time.Second*1, 2)

	t.Run("Event callbacks", func(t *testing.T) {
		var (
			mu             sync.Mutex
			epochCallbacks int
			slotCallbacks  int
		)

		beacon.OnEpochChanged(func(epoch Epoch) {
			mu.Lock()
			epochCallbacks++
			mu.Unlock()
		})

		beacon.OnSlotChanged(func(slot Slot) {
			mu.Lock()
			slotCallbacks++
			mu.Unlock()
		})

		time.Sleep(5100 * time.Millisecond)

		mu.Lock()
		epochCount := epochCallbacks
		slotCount := slotCallbacks
		mu.Unlock()

		if epochCount != 2 {
			t.Errorf("incorrect number of epoch callbacks: got %v, want %v", epochCount, 2)
		}

		if slotCount != 5 {
			t.Errorf("incorrect number of slot callbacks: got %v, want %v", slotCount, 5)
		}
	})

	beacon.Stop()
}

func TestConcurrentCallbackRegistration(t *testing.T) {
	// This test verifies two potential race conditions:
	// 1. Concurrent registration of callbacks (modifying the callback slices)
	// 2. Concurrent execution of callbacks (callbacks running in parallel)

	const (
		numCallbacks = 100
		slotDuration = 10 * time.Millisecond
		testDuration = 200 * time.Millisecond
	)

	beacon := NewEthereumBeaconChain(time.Now(), slotDuration, 2)
	defer beacon.Stop()

	var (
		wg                  sync.WaitGroup
		mu                  sync.Mutex
		callbacksExecuted   int
		callbacksRegistered int
	)

	// Concurrently register callbacks while slots are ticking
	for i := 0; i < numCallbacks; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			beacon.OnSlotChanged(func(slot Slot) {
				// Track both registration and execution
				mu.Lock()
				callbacksExecuted++
				mu.Unlock()
			})

			mu.Lock()
			callbacksRegistered++
			mu.Unlock()
		}()
	}

	// Let the test run for a fixed duration
	time.Sleep(testDuration)

	// Wait for all registration goroutines to complete
	wg.Wait()

	// Check results
	mu.Lock()
	registered := callbacksRegistered
	executed := callbacksExecuted
	mu.Unlock()

	// Verify all callbacks were registered
	if registered != numCallbacks {
		t.Errorf("not all callbacks were registered: got %d, want %d", registered, numCallbacks)
	}

	// Verify callbacks were actually executed
	// We should have at least some executions given our test duration and slot duration
	expectedMinExecutions := int(testDuration/slotDuration) * numCallbacks / 2
	if executed < expectedMinExecutions {
		t.Errorf("too few callback executions: got %d, want at least %d", executed, expectedMinExecutions)
	}

	t.Logf("Test completed: registered %d callbacks, executed %d times", registered, executed)
}

/*
BenchmarkCallbackRegistration-10                 8334013               217.0 ns/op            65 B/op          0 allocs/op
BenchmarkCallbackRegistration-10                 7444645               258.7 ns/op            77 B/op          0 allocs/op
BenchmarkCallbackRegistration-10                 7049173               269.3 ns/op            80 B/op          0 allocs/op
BenchmarkCallbackRegistration-10                 7852233               230.5 ns/op            70 B/op          0 allocs/op
BenchmarkCallbackRegistration-10                 7653772               387.5 ns/op            93 B/op          1 allocs/op
*/
func BenchmarkCallbackRegistration(b *testing.B) {
	beacon := NewEthereumBeaconChain(time.Now(), time.Millisecond, 2)
	defer beacon.Stop()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			beacon.OnSlotChanged(func(slot Slot) {})
		}
	})
}

/*
BenchmarkCallbackRegistrationSequential-10      19174398               191.2 ns/op            66 B/op          0 allocs/op
BenchmarkCallbackRegistrationSequential-10      22953746               159.0 ns/op            65 B/op          0 allocs/op
BenchmarkCallbackRegistrationSequential-10      17665375               221.3 ns/op            74 B/op          0 allocs/op
BenchmarkCallbackRegistrationSequential-10      16663228               234.8 ns/op            80 B/op          0 allocs/op
BenchmarkCallbackRegistrationSequential-10      22371304               187.7 ns/op            69 B/op          0 allocs/op
*/
func BenchmarkCallbackRegistrationSequential(b *testing.B) {
	beacon := NewEthereumBeaconChain(time.Now(), time.Millisecond, 2)
	defer beacon.Stop()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		beacon.OnSlotChanged(func(slot Slot) {})
	}
}
