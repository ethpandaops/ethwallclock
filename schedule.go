package ethwallclock

import (
	"errors"
	"sort"
	"time"
)

// SlotDuration is one entry of a slot duration schedule: slots from Epoch
// onwards last Duration.
//
// EIP-8198 ("quick slots") replaces the single SLOT_DURATION_MS with a
// SLOT_DURATION_SCHEDULE of such entries, starting at epoch 0.
type SlotDuration struct {
	Epoch    uint64
	Duration time.Duration
}

// slotSchedule converts between slot numbers and offsets from genesis,
// piecewise over a slot duration schedule (the consensus-specs
// compute_time_at_slot_ms / compute_slot_at_time_ms).
type slotSchedule struct {
	entries       []SlotDuration
	slotsPerEpoch uint64
}

func newSlotSchedule(schedule []SlotDuration, slotsPerEpoch uint64) (*slotSchedule, error) {
	if len(schedule) == 0 {
		return nil, errors.New("slot duration schedule is empty")
	}

	if slotsPerEpoch == 0 {
		return nil, errors.New("slots per epoch must be positive")
	}

	entries := make([]SlotDuration, len(schedule))
	copy(entries, schedule)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Epoch < entries[j].Epoch
	})

	if entries[0].Epoch != 0 {
		return nil, errors.New("slot duration schedule must start at epoch 0")
	}

	for i, entry := range entries {
		if entry.Duration <= 0 {
			return nil, errors.New("slot durations must be positive")
		}

		if i > 0 && entry.Epoch == entries[i-1].Epoch {
			return nil, errors.New("slot duration schedule has duplicate epochs")
		}
	}

	return &slotSchedule{entries: entries, slotsPerEpoch: slotsPerEpoch}, nil
}

// singleSlotSchedule is the schedule of a chain whose slot duration never
// changes.
func singleSlotSchedule(durationPerSlot time.Duration, slotsPerEpoch uint64) *slotSchedule {
	if slotsPerEpoch == 0 {
		slotsPerEpoch = 1
	}

	return &slotSchedule{
		entries:       []SlotDuration{{Epoch: 0, Duration: durationPerSlot}},
		slotsPerEpoch: slotsPerEpoch,
	}
}

// slotOffset returns the start of slot relative to genesis.
func (s *slotSchedule) slotOffset(slot uint64) time.Duration {
	endSlot := slot
	offset := time.Duration(0)

	for i := len(s.entries) - 1; i >= 0; i-- {
		entry := s.entries[i]
		entrySlot := entry.Epoch * s.slotsPerEpoch

		if entrySlot < endSlot {
			offset += time.Duration(endSlot-entrySlot) * entry.Duration
			endSlot = entrySlot
		}
	}

	return offset
}

// slotAt returns the slot containing offset (relative to genesis).
func (s *slotSchedule) slotAt(offset time.Duration) uint64 {
	if offset < 0 {
		// Before genesis: keep the historical behaviour of the fixed-duration
		// creators (integer division of the negative offset).
		return uint64(offset / s.entries[0].Duration)
	}

	for i := len(s.entries) - 1; i >= 0; i-- {
		entry := s.entries[i]
		entrySlot := entry.Epoch * s.slotsPerEpoch
		entryOffset := s.slotOffset(entrySlot)

		if offset >= entryOffset {
			return entrySlot + uint64((offset-entryOffset)/entry.Duration)
		}
	}

	return 0
}
