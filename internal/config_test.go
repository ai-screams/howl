package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Preset != "full" {
		t.Errorf("expected full preset, got %s", cfg.Preset)
	}
	if !cfg.Features.Account {
		t.Errorf("full preset should have account enabled")
	}
	if !cfg.Features.Git {
		t.Errorf("full preset should have git enabled")
	}
	if !cfg.Features.Tools {
		t.Errorf("full preset should have tools enabled")
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	// Use non-existent path
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := LoadConfig()
	if cfg.Preset != "full" {
		t.Errorf("no config file should return full preset, got %s", cfg.Preset)
	}
	if !cfg.Features.Account {
		t.Errorf("default should have account enabled")
	}
}

func TestLoadConfig_MalformedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".claude", "hud")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.json")

	// Write invalid JSON
	os.WriteFile(configPath, []byte("{invalid json"), 0644)

	cfg := LoadConfig()
	if cfg.Preset != "full" {
		t.Errorf("malformed JSON should fallback to full, got %s", cfg.Preset)
	}
}

func TestLoadConfig_EmptyPreset(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".claude", "hud")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.json")

	// Write empty preset
	os.WriteFile(configPath, []byte(`{"preset":""}`), 0644)

	cfg := LoadConfig()
	if cfg.Preset != "full" {
		t.Errorf("empty preset should normalize to full, got %s", cfg.Preset)
	}
}

func TestLoadConfig_UnknownPreset(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".claude", "hud")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.json")

	// Write unknown preset
	os.WriteFile(configPath, []byte(`{"preset":"unknown"}`), 0644)

	cfg := LoadConfig()
	if cfg.Preset != "full" {
		t.Errorf("unknown preset should fallback to full, got %s", cfg.Preset)
	}
	if !cfg.Features.Account {
		t.Errorf("fallback should have full features")
	}
}

func TestLoadConfig_ValidPresets(t *testing.T) {
	tests := []struct {
		preset        string
		expectAccount bool
		expectGit     bool
		expectTools   bool
		expectQuota   bool
		expectCostVel bool
	}{
		{"full", true, true, true, true, true},
		{"minimal", false, false, false, false, false},
		{"developer", true, true, false, false, false},
		{"cost-focused", false, false, false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.preset, func(t *testing.T) {
			tmpDir := t.TempDir()
			t.Setenv("HOME", tmpDir)

			configDir := filepath.Join(tmpDir, ".claude", "hud")
			os.MkdirAll(configDir, 0755)
			configPath := filepath.Join(configDir, "config.json")

			content := `{"preset":"` + tt.preset + `"}`
			os.WriteFile(configPath, []byte(content), 0644)

			cfg := LoadConfig()
			if cfg.Preset != tt.preset {
				t.Errorf("expected %s, got %s", tt.preset, cfg.Preset)
			}
			if cfg.Features.Account != tt.expectAccount {
				t.Errorf("%s: account should be %v", tt.preset, tt.expectAccount)
			}
			if cfg.Features.Git != tt.expectGit {
				t.Errorf("%s: git should be %v", tt.preset, tt.expectGit)
			}
			if cfg.Features.Tools != tt.expectTools {
				t.Errorf("%s: tools should be %v", tt.preset, tt.expectTools)
			}
			if cfg.Features.Quota != tt.expectQuota {
				t.Errorf("%s: quota should be %v", tt.preset, tt.expectQuota)
			}
			if cfg.Features.CostVelocity != tt.expectCostVel {
				t.Errorf("%s: cost_velocity should be %v", tt.preset, tt.expectCostVel)
			}
		})
	}
}

func TestLoadConfig_Normalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`{"preset":"Full "}`, "full"},
		{`{"preset":"MINIMAL"}`, "minimal"},
		{`{"preset":"  Developer  "}`, "developer"},
		{`{"preset":"COST-FOCUSED"}`, "cost-focused"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tmpDir := t.TempDir()
			t.Setenv("HOME", tmpDir)

			configDir := filepath.Join(tmpDir, ".claude", "hud")
			os.MkdirAll(configDir, 0755)
			configPath := filepath.Join(configDir, "config.json")

			os.WriteFile(configPath, []byte(tt.input), 0644)

			cfg := LoadConfig()
			if cfg.Preset != tt.expected {
				t.Errorf("expected normalized %s, got %s", tt.expected, cfg.Preset)
			}
		})
	}
}

