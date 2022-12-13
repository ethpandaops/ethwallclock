package ethwallclock

import (
	"time"
)

type EthereumBeaconChain struct {
	slots  *DefaultSlotCreator
	epochs *DefaultEpochCreator

	epochChangedCallbacks []func(current Epoch)
	slotChangedCallbacks  []func(current Slot)
}

func NewEthereumBeaconChain(genesis time.Time, durationPerSlot time.Duration, slotsPerEpoch uint64) *EthereumBeaconChain {
	e := &EthereumBeaconChain{
		slots:  NewDefaultSlotCreator(genesis, durationPerSlot),
		epochs: NewDefaultEpochCreator(genesis, durationPerSlot, slotsPerEpoch),

		epochChangedCallbacks: []func(current Epoch){},
		slotChangedCallbacks:  []func(current Slot){},
	}

	go func() {
		for {
			slot := e.slots.Current()

			time.Sleep(slot.TimeWindow().End().Sub(time.Now()))

			slot = e.slots.Current()
			for _, callback := range e.slotChangedCallbacks {
				go callback(slot)
			}
		}
	}()

	go func() {
		for {
			epoch := e.epochs.Current()

			time.Sleep(epoch.TimeWindow().End().Sub(time.Now()))

			epoch = e.epochs.Current()
			for _, callback := range e.epochChangedCallbacks {
				go callback(epoch)
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
	e.epochChangedCallbacks = append(e.epochChangedCallbacks, callback)
}

func (e *EthereumBeaconChain) OnSlotChanged(callback func(current Slot)) {
	e.slotChangedCallbacks = append(e.slotChangedCallbacks, callback)
}
