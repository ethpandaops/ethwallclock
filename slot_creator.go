package ethwallclock

import (
	"time"
)

func NewDefaultSlotCreator(genesis time.Time, durationPerSlot time.Duration) *DefaultSlotCreator {
	return newSlotCreator(genesis, durationPerSlot, 0, nil)
}

func newSlotCreator(genesis time.Time, durationPerSlot time.Duration, slotsPerEpoch uint64, opts []Option) *DefaultSlotCreator {
	return &DefaultSlotCreator{
		genesis: genesis,
		timing:  newSlotTiming(durationPerSlot, slotsPerEpoch, opts),
	}
}

type DefaultSlotCreator struct {
	genesis time.Time
	timing  slotTiming
}

func (s *DefaultSlotCreator) FromNumber(number uint64) Slot {
	return NewSlot(number, s.genesis.Add(s.timing.slotOffset(number)), s.genesis.Add(s.timing.slotOffset(number+1)))
}

func (s *DefaultSlotCreator) Current() Slot {
	return s.FromTime(time.Now())
}

func (s *DefaultSlotCreator) FromTime(t time.Time) Slot {
	return s.FromNumber(s.timing.slotAt(t.Sub(s.genesis)))
}

// DurationAt returns the duration of the given slot.
func (s *DefaultSlotCreator) DurationAt(number uint64) time.Duration {
	return s.timing.duration(number)
}
