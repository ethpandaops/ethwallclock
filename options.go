package ethwallclock

import "time"

// Option configures a wall clock.
type Option func(*config)

type config struct {
	hezeForkEpoch    uint64
	hezeSlotDuration time.Duration
}

// WithHezeSlotDuration switches the slot duration to SLOT_DURATION_MS_HEZE from
// HEZE_FORK_EPOCH onwards. A zero duration or an unreachable epoch
// (FAR_FUTURE_EPOCH) leaves the slot duration unchanged.
func WithHezeSlotDuration(hezeForkEpoch uint64, slotDuration time.Duration) Option {
	return func(c *config) {
		c.hezeForkEpoch = hezeForkEpoch
		c.hezeSlotDuration = slotDuration
	}
}