func TestLoadConfig_FileTooLarge(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".claude", "hud")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.json")

	// Create file >4KB
	largeContent := strings.Repeat("x", maxConfigSize+1)
	os.WriteFile(configPath, []byte(largeContent), 0644)

	cfg := LoadConfig()
	if cfg.Preset != "full" {
		t.Errorf("oversized file should fallback to full, got %s", cfg.Preset)
	}
}

func TestLoadConfig_FeaturesIgnored(t *testing.T) {
	// v1.1: features field now overrides preset base
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".claude", "hud")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.json")

	// Enable account in minimal preset via features field
	content := `{"preset":"minimal","features":{"account":true}}`
	os.WriteFile(configPath, []byte(content), 0644)

	cfg := LoadConfig()
	if cfg.Preset != "minimal" {
		t.Errorf("expected minimal, got %s", cfg.Preset)
	}
	// v1.1: features.account=true should enable account (override minimal base)
	if !cfg.Features.Account {
		t.Errorf("v1.1 should merge features field, account should be true")
	}
}

// --- v1.1 Tests: mergeFeatures + validatePriority ---

func TestMergeFeatures_AllFalse(t *testing.T) {
	base := FeatureToggles{Account: true, Git: true, Tools: true}
	override := FeatureToggles{} // all false

	result := mergeFeatures(base, override)
	if !result.Account || !result.Git || !result.Tools {
		t.Errorf("base should be preserved when override is all false")
	}
}

func TestMergeFeatures_Override(t *testing.T) {
	base := FeatureToggles{Account: false, Git: false}
	override := FeatureToggles{Account: true, Quota: true}

	result := mergeFeatures(base, override)
	if !result.Account {
		t.Errorf("override.Account=true should set result.Account=true")
	}
	if !result.Quota {
		t.Errorf("override.Quota=true should set result.Quota=true")
	}
	if result.Git {
		t.Errorf("base.Git=false + override.Git=false should keep result.Git=false")
	}
}

func TestMergeFeatures_AllTrue(t *testing.T) {
	base := FeatureToggles{Account: false}
	override := FeatureToggles{
		Account: true, Git: true, LineChanges: true, ResponseSpeed: true,
		Quota: true, Tools: true, Agents: true, CacheEfficiency: true,
		APIWaitRatio: true, CostVelocity: true, VimMode: true, AgentName: true,
	}

	result := mergeFeatures(base, override)
	if !result.Account || !result.Git || !result.Quota || !result.Tools {
		t.Errorf("all override fields should be true in result")
	}
}

func TestValidatePriority_Empty(t *testing.T) {
	result := validatePriority([]string{})
	if result != nil {
		t.Errorf("empty input should return nil, got %v", result)
	}
}

func TestValidatePriority_Duplicates(t *testing.T) {
	input := []string{"account", "git", "account", "quota"}
	result := validatePriority(input)
	if len(result) != 3 {
		t.Errorf("expected 3 unique items, got %d: %v", len(result), result)
	}
	// Check account appears only once
	count := 0
	for _, m := range result {
		if m == "account" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected account to appear once, got %d", count)
	}
}

func TestValidatePriority_Normalization(t *testing.T) {
	input := []string{"Account", "GIT ", " line_changes", "RESPONSE_SPEED"}
	result := validatePriority(input)
	if len(result) != 4 {
		t.Errorf("expected 4 items, got %d: %v", len(result), result)
	}
	for _, m := range result {
		if m != strings.ToLower(strings.TrimSpace(m)) {
			t.Errorf("metric %s should be normalized", m)
		}
	}
}

func TestValidatePriority_InvalidMetrics(t *testing.T) {
	input := []string{"account", "tools", "cache_efficiency", "git"}
	result := validatePriority(input)
	// tools and cache_efficiency are Line 3/4, should be filtered out
	if len(result) != 2 {
		t.Errorf("expected 2 valid Line 2 metrics, got %d: %v", len(result), result)
	}
	for _, m := range result {
		if !validLine2Metrics[m] {
			t.Errorf("result should only contain Line 2 metrics, got %s", m)
		}
	}
}

func TestValidatePriority_MaxFive(t *testing.T) {
	input := []string{"account", "git", "line_changes", "response_speed", "quota", "extra1", "extra2"}
	result := validatePriority(input)
	if len(result) != 5 {
		t.Errorf("expected max 5 items, got %d: %v", len(result), result)
	}
}

