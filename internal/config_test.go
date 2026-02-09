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
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := LoadConfig()
	if cfg.Preset != "full" {
		t.Errorf("no config file should return full preset, got %s", cfg.Preset)
	}
	if !cfg.Features.Account {
		t.Errorf("default should have account enabled")
	}
}

func TestLoadConfig_MalformedJSON(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

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
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

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
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

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
			origHome := os.Getenv("HOME")
			tmpDir := t.TempDir()
			os.Setenv("HOME", tmpDir)
			defer os.Setenv("HOME", origHome)

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
			origHome := os.Getenv("HOME")
			tmpDir := t.TempDir()
			os.Setenv("HOME", tmpDir)
			defer os.Setenv("HOME", origHome)

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
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

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
	// v1.0: features field in JSON is parsed but ignored, preset overrides
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".claude", "hud")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.json")

	// Try to enable account in minimal preset via features field
	content := `{"preset":"minimal","features":{"account":true}}`
	os.WriteFile(configPath, []byte(content), 0644)

	cfg := LoadConfig()
	if cfg.Preset != "minimal" {
		t.Errorf("expected minimal, got %s", cfg.Preset)
	}
	// In v1.0, preset overrides features, so account should be false
	if cfg.Features.Account {
		t.Errorf("v1.0 should ignore features field, account should be false in minimal")
	}
}
