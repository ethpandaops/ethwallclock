package ethwallclock

import (
	"sync"
	"testing"
	"time"
)

// MetadataService mocks the metadata service we use across xatu.
type MetadataService struct {
	wallclock *EthereumBeaconChain
}

// Wallclock returns the wallclock instance (can be nil).
func (m *MetadataService) Wallclock() *EthereumBeaconChain {
	return m.wallclock
}

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

// TestConcurrentStopAndCallback tests that there's no race condition
// between stopping the beacon chain and registering/executing callbacks.
func TestConcurrentStopAndCallback(t *testing.T) {
	beacon := NewEthereumBeaconChain(time.Now(), time.Second*1, 2)

	// Set up a sync WaitGroup to coordinate goroutines.
	var wg sync.WaitGroup

	// Start multiple goroutines that try to register callbacks.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Try to register a callback - should not panic even if Stop is called concurrently.
			beacon.OnEpochChanged(func(epoch Epoch) {})
		}(i)
	}

	// Start a goroutine that stops the beacon chain.
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Small delay to increase chance of concurrent execution
		time.Sleep(5 * time.Millisecond)
		beacon.Stop()
	}()

	// Wait for all goroutines to finish.
	wg.Wait()
}

// TestNilWallclockScenario specifically tests for the panic seen in production:
// when OnEpochChanged is called on a nil receiver.
func TestNilWallclockScenario(t *testing.T) {
	// Create a metadata service with a valid wallclock
	metadata := &MetadataService{
		wallclock: NewEthereumBeaconChain(time.Now(), time.Second*1, 2),
	}

	wc := metadata.Wallclock()
	if wc == nil {
		t.Fatal("Wallclock should not be nil")
	}

	// Register a callback.
	wc.OnEpochChanged(func(epoch Epoch) {})

	// If the beacon chain connection fails/is-lost, the wallclock becomes nil.
	// Subsequent callbacks then attempt to call the nil wallclock, which panics.
	metadata.wallclock = nil

	shouldPanic := func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when using nil wallclock, but no panic occurred")
			} else {
				t.Logf("Got expected panic: %v", r)
			}
		}()

		wc := metadata.Wallclock() // Get nil wallclock.
		wc.OnEpochChanged(func(epoch Epoch) {})
	}

	shouldPanic()
}
