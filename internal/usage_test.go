package internal

import (
	"testing"
	"time"
)

func TestQuotaColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		remaining float64
		want      string
	}{
		{"critical below 10", 5.0, boldRed},
		{"critical at boundary", 9.99, boldRed},
		{"low at 10", 10.0, red},
		{"low below 25", 24.9, red},
		{"medium at 25", 25.0, orange},
		{"medium below 50", 49.9, orange},
		{"high at 50", 50.0, yellow},
		{"high below 75", 74.9, yellow},
		{"healthy at 75", 75.0, green},
		{"healthy at 100", 100.0, green},
		{"zero critical", 0.0, boldRed},
		{"edge case 10.01", 10.01, red},
		{"edge case 25.01", 25.01, orange},
		{"edge case 50.01", 50.01, yellow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := quotaColor(tt.remaining, DefaultThresholds())
			if got != tt.want {
				t.Errorf("quotaColor(%v) = %q, want %q", tt.remaining, got, tt.want)
			}
		})
	}
}

func TestFormatTimeUntil(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2026, 2, 7, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		now    time.Time
		target time.Time
		want   string
	}{
		{"zero target", baseTime, time.Time{}, "?"},
		{"expired past", baseTime, baseTime.Add(-1 * time.Hour), "0d00h"},
		{"expired exact", baseTime, baseTime, "0d00h"},
		{"3 hours future", baseTime, baseTime.Add(3 * time.Hour), "0d03h"},
		{"25 hours future", baseTime, baseTime.Add(25 * time.Hour), "1d01h"},
		{"48 hours future", baseTime, baseTime.Add(48 * time.Hour), "2d00h"},
		{"1 minute future", baseTime, baseTime.Add(1 * time.Minute), "0d00h"},
		{"23 hours 59 minutes", baseTime, baseTime.Add(23*time.Hour + 59*time.Minute), "0d23h"},
		{"7 days exact", baseTime, baseTime.Add(7 * 24 * time.Hour), "7d00h"},
		{"3 days 12 hours", baseTime, baseTime.Add(3*24*time.Hour + 12*time.Hour), "3d12h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatTimeUntil(tt.now, tt.target)
			if got != tt.want {
				t.Errorf("formatTimeUntil() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUsageFromRateLimits(t *testing.T) {
	t.Parallel()

	t.Run("nil rate_limits returns nil", func(t *testing.T) {
		t.Parallel()
		if got := UsageFromRateLimits(nil); got != nil {
			t.Errorf("UsageFromRateLimits(nil) = %+v, want nil", got)
		}
	})

	t.Run("both windows absent returns nil", func(t *testing.T) {
		t.Parallel()
		if got := UsageFromRateLimits(&RateLimits{}); got != nil {
			t.Errorf("UsageFromRateLimits(empty) = %+v, want nil", got)
		}
	})

	t.Run("both windows present", func(t *testing.T) {
		t.Parallel()
		got := UsageFromRateLimits(&RateLimits{
			FiveHour: &RateLimitWindow{UsedPercentage: 23.5, ResetsAt: 1738425600},
			SevenDay: &RateLimitWindow{UsedPercentage: 41.2, ResetsAt: 1738857600},
		})
		if got == nil {
			t.Fatal("UsageFromRateLimits() = nil, want non-nil")
		}
		if got.FiveHour == nil || got.SevenDay == nil {
			t.Fatalf("expected both windows present, got %+v", got)
		}
		if got.FiveHour.RemainingPercent != 76.5 {
			t.Errorf("FiveHour.RemainingPercent = %v, want 76.5", got.FiveHour.RemainingPercent)
		}
		if got.SevenDay.RemainingPercent != 58.8 {
			t.Errorf("SevenDay.RemainingPercent = %v, want 58.8", got.SevenDay.RemainingPercent)
		}
		if !got.FiveHour.ResetsAt.Equal(time.Unix(1738425600, 0)) {
			t.Errorf("FiveHour.ResetsAt = %v, want %v", got.FiveHour.ResetsAt, time.Unix(1738425600, 0))
		}
		if !got.SevenDay.ResetsAt.Equal(time.Unix(1738857600, 0)) {
			t.Errorf("SevenDay.ResetsAt = %v, want %v", got.SevenDay.ResetsAt, time.Unix(1738857600, 0))
		}
	})

	t.Run("only 5h present", func(t *testing.T) {
		t.Parallel()
		got := UsageFromRateLimits(&RateLimits{
			FiveHour: &RateLimitWindow{UsedPercentage: 10, ResetsAt: 1738425600},
		})
		if got == nil || got.FiveHour == nil {
			t.Fatalf("expected 5h window, got %+v", got)
		}
		if got.SevenDay != nil {
			t.Errorf("SevenDay = %+v, want nil (absent window must not be fabricated)", got.SevenDay)
		}
	})

	t.Run("only 7d present", func(t *testing.T) {
		t.Parallel()
		got := UsageFromRateLimits(&RateLimits{
			SevenDay: &RateLimitWindow{UsedPercentage: 90, ResetsAt: 1738857600},
		})
		if got == nil || got.SevenDay == nil {
			t.Fatalf("expected 7d window, got %+v", got)
		}
		if got.FiveHour != nil {
			t.Errorf("FiveHour = %+v, want nil", got.FiveHour)
		}
	})

	t.Run("clamps out-of-range percentages", func(t *testing.T) {
		t.Parallel()
		got := UsageFromRateLimits(&RateLimits{
			FiveHour: &RateLimitWindow{UsedPercentage: 150}, // -> remaining 0
			SevenDay: &RateLimitWindow{UsedPercentage: -20}, // -> remaining 100
		})
		if got == nil || got.FiveHour == nil || got.SevenDay == nil {
			t.Fatalf("expected both windows, got %+v", got)
		}
		if got.FiveHour.RemainingPercent != 0 {
			t.Errorf("FiveHour.RemainingPercent = %v, want 0 (clamped from used 150)", got.FiveHour.RemainingPercent)
		}
		if got.SevenDay.RemainingPercent != 100 {
			t.Errorf("SevenDay.RemainingPercent = %v, want 100 (clamped from used -20)", got.SevenDay.RemainingPercent)
		}
	})

	t.Run("zero used is a valid 100% remaining window, not absent", func(t *testing.T) {
		t.Parallel()
		got := UsageFromRateLimits(&RateLimits{
			FiveHour: &RateLimitWindow{UsedPercentage: 0, ResetsAt: 1738425600},
		})
		if got == nil || got.FiveHour == nil {
			t.Fatalf("expected 5h window present, got %+v", got)
		}
		if got.FiveHour.RemainingPercent != 100 {
			t.Errorf("RemainingPercent = %v, want 100", got.FiveHour.RemainingPercent)
		}
	})
}
