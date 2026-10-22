package ethwallclock

import (
	"time"
)

func NewDefaultEpochCreator(genesis time.Time, durationPerSlot time.Duration, slotsPerEpoch uint64, opts ...Option) *DefaultEpochCreator {
	return &DefaultEpochCreator{
		genesis:       genesis,
		timing:        newSlotTiming(durationPerSlot, slotsPerEpoch, opts),
		slotsPerEpoch: slotsPerEpoch,
	}
}

type DefaultEpochCreator struct {
	genesis       time.Time
	timing        slotTiming
	slotsPerEpoch uint64
}

func (e *DefaultEpochCreator) FromNumber(number uint64) Epoch {
	return NewEpoch(number,
		e.genesis.Add(e.timing.slotOffset(number*e.slotsPerEpoch)),
		e.genesis.Add(e.timing.slotOffset((number+1)*e.slotsPerEpoch)),
	)
}

func (e *DefaultEpochCreator) Current() Epoch {
	return e.FromTime(time.Now())
}

func (e *DefaultEpochCreator) FromTime(t time.Time) Epoch {
	offset := t.Sub(e.genesis)
	if offset < 0 {
		return e.FromNumber(uint64(offset / (e.timing.durationPerSlot * time.Duration(e.slotsPerEpoch))))
	}

	return e.FromNumber(e.timing.slotAt(offset) / e.slotsPerEpoch)
}

func (e *DefaultEpochCreator) FromSlot(slot uint64) Epoch {
	return e.FromNumber(slot / e.slotsPerEpoch)
}