func TestLoadConfig_WithFeaturesOverride(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".claude", "hud")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.json")

	// minimal preset + quota override
	content := `{"preset":"minimal","features":{"quota":true,"git":true}}`
	os.WriteFile(configPath, []byte(content), 0644)

	cfg := LoadConfig()
	if cfg.Preset != "minimal" {
		t.Errorf("expected minimal, got %s", cfg.Preset)
	}
	if !cfg.Features.Quota {
		t.Errorf("features.quota=true should enable quota")
	}
	if !cfg.Features.Git {
		t.Errorf("features.git=true should enable git")
	}
	if cfg.Features.Account {
		t.Errorf("minimal base should keep account=false")
	}
}

func TestLoadConfig_WithPriority(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".claude", "hud")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.json")

	content := `{"preset":"developer","priority":["quota","git","account"]}`
	os.WriteFile(configPath, []byte(content), 0644)

	cfg := LoadConfig()
	if len(cfg.Priority) != 3 {
		t.Errorf("expected 3 priority items, got %d: %v", len(cfg.Priority), cfg.Priority)
	}
	if cfg.Priority[0] != "quota" || cfg.Priority[1] != "git" || cfg.Priority[2] != "account" {
		t.Errorf("priority order should be [quota, git, account], got %v", cfg.Priority)
	}
}

func TestLoadConfig_PriorityWithInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".claude", "hud")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.json")

	// Include invalid metrics (tools, cache_efficiency are Line 3/4)
	content := `{"preset":"full","priority":["quota","tools","git","cache_efficiency","account"]}`
	os.WriteFile(configPath, []byte(content), 0644)

	cfg := LoadConfig()
	// Only quota, git, account should remain
	if len(cfg.Priority) != 3 {
		t.Errorf("expected 3 valid items (invalid filtered), got %d: %v", len(cfg.Priority), cfg.Priority)
	}
	for _, m := range cfg.Priority {
		if !validLine2Metrics[m] {
			t.Errorf("priority should only contain Line 2 metrics, got %s", m)
		}
	}
}

// --- v1.5 Tests: Thresholds ---

func TestDefaultThresholds(t *testing.T) {
	th := DefaultThresholds()

	// Verify all 16 fields match constants
	if th.ContextDanger != DangerThreshold {
		t.Errorf("ContextDanger: expected %d, got %d", DangerThreshold, th.ContextDanger)
	}
	if th.ContextWarning != WarningThreshold {
		t.Errorf("ContextWarning: expected %d, got %d", WarningThreshold, th.ContextWarning)
	}
	if th.ContextModerate != ModerateThreshold {
		t.Errorf("ContextModerate: expected %d, got %d", ModerateThreshold, th.ContextModerate)
	}
	if th.SessionCostHigh != SessionCostHigh {
		t.Errorf("SessionCostHigh: expected %.2f, got %.2f", SessionCostHigh, th.SessionCostHigh)
	}
	if th.SessionCostMed != SessionCostMedium {
		t.Errorf("SessionCostMed: expected %.2f, got %.2f", SessionCostMedium, th.SessionCostMed)
	}
	if th.CacheExcellent != CacheExcellent {
		t.Errorf("CacheExcellent: expected %d, got %d", CacheExcellent, th.CacheExcellent)
	}
	if th.CacheGood != CacheGood {
		t.Errorf("CacheGood: expected %d, got %d", CacheGood, th.CacheGood)
	}
	if th.WaitHigh != WaitHigh {
		t.Errorf("WaitHigh: expected %d, got %d", WaitHigh, th.WaitHigh)
	}
	if th.WaitMedium != WaitMedium {
		t.Errorf("WaitMedium: expected %d, got %d", WaitMedium, th.WaitMedium)
	}
	if th.SpeedFast != SpeedFast {
		t.Errorf("SpeedFast: expected %d, got %d", SpeedFast, th.SpeedFast)
	}
	if th.SpeedModerate != SpeedModerate {
		t.Errorf("SpeedModerate: expected %d, got %d", SpeedModerate, th.SpeedModerate)
	}
	if th.CostVelocityHigh != CostHigh {
		t.Errorf("CostVelocityHigh: expected %.2f, got %.2f", CostHigh, th.CostVelocityHigh)
	}
	if th.CostVelocityMed != CostMedium {
		t.Errorf("CostVelocityMed: expected %.2f, got %.2f", CostMedium, th.CostVelocityMed)
	}
	if th.QuotaCritical != QuotaCritical {
		t.Errorf("QuotaCritical: expected %.1f, got %.1f", float64(QuotaCritical), th.QuotaCritical)
	}
	if th.QuotaLow != QuotaLow {
		t.Errorf("QuotaLow: expected %.1f, got %.1f", float64(QuotaLow), th.QuotaLow)
	}
	if th.QuotaMedium != QuotaMedium {
		t.Errorf("QuotaMedium: expected %.1f, got %.1f", float64(QuotaMedium), th.QuotaMedium)
	}
	if th.QuotaHigh != QuotaHigh {
		t.Errorf("QuotaHigh: expected %.1f, got %.1f", float64(QuotaHigh), th.QuotaHigh)
	}
}

