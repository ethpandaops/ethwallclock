package ethwallclock

import (
	"time"
)

func NewDefaultEpochCreator(genesis time.Time, durationPerSlot time.Duration, slotsPerEpoch uint64) *DefaultEpochCreator {
	return &DefaultEpochCreator{
		genesis:         genesis,
		durationPerSlot: durationPerSlot,
		slotsPerEpoch:   slotsPerEpoch,
	}
}

// NewDefaultEpochCreatorWithSchedule creates an epoch creator for a chain
// whose slot duration changes over time (EIP-8198 SLOT_DURATION_SCHEDULE).
// The schedule must start at epoch 0.
func NewDefaultEpochCreatorWithSchedule(genesis time.Time, schedule []SlotDuration, slotsPerEpoch uint64) (*DefaultEpochCreator, error) {
	s, err := newSlotSchedule(schedule, slotsPerEpoch)
	if err != nil {
		return nil, err
	}

	return &DefaultEpochCreator{
		genesis:         genesis,
		durationPerSlot: s.entries[0].Duration,
		slotsPerEpoch:   slotsPerEpoch,
		schedule:        s,
	}, nil
}

type DefaultEpochCreator struct {
	genesis         time.Time
	durationPerSlot time.Duration
	slotsPerEpoch   uint64
	// schedule is set for chains with a slot duration schedule; nil means
	// every slot lasts durationPerSlot.
	schedule *slotSchedule
}

func (e *DefaultEpochCreator) FromNumber(number uint64) Epoch {
	if e.schedule != nil {
		return NewEpoch(number,
			e.genesis.Add(e.schedule.slotOffset(number*e.slotsPerEpoch)),
			e.genesis.Add(e.schedule.slotOffset((number+1)*e.slotsPerEpoch)),
		)
	}

	return NewEpoch(number,
		e.genesis.Add(time.Duration(number*e.slotsPerEpoch)*e.durationPerSlot),
		e.genesis.Add(time.Duration((number+1)*e.slotsPerEpoch)*e.durationPerSlot),
	)
}

func (e *DefaultEpochCreator) Current() Epoch {
	return e.FromTime(time.Now())
}

func (e *DefaultEpochCreator) FromTime(t time.Time) Epoch {
	if e.schedule != nil {
		offset := t.Sub(e.genesis)
		if offset < 0 {
			// Before genesis: same as the fixed-duration arithmetic.
			return e.FromNumber(uint64(offset / (e.durationPerSlot * time.Duration(e.slotsPerEpoch))))
		}

		return e.FromNumber(e.schedule.slotAt(offset) / e.slotsPerEpoch)
	}

	number := uint64(t.Sub(e.genesis) / (e.durationPerSlot * time.Duration(e.slotsPerEpoch)))
	return e.FromNumber(number)
}

func (e *DefaultEpochCreator) FromSlot(slot uint64) Epoch {
	return e.FromNumber(slot / e.slotsPerEpoch)
}
