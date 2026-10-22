package ethwallclock

import (
	"math"
	"testing"
	"time"
)

func TestHezeSlotDuration(t *testing.T) {
	genesis := time.Unix(1_700_000_000, 0)
	chain := NewEthereumBeaconChain(genesis, 12*time.Second, 32, WithHezeSlotDuration(5, 9*time.Second))
	defer chain.Stop()

	slots := chain.Slots()
	epochs := chain.Epochs()

	if got := slotStart(slots.FromNumber(159)); !got.Equal(genesis.Add(159 * 12 * time.Second)) {
		t.Fatalf("pre-fork slot start wrong: %v", got)
	}

	forkStart := genesis.Add(160 * 12 * time.Second)
	if got := slotStart(slots.FromNumber(160)); !got.Equal(forkStart) {
		t.Fatalf("fork slot start wrong: %v", got)
	}

	if got := slotStart(slots.FromNumber(161)); !got.Equal(forkStart.Add(9 * time.Second)) {
		t.Fatalf("post-fork slot start wrong: %v", got)
	}

	if got := slotNumber(slots.FromTime(forkStart.Add(9*time.Second*228 + 4*time.Second))); got != 388 {
		t.Fatalf("expected slot 388, got %d", got)
	}

	if got := slotNumber(slots.FromTime(forkStart.Add(-time.Second))); got != 159 {
		t.Fatalf("expected slot 159 just before the fork, got %d", got)
	}

	if slots.DurationAt(159) != 12*time.Second || slots.DurationAt(160) != 9*time.Second {
		t.Fatal("wrong slot durations around the fork")
	}

	epoch6 := epochs.FromNumber(6)
	window := epoch6.TimeWindow()
	if !window.Start().Equal(forkStart.Add(32*9*time.Second)) || !window.End().Equal(forkStart.Add(64*9*time.Second)) {
		t.Fatalf("post-fork epoch window wrong: %v - %v", window.Start(), window.End())
	}

	if got := epochNumber(epochs.FromTime(forkStart.Add(32*9*time.Second + time.Second))); got != 6 {
		t.Fatalf("expected epoch 6, got %d", got)
	}
}

func TestHezeSlotDurationUnscheduled(t *testing.T) {
	genesis := time.Unix(1_700_000_000, 0)

	for _, opt := range []Option{WithHezeSlotDuration(math.MaxUint64, 9*time.Second), WithHezeSlotDuration(5, 0), WithHezeSlotDuration(5, 12*time.Second)} {
		chain := NewEthereumBeaconChain(genesis, 12*time.Second, 32, opt)

		if got := slotStart(chain.Slots().FromNumber(1000)); !got.Equal(genesis.Add(1000 * 12 * time.Second)) {
			t.Fatalf("slot duration changed without a scheduled fork: %v", got)
		}

		chain.Stop()
	}
}

func slotStart(slot Slot) time.Time {
	return slot.TimeWindow().Start()
}

func slotNumber(slot Slot) uint64 {
	return slot.Number()
}

func epochNumber(epoch Epoch) uint64 {
	return epoch.Number()
}
