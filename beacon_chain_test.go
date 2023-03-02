package ethwallclock

import (
	"testing"
	"time"
)

func TestBeaconChainEventCallbacks(t *testing.T) {
	beacon := NewEthereumBeaconChain(time.Now(), time.Second*1, 2)

	t.Run("Event callbacks", func(t *testing.T) {
		epochCallbacks := 0
		slotCallbacks := 0

		beacon.OnEpochChanged(func(epoch Epoch) {
			epochCallbacks++
		})

		beacon.OnSlotChanged(func(slot Slot) {
			slotCallbacks++
		})

		time.Sleep(5100 * time.Millisecond)

		if epochCallbacks != 2 {
			t.Errorf("incorrect number of epoch callbacks: got %v, want %v", epochCallbacks, 2)
		}

		if slotCallbacks != 5 {
			t.Errorf("incorrect number of slot callbacks: got %v, want %v", slotCallbacks, 5)
		}
	})

	beacon.Stop()
}
