package ethwallclock

import (
	"testing"
	"time"
)

// 12s slots for epochs 0-1, 11s for epochs 2-3, then 8s; 32 slots per epoch.
var testSchedule = []SlotDuration{
	{Epoch: 4, Duration: 8 * time.Second}, // unsorted on purpose
	{Epoch: 0, Duration: 12 * time.Second},
	{Epoch: 2, Duration: 11 * time.Second},
}

func TestSlotCreatorWithSchedule(t *testing.T) {
	genesis := mustParseTime(time.RFC3339, "2020-01-01T12:00:23Z")

	slots, err := NewDefaultSlotCreatorWithSchedule(genesis, testSchedule, 32)
	if err != nil {
		t.Fatal(err)
	}

	epoch2 := 64 * 12 * time.Second
	epoch4 := epoch2 + 64*11*time.Second

	tests := []struct {
		slot  uint64
		start time.Duration
		end   time.Duration
	}{
		{0, 0, 12 * time.Second},
		{63, 63 * 12 * time.Second, epoch2},
		{64, epoch2, epoch2 + 11*time.Second},
		{127, epoch2 + 63*11*time.Second, epoch4},
		{128, epoch4, epoch4 + 8*time.Second},
		{130, epoch4 + 16*time.Second, epoch4 + 24*time.Second},
	}

	for _, test := range tests {
		got := slots.FromNumber(test.slot)
		if !got.TimeWindow().Start().Equal(genesis.Add(test.start)) || !got.TimeWindow().End().Equal(genesis.Add(test.end)) {
			t.Errorf("slot %d: got [%v, %v), want [%v, %v)", test.slot,
				got.TimeWindow().Start().Sub(genesis), got.TimeWindow().End().Sub(genesis), test.start, test.end)
		}

		// FromTime round-trips at the first and last instant of the slot.
		for _, at := range []time.Time{genesis.Add(test.start), genesis.Add(test.end - time.Nanosecond)} {
			if n := slots.FromTime(at); n.Number() != test.slot {
				t.Errorf("FromTime(%v) = %d, want %d", at.Sub(genesis), n.Number(), test.slot)
			}
		}

		if d := slots.DurationAt(test.slot); d != test.end-test.start {
			t.Errorf("DurationAt(%d) = %v, want %v", test.slot, d, test.end-test.start)
		}
	}
}

func TestEpochCreatorWithSchedule(t *testing.T) {
	genesis := mustParseTime(time.RFC3339, "2020-01-01T12:00:23Z")

	epochs, err := NewDefaultEpochCreatorWithSchedule(genesis, testSchedule, 32)
	if err != nil {
		t.Fatal(err)
	}

	epoch2 := 64 * 12 * time.Second
	epoch3 := epoch2 + 32*11*time.Second
	epoch4 := epoch3 + 32*11*time.Second

	got := epochs.FromNumber(3)
	if !got.TimeWindow().Start().Equal(genesis.Add(epoch3)) || !got.TimeWindow().End().Equal(genesis.Add(epoch4)) {
		t.Errorf("epoch 3: got [%v, %v)", got.TimeWindow().Start().Sub(genesis), got.TimeWindow().End().Sub(genesis))
	}

	if e := epochs.FromTime(genesis.Add(epoch4 - time.Nanosecond)); e.Number() != 3 {
		t.Errorf("FromTime(end of epoch 3) = %d", e.Number())
	}

	if e := epochs.FromTime(genesis.Add(epoch4)); e.Number() != 4 {
		t.Errorf("FromTime(start of epoch 4) = %d", e.Number())
	}

	if e := epochs.FromSlot(100); e.Number() != 3 {
		t.Errorf("FromSlot(100) = %d", e.Number())
	}
}

func TestSingleEntryScheduleMatchesFixedDuration(t *testing.T) {
	genesis := mustParseTime(time.RFC3339, "2020-01-01T12:00:23Z")

	fixedSlots := NewDefaultSlotCreator(genesis, 12*time.Second)
	fixedEpochs := NewDefaultEpochCreator(genesis, 12*time.Second, 32)

	slots, err := NewDefaultSlotCreatorWithSchedule(genesis, []SlotDuration{{Epoch: 0, Duration: 12 * time.Second}}, 32)
	if err != nil {
		t.Fatal(err)
	}

	epochs, err := NewDefaultEpochCreatorWithSchedule(genesis, []SlotDuration{{Epoch: 0, Duration: 12 * time.Second}}, 32)
	if err != nil {
		t.Fatal(err)
	}

	for _, offset := range []time.Duration{-25 * time.Second, 0, 11 * time.Second, 12 * time.Second, 12345 * time.Second} {
		at := genesis.Add(offset)

		if a, b := fixedSlots.FromTime(at), slots.FromTime(at); a.Number() != b.Number() || !a.TimeWindow().Start().Equal(b.TimeWindow().Start()) {
			t.Errorf("slot at %v: fixed %d, schedule %d", offset, a.Number(), b.Number())
		}

		if a, b := fixedEpochs.FromTime(at), epochs.FromTime(at); a.Number() != b.Number() || !a.TimeWindow().End().Equal(b.TimeWindow().End()) {
			t.Errorf("epoch at %v: fixed %d, schedule %d", offset, a.Number(), b.Number())
		}
	}
}

func TestScheduleValidation(t *testing.T) {
	genesis := time.Now()

	invalid := map[string][]SlotDuration{
		"empty":             {},
		"no genesis entry":  {{Epoch: 1, Duration: time.Second}},
		"zero duration":     {{Epoch: 0, Duration: 0}},
		"duplicate epochs":  {{Epoch: 0, Duration: time.Second}, {Epoch: 0, Duration: 2 * time.Second}},
		"negative duration": {{Epoch: 0, Duration: -time.Second}},
	}

	for name, schedule := range invalid {
		if _, err := NewEthereumBeaconChainWithSchedule(genesis, schedule, 32); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}

	if _, err := NewEthereumBeaconChainWithSchedule(genesis, testSchedule, 0); err == nil {
		t.Error("zero slots per epoch: expected an error")
	}
}

func TestBeaconChainWithScheduleTicksAcrossDurationChange(t *testing.T) {
	// Genesis 1.5 slots ago with 1s slots in epoch 0 and 200ms slots from
	// epoch 1 (2 slots per epoch): the next boundaries are 0.5s and 0.7s away.
	genesis := time.Now().Add(-1500 * time.Millisecond)

	chain, err := NewEthereumBeaconChainWithSchedule(genesis, []SlotDuration{
		{Epoch: 0, Duration: time.Second},
		{Epoch: 1, Duration: 200 * time.Millisecond},
	}, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer chain.Stop()

	slotCh := make(chan Slot, 4)
	chain.OnSlotChanged(func(current Slot) { slotCh <- current })

	for _, want := range []uint64{2, 3} {
		select {
		case got := <-slotCh:
			if got.Number() != want {
				t.Fatalf("got slot %d, want %d", got.Number(), want)
			}

			if d := got.TimeWindow().End().Sub(got.TimeWindow().Start()); d != 200*time.Millisecond {
				t.Fatalf("slot %d lasts %v, want 200ms", got.Number(), d)
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("timed out waiting for slot %d", want)
		}
	}
}
