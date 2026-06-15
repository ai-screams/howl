package internal

import "time"

// UsageWindow is the render model for a single rate-limit window. Presence is
// signalled by a non-nil pointer — zero is a valid quota state, never "absent".
type UsageWindow struct {
	RemainingPercent float64
	ResetsAt         time.Time
}

// UsageData holds quota for display. Each window is independently optional,
// matching the Claude Code `rate_limits` contract.
type UsageData struct {
	FiveHour *UsageWindow
	SevenDay *UsageWindow
}

// UsageFromRateLimits converts the stdin `rate_limits` object into the render
// model. Pure function — no network, cache, or Keychain. Returns nil when no
// quota data is available so the renderer omits the section.
func UsageFromRateLimits(rl *RateLimits) *UsageData {
	if rl == nil {
		return nil
	}
	usage := &UsageData{
		FiveHour: usageWindowFromRateLimit(rl.FiveHour),
		SevenDay: usageWindowFromRateLimit(rl.SevenDay),
	}
	if usage.FiveHour == nil && usage.SevenDay == nil {
		return nil
	}
	return usage
}

func usageWindowFromRateLimit(w *RateLimitWindow) *UsageWindow {
	if w == nil {
		return nil
	}
	used := min(max(w.UsedPercentage, 0), 100) // contract is 0-100; clamp defensively
	return &UsageWindow{
		RemainingPercent: 100 - used,
		ResetsAt:         time.Unix(w.ResetsAt, 0),
	}
}