func TestMergeThresholds_AllZero(t *testing.T) {
	base := DefaultThresholds()
	override := Thresholds{} // all zero

	result := mergeThresholds(base, override)

	// All fields should remain at default
	if result.ContextDanger != DangerThreshold {
		t.Errorf("ContextDanger should be preserved: expected %d, got %d", DangerThreshold, result.ContextDanger)
	}
	if result.SpeedFast != SpeedFast {
		t.Errorf("SpeedFast should be preserved: expected %d, got %d", SpeedFast, result.SpeedFast)
	}
	if result.QuotaCritical != QuotaCritical {
		t.Errorf("QuotaCritical should be preserved: expected %.1f, got %.1f", float64(QuotaCritical), result.QuotaCritical)
	}
}

func TestMergeThresholds_PartialOverride(t *testing.T) {
	base := DefaultThresholds()
	override := Thresholds{
		ContextDanger:   90,
		SpeedFast:       100,
		SessionCostHigh: 10.0,
		CostVelocityMed: 0.25,
		QuotaCritical:   15.0,
	}

	result := mergeThresholds(base, override)

	// Overridden int fields
	if result.ContextDanger != 90 {
		t.Errorf("ContextDanger should be overridden: expected 90, got %d", result.ContextDanger)
	}
	if result.SpeedFast != 100 {
		t.Errorf("SpeedFast should be overridden: expected 100, got %d", result.SpeedFast)
	}

	// Overridden float64 fields
	if result.SessionCostHigh != 10.0 {
		t.Errorf("SessionCostHigh should be overridden: expected 10.0, got %.2f", result.SessionCostHigh)
	}
	if result.CostVelocityMed != 0.25 {
		t.Errorf("CostVelocityMed should be overridden: expected 0.25, got %.2f", result.CostVelocityMed)
	}
	if result.QuotaCritical != 15.0 {
		t.Errorf("QuotaCritical should be overridden: expected 15.0, got %.1f", result.QuotaCritical)
	}

	// Preserved defaults
	if result.ContextWarning != WarningThreshold {
		t.Errorf("ContextWarning should remain default: expected %d, got %d", WarningThreshold, result.ContextWarning)
	}
	if result.CacheExcellent != CacheExcellent {
		t.Errorf("CacheExcellent should remain default: expected %d, got %d", CacheExcellent, result.CacheExcellent)
	}
	if result.QuotaLow != QuotaLow {
		t.Errorf("QuotaLow should remain default: expected %.1f, got %.1f", float64(QuotaLow), result.QuotaLow)
	}
}

func TestMergeThresholds_NegativeIgnored(t *testing.T) {
	base := DefaultThresholds()
	override := Thresholds{
		ContextDanger: -1,
		CacheGood:     -50,
		QuotaCritical: -10.0,
	}

	result := mergeThresholds(base, override)

	// Negative values should be ignored, base preserved
	if result.ContextDanger != DangerThreshold {
		t.Errorf("negative ContextDanger should be ignored: expected %d, got %d", DangerThreshold, result.ContextDanger)
	}
	if result.CacheGood != CacheGood {
		t.Errorf("negative CacheGood should be ignored: expected %d, got %d", CacheGood, result.CacheGood)
	}
	if result.QuotaCritical != QuotaCritical {
		t.Errorf("negative QuotaCritical should be ignored: expected %.1f, got %.1f", float64(QuotaCritical), result.QuotaCritical)
	}
}

