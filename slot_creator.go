package ethwallclock

import (
	"time"
)

func NewDefaultSlotCreator(genesis time.Time, durationPerSlot time.Duration) *DefaultSlotCreator {
	return &DefaultSlotCreator{
		genesis:         genesis,
		durationPerSlot: durationPerSlot,
	}
}

// NewDefaultSlotCreatorWithSchedule creates a slot creator for a chain whose
// slot duration changes over time (EIP-8198 SLOT_DURATION_SCHEDULE). The
// schedule must start at epoch 0.
func NewDefaultSlotCreatorWithSchedule(genesis time.Time, schedule []SlotDuration, slotsPerEpoch uint64) (*DefaultSlotCreator, error) {
	s, err := newSlotSchedule(schedule, slotsPerEpoch)
	if err != nil {
		return nil, err
	}

	return &DefaultSlotCreator{
		genesis:         genesis,
		durationPerSlot: s.entries[0].Duration,
		schedule:        s,
	}, nil
}

type DefaultSlotCreator struct {
	genesis         time.Time
	durationPerSlot time.Duration
	// schedule is set for chains with a slot duration schedule; nil means
	// every slot lasts durationPerSlot.
	schedule *slotSchedule
}

func (s *DefaultSlotCreator) FromNumber(number uint64) Slot {
	if s.schedule != nil {
		return NewSlot(number, s.genesis.Add(s.schedule.slotOffset(number)), s.genesis.Add(s.schedule.slotOffset(number+1)))
	}

	return NewSlot(number, s.genesis.Add(time.Duration(number)*s.durationPerSlot), s.genesis.Add(time.Duration(number+1)*s.durationPerSlot))
}

func (s *DefaultSlotCreator) Current() Slot {
	return s.FromTime(time.Now())
}

func (s *DefaultSlotCreator) FromTime(t time.Time) Slot {
	if s.schedule != nil {
		return s.FromNumber(s.schedule.slotAt(t.Sub(s.genesis)))
	}

	number := uint64(t.Sub(s.genesis) / s.durationPerSlot)
	return s.FromNumber(number)
}

// DurationAt returns the duration of the given slot.
func (s *DefaultSlotCreator) DurationAt(number uint64) time.Duration {
	if s.schedule != nil {
		return s.schedule.slotOffset(number+1) - s.schedule.slotOffset(number)
	}

	return s.durationPerSlot
}
