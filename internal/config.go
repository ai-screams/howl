package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const maxConfigSize = 4096 // 4KB limit

// Config represents user's statusline configuration.
type Config struct {
	Preset   string         `json:"preset"`
	Features FeatureToggles `json:"features"` // v1.0에서 무시, v1.1에서 활성화
}

// FeatureToggles controls which metrics are displayed.
type FeatureToggles struct {
	Account         bool `json:"account"`
	Git             bool `json:"git"`
	LineChanges     bool `json:"line_changes"`
	ResponseSpeed   bool `json:"response_speed"`
	Quota           bool `json:"quota"`
	Tools           bool `json:"tools"`
	Agents          bool `json:"agents"`
	CacheEfficiency bool `json:"cache_efficiency"`
	APIWaitRatio    bool `json:"api_wait_ratio"`
	CostVelocity    bool `json:"cost_velocity"`
	VimMode         bool `json:"vim_mode"`
	AgentName       bool `json:"agent_name"`
	// TokenBreakdown bool `json:"token_breakdown"` // Reserved for danger mode only (v1.1+)
}

var presets = map[string]FeatureToggles{
	"full": {
		Account:         true,
		Git:             true,
		LineChanges:     true,
		ResponseSpeed:   true,
		Quota:           true,
		Tools:           true,
		Agents:          true,
		CacheEfficiency: true,
		APIWaitRatio:    true,
		CostVelocity:    true,
		VimMode:         true,
		AgentName:       true,
	},
	"minimal": {}, // all false
	"developer": {
		Account:         true,
		Git:             true,
		LineChanges:     true,
		ResponseSpeed:   true,
		CacheEfficiency: true,
		VimMode:         true,
	},
	"cost-focused": {
		Quota:        true,
		APIWaitRatio: true,
		CostVelocity: true,
	},
}

// DefaultConfig returns full preset (current behavior).
func DefaultConfig() Config {
	return PresetConfig("full")
}

// PresetConfig returns a Config with the specified preset.
// Used for testing. If preset is unknown, returns full.
func PresetConfig(name string) Config {
	name = strings.ToLower(strings.TrimSpace(name))
	if features, ok := presets[name]; ok {
		return Config{Preset: name, Features: features}
	}
	return Config{Preset: "full", Features: presets["full"]}
}

// LoadConfig loads configuration from ~/.claude/hud/config.json.
// On any error (file not found, parse error, invalid preset), returns DefaultConfig().
func LoadConfig() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		return DefaultConfig()
	}

	path := filepath.Join(home, ".claude", "hud", "config.json")

	// Guard: check file size before reading
	stat, err := os.Stat(path)
	if err != nil {
		return DefaultConfig() // File not found
	}
	if stat.Size() > maxConfigSize {
		return DefaultConfig() // Too large, DoS protection
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig()
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig() // Malformed JSON
	}

	// Normalize preset name: lowercase + trim whitespace
	cfg.Preset = strings.ToLower(strings.TrimSpace(cfg.Preset))
	if cfg.Preset == "" {
		cfg.Preset = "full"
	}

	// Apply preset (v1.0: preset overrides features field)
	if p, ok := presets[cfg.Preset]; ok {
		cfg.Features = p
	} else {
		// Unknown preset -> fallback to full silently
		cfg.Preset = "full"
		cfg.Features = presets["full"]
	}

	return cfg
}
