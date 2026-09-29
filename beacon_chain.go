package ethwallclock

import (
	"sync"
	"time"
)

type EthereumBeaconChain struct {
	slots  *DefaultSlotCreator
	epochs *DefaultEpochCreator

	mu                    sync.RWMutex
	epochChangedCallbacks []func(current Epoch)
	slotChangedCallbacks  []func(current Slot)

	slotCh  chan struct{}
	epochCh chan struct{}
	stopCh  chan struct{}
	stopped bool
}

func NewEthereumBeaconChain(genesis time.Time, durationPerSlot time.Duration, slotsPerEpoch uint64) *EthereumBeaconChain {
	return newEthereumBeaconChain(
		NewDefaultSlotCreator(genesis, durationPerSlot),
		NewDefaultEpochCreator(genesis, durationPerSlot, slotsPerEpoch),
	)
}

// NewEthereumBeaconChainWithSchedule creates a wall clock for a chain whose
// slot duration changes over time (EIP-8198 SLOT_DURATION_SCHEDULE: a list
// of {epoch, slot duration} entries starting at epoch 0). Slot and epoch
// boundaries, and the OnSlotChanged/OnEpochChanged callbacks, follow the
// schedule.
func NewEthereumBeaconChainWithSchedule(genesis time.Time, schedule []SlotDuration, slotsPerEpoch uint64) (*EthereumBeaconChain, error) {
	slots, err := NewDefaultSlotCreatorWithSchedule(genesis, schedule, slotsPerEpoch)
	if err != nil {
		return nil, err
	}

	epochs, err := NewDefaultEpochCreatorWithSchedule(genesis, schedule, slotsPerEpoch)
	if err != nil {
		return nil, err
	}

	return newEthereumBeaconChain(slots, epochs), nil
}

func newEthereumBeaconChain(slots *DefaultSlotCreator, epochs *DefaultEpochCreator) *EthereumBeaconChain {
	e := &EthereumBeaconChain{
		slots:  slots,
		epochs: epochs,

		epochChangedCallbacks: []func(current Epoch){},
		slotChangedCallbacks:  []func(current Slot){},

		slotCh:  make(chan struct{}),
		epochCh: make(chan struct{}),
		stopCh:  make(chan struct{}),
		stopped: false,
	}

	go func() {
		for {
			select {
			case <-e.slotCh:
				return
			case <-e.stopCh:
				return
			default:
				slot := e.slots.Current()

				time.Sleep(time.Until(slot.TimeWindow().End()))

				slot = e.slots.Current()

				// Take a read lock and copy the callbacks.
				e.mu.RLock()

				if e.stopped {
					e.mu.RUnlock()

					return
				}

				callbacks := make([]func(current Slot), len(e.slotChangedCallbacks))
				copy(callbacks, e.slotChangedCallbacks)
				e.mu.RUnlock()

				// Execute callbacks from our copy.
				for _, callback := range callbacks {
					go callback(slot)
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-e.epochCh:
				return
			case <-e.stopCh:
				return
			default:
				epoch := e.epochs.Current()

				time.Sleep(time.Until(epoch.TimeWindow().End()))

				epoch = e.epochs.Current()

				// Take a read lock and copy the callbacks.
				e.mu.RLock()

				if e.stopped {
					e.mu.RUnlock()

					return
				}

				callbacks := make([]func(current Epoch), len(e.epochChangedCallbacks))
				copy(callbacks, e.epochChangedCallbacks)
				e.mu.RUnlock()

				// Execute callbacks from our copy.
				for _, callback := range callbacks {
					go callback(epoch)
				}
			}
		}
	}()

	return e
}

func (e *EthereumBeaconChain) Now() (Slot, Epoch, error) {
	slot := e.slots.Current()
	epoch := e.epochs.Current()

	return slot, epoch, nil
}

func (e *EthereumBeaconChain) FromTime(t time.Time) (Slot, Epoch, error) {
	slot := e.slots.FromTime(t)
	epoch := e.epochs.FromTime(t)

	return slot, epoch, nil
}

func (e *EthereumBeaconChain) Slots() *DefaultSlotCreator {
	return e.slots
}

func (e *EthereumBeaconChain) Epochs() *DefaultEpochCreator {
	return e.epochs
}

func (e *EthereumBeaconChain) OnEpochChanged(callback func(current Epoch)) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.stopped {
		return
	}

	e.epochChangedCallbacks = append(e.epochChangedCallbacks, callback)
}

func (e *EthereumBeaconChain) OnSlotChanged(callback func(current Slot)) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.stopped {
		return
	}

	e.slotChangedCallbacks = append(e.slotChangedCallbacks, callback)
}

func (e *EthereumBeaconChain) Stop() {
	e.mu.Lock()

	if e.stopped {
		e.mu.Unlock()

		return
	}

	e.stopped = true
	e.mu.Unlock()

	close(e.stopCh)

	// Send a signal to the other channels, but don't close them yet
	// to avoid "send on closed channel" panics from any other goroutines.
	select {
	case e.slotCh <- struct{}{}:
	default:
	}

	select {
	case e.epochCh <- struct{}{}:
	default:
	}

	// Small delay to allow goroutines to exit
	time.Sleep(100 * time.Millisecond)

	// Now safe to close
	close(e.slotCh)
	close(e.epochCh)
}
