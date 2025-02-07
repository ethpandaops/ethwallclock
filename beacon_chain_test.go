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