func TestLoadConfig_WithThresholds(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".claude", "hud")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.json")

	// Write config with custom thresholds
	content := `{"preset":"full","thresholds":{"context_danger":90,"speed_fast":100}}`
	os.WriteFile(configPath, []byte(content), 0644)

	cfg := LoadConfig()

	// Verify overridden values
	if cfg.Thresholds.ContextDanger != 90 {
		t.Errorf("ContextDanger should be 90, got %d", cfg.Thresholds.ContextDanger)
	}
	if cfg.Thresholds.SpeedFast != 100 {
		t.Errorf("SpeedFast should be 100, got %d", cfg.Thresholds.SpeedFast)
	}

	// Verify default values preserved
	if cfg.Thresholds.ContextWarning != WarningThreshold {
		t.Errorf("ContextWarning should remain default %d, got %d", WarningThreshold, cfg.Thresholds.ContextWarning)
	}
	if cfg.Thresholds.CacheExcellent != CacheExcellent {
		t.Errorf("CacheExcellent should remain default %d, got %d", CacheExcellent, cfg.Thresholds.CacheExcellent)
	}
	if cfg.Thresholds.QuotaLow != QuotaLow {
		t.Errorf("QuotaLow should remain default %.1f, got %.1f", float64(QuotaLow), cfg.Thresholds.QuotaLow)
	}
}

func TestValidateThresholds_InvertedContext(t *testing.T) {
	th := Thresholds{
		ContextDanger:   50,
		ContextWarning:  90,
		ContextModerate: 95,
	}
	validateThresholds(&th)

	if th.ContextWarning >= th.ContextDanger {
		t.Errorf("warning (%d) should be < danger (%d)", th.ContextWarning, th.ContextDanger)
	}
	if th.ContextModerate >= th.ContextWarning {
		t.Errorf("moderate (%d) should be < warning (%d)", th.ContextModerate, th.ContextWarning)
	}
}

func TestValidateThresholds_InvertedCost(t *testing.T) {
	th := Thresholds{
		SessionCostHigh:  1.0,
		SessionCostMed:   5.0,
		CostVelocityHigh: 0.10,
		CostVelocityMed:  0.50,
	}
	validateThresholds(&th)

	if th.SessionCostMed >= th.SessionCostHigh {
		t.Errorf("session cost med (%.2f) should be < high (%.2f)", th.SessionCostMed, th.SessionCostHigh)
	}
	if th.CostVelocityMed >= th.CostVelocityHigh {
		t.Errorf("cost velocity med (%.2f) should be < high (%.2f)", th.CostVelocityMed, th.CostVelocityHigh)
	}
}

func TestValidateThresholds_InvertedQuota(t *testing.T) {
	// Quota ordering: critical < low < medium < high
	th := Thresholds{
		QuotaCritical: 80,
		QuotaLow:      60,
		QuotaMedium:   40,
		QuotaHigh:     20,
	}
	validateThresholds(&th)

	if th.QuotaLow <= th.QuotaCritical {
		t.Errorf("quota low (%.1f) should be > critical (%.1f)", th.QuotaLow, th.QuotaCritical)
	}
	if th.QuotaMedium <= th.QuotaLow {
		t.Errorf("quota medium (%.1f) should be > low (%.1f)", th.QuotaMedium, th.QuotaLow)
	}
	if th.QuotaHigh <= th.QuotaMedium {
		t.Errorf("quota high (%.1f) should be > medium (%.1f)", th.QuotaHigh, th.QuotaMedium)
	}
}

func TestValidateThresholds_ValidUnchanged(t *testing.T) {
	th := DefaultThresholds()
	original := th
	validateThresholds(&th)

	if th != original {
		t.Errorf("valid defaults should not be modified by validateThresholds")
	}
}

func TestLoadConfig_InvertedThresholdsCorrected(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".claude", "hud")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.json")

	// Intentionally inverted: warning > danger
	content := `{"preset":"full","thresholds":{"context_danger":50,"context_warning":90}}`
	os.WriteFile(configPath, []byte(content), 0644)

	cfg := LoadConfig()

	if cfg.Thresholds.ContextWarning >= cfg.Thresholds.ContextDanger {
		t.Errorf("LoadConfig should fix inverted thresholds: warning=%d >= danger=%d",
			cfg.Thresholds.ContextWarning, cfg.Thresholds.ContextDanger)
	}
}
