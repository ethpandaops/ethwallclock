package ethwallclock

import (
	"math"
	"time"
)

// slotTiming converts between slot numbers and offsets from genesis. Slots
// last durationPerSlot until forkSlot and forkDuration from then on; a zero
// forkDuration means the slot duration never changes.
type slotTiming struct {
	durationPerSlot time.Duration
	forkSlot        uint64
	forkDuration    time.Duration
}

func newSlotTiming(durationPerSlot time.Duration, slotsPerEpoch uint64, opts []Option) slotTiming {
	cfg := config{}
	for _, opt := range opts {
		opt(&cfg)
	}

	timing := slotTiming{durationPerSlot: durationPerSlot}

	if cfg.hezeSlotDuration > 0 && slotsPerEpoch > 0 && cfg.hezeForkEpoch <= math.MaxUint64/slotsPerEpoch {
		timing.forkSlot = cfg.hezeForkEpoch * slotsPerEpoch
		timing.forkDuration = cfg.hezeSlotDuration
	}

	return timing
}

func (t slotTiming) slotOffset(slot uint64) time.Duration {
	if t.forkDuration == 0 || slot <= t.forkSlot {
		return time.Duration(slot) * t.durationPerSlot
	}

	return time.Duration(t.forkSlot)*t.durationPerSlot + time.Duration(slot-t.forkSlot)*t.forkDuration
}

func (t slotTiming) slotAt(offset time.Duration) uint64 {
	forkOffset := time.Duration(t.forkSlot) * t.durationPerSlot
	if t.forkDuration == 0 || offset < forkOffset {
		return uint64(offset / t.durationPerSlot)
	}

	return t.forkSlot + uint64((offset-forkOffset)/t.forkDuration)
}

func (t slotTiming) duration(slot uint64) time.Duration {
	if t.forkDuration == 0 || slot < t.forkSlot {
		return t.durationPerSlot
	}

	return t.forkDuration
}
