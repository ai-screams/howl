package internal

import (
	"encoding/json"
	"testing"
)

// fullSchemaJSON mirrors the documented Claude Code 2.1 statusline schema
// (https://code.claude.com/docs/en/statusline) with every optional field present.
const fullSchemaJSON = `{
  "cwd": "/work/proj",
  "session_id": "abc123",
  "session_name": "my-session",
  "transcript_path": "/t.jsonl",
  "model": {"id": "claude-opus-4-8", "display_name": "Opus"},
  "workspace": {
    "current_dir": "/work/proj",
    "project_dir": "/work/proj",
    "added_dirs": ["/work/extra"],
    "git_worktree": "feature-xyz",
    "repo": {"host": "github.com", "owner": "ai-screams", "name": "howl"}
  },
  "version": "2.1.177",
  "output_style": {"name": "default"},
  "cost": {"total_cost_usd": 0.01, "total_duration_ms": 45000, "total_api_duration_ms": 2300, "total_lines_added": 5, "total_lines_removed": 1},
  "context_window": {"total_input_tokens": 15500, "total_output_tokens": 1200, "context_window_size": 200000, "used_percentage": 8, "remaining_percentage": 92, "current_usage": {"input_tokens": 8500, "output_tokens": 1200, "cache_creation_input_tokens": 5000, "cache_read_input_tokens": 2000}},
  "exceeds_200k_tokens": false,
  "effort": {"level": "high"},
  "thinking": {"enabled": true},
  "rate_limits": {"five_hour": {"used_percentage": 23.5, "resets_at": 1738425600}, "seven_day": {"used_percentage": 41.2, "resets_at": 1738857600}},
  "vim": {"mode": "NORMAL"},
  "agent": {"name": "security-reviewer"},
  "pr": {"number": 1234, "url": "https://github.com/ai-screams/howl/pull/1234", "review_state": "pending"},
  "worktree": {"name": "my-feature", "path": "/wt/my-feature", "branch": "worktree-my-feature", "original_cwd": "/work/proj", "original_branch": "main"}
}`

func TestStdinDataFullSchemaDecode(t *testing.T) {
	t.Parallel()

	var d StdinData
	if err := json.Unmarshal([]byte(fullSchemaJSON), &d); err != nil {
		t.Fatalf("decode full schema: %v", err)
	}

	if d.SessionName != "my-session" {
		t.Errorf("SessionName = %q, want my-session", d.SessionName)
	}
	if d.Workspace.GitWorktree != "feature-xyz" {
		t.Errorf("Workspace.GitWorktree = %q, want feature-xyz", d.Workspace.GitWorktree)
	}
	if len(d.Workspace.AddedDirs) != 1 || d.Workspace.AddedDirs[0] != "/work/extra" {
		t.Errorf("Workspace.AddedDirs = %v, want [/work/extra]", d.Workspace.AddedDirs)
	}
	if d.Workspace.Repo == nil || d.Workspace.Repo.Name != "howl" {
		t.Errorf("Workspace.Repo = %+v, want name=howl", d.Workspace.Repo)
	}
	if d.Effort == nil || d.Effort.Level != "high" {
		t.Errorf("Effort = %+v, want level=high", d.Effort)
	}
	if d.Thinking == nil || !d.Thinking.Enabled {
		t.Errorf("Thinking = %+v, want enabled=true", d.Thinking)
	}
	if d.RateLimits == nil || d.RateLimits.FiveHour == nil || d.RateLimits.FiveHour.UsedPercentage != 23.5 {
		t.Errorf("RateLimits.FiveHour = %+v, want used=23.5", d.RateLimits)
	}
	if d.RateLimits.FiveHour.ResetsAt != 1738425600 {
		t.Errorf("RateLimits.FiveHour.ResetsAt = %d, want 1738425600", d.RateLimits.FiveHour.ResetsAt)
	}
	if d.PR == nil || d.PR.Number != 1234 || d.PR.ReviewState != "pending" {
		t.Errorf("PR = %+v, want number=1234 review_state=pending", d.PR)
	}
	if d.Worktree == nil || d.Worktree.Name != "my-feature" || d.Worktree.OriginalBranch != "main" {
		t.Errorf("Worktree = %+v, want name=my-feature original_branch=main", d.Worktree)
	}
}

// TestStdinDataOptionalAbsence asserts that missing top-level objects and missing
// subfields decode safely (nil pointers / zero strings), so renderers can omit them.
func TestStdinDataOptionalAbsence(t *testing.T) {
	t.Parallel()

	t.Run("minimal payload leaves optionals nil", func(t *testing.T) {
		t.Parallel()
		var d StdinData
		if err := json.Unmarshal([]byte(`{"session_id":"x","model":{"display_name":"Opus"}}`), &d); err != nil {
			t.Fatalf("decode minimal: %v", err)
		}
		for name, isNil := range map[string]bool{
			"Effort":     d.Effort == nil,
			"Thinking":   d.Thinking == nil,
			"RateLimits": d.RateLimits == nil,
			"PR":         d.PR == nil,
			"Worktree":   d.Worktree == nil,
			"Repo":       d.Workspace.Repo == nil,
		} {
			if !isNil {
				t.Errorf("%s should be nil when absent", name)
			}
		}
	})

	t.Run("present objects with absent subfields", func(t *testing.T) {
		t.Parallel()
		var d StdinData
		// pr without review_state; worktree without branch/original_branch
		js := `{"pr":{"number":7},"worktree":{"name":"wt","path":"/p"}}`
		if err := json.Unmarshal([]byte(js), &d); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if d.PR == nil || d.PR.Number != 7 || d.PR.ReviewState != "" {
			t.Errorf("PR = %+v, want number=7 review_state empty", d.PR)
		}
		if d.Worktree == nil || d.Worktree.Branch != "" || d.Worktree.OriginalBranch != "" {
			t.Errorf("Worktree = %+v, want empty branch fields", d.Worktree)
		}
	})

	t.Run("rate_limits with only one window", func(t *testing.T) {
		t.Parallel()
		var d StdinData
		if err := json.Unmarshal([]byte(`{"rate_limits":{"five_hour":{"used_percentage":10,"resets_at":1}}}`), &d); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if d.RateLimits == nil || d.RateLimits.FiveHour == nil {
			t.Fatalf("expected five_hour window, got %+v", d.RateLimits)
		}
		if d.RateLimits.SevenDay != nil {
			t.Errorf("SevenDay = %+v, want nil when absent", d.RateLimits.SevenDay)
		}
	})
}

func TestClassifyModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		model Model
		want  ModelTier
	}{
		{
			name:  "opus via DisplayName",
			model: Model{DisplayName: "Claude Opus 4"},
			want:  TierOpus,
		},
		{
			name:  "sonnet via DisplayName",
			model: Model{DisplayName: "Claude Sonnet 4.5"},
			want:  TierSonnet,
		},
		{
			name:  "haiku via DisplayName",
			model: Model{DisplayName: "Claude Haiku 4.5"},
			want:  TierHaiku,
		},
		{
			name:  "unknown model",
			model: Model{DisplayName: "GPT-4"},
			want:  TierUnknown,
		},
		{
			name:  "empty DisplayName falls back to ID with opus",
			model: Model{ID: "claude-opus-4", DisplayName: ""},
			want:  TierOpus,
		},
		{
			name:  "empty DisplayName falls back to ID with sonnet",
			model: Model{ID: "claude-sonnet-3-5", DisplayName: ""},
			want:  TierSonnet,
		},
		{
			name:  "empty DisplayName falls back to ID with haiku",
			model: Model{ID: "claude-haiku-3", DisplayName: ""},
			want:  TierHaiku,
		},
		{
			name:  "both empty returns TierUnknown",
			model: Model{ID: "", DisplayName: ""},
			want:  TierUnknown,
		},
		{
			name:  "case insensitive OPUS uppercase",
			model: Model{DisplayName: "Claude OPUS 4"},
			want:  TierOpus,
		},
		{
			name:  "case insensitive opus lowercase",
			model: Model{DisplayName: "claude opus 4"},
			want:  TierOpus,
		},
		{
			name:  "bedrock-style ID with sonnet",
			model: Model{ID: "anthropic.claude-sonnet-3-5-v2:0"},
			want:  TierSonnet,
		},
		{
			name:  "mixed case in ID",
			model: Model{ID: "Claude-HAIKU-3"},
			want:  TierHaiku,
		},
		{
			name:  "opus takes precedence over sonnet",
			model: Model{DisplayName: "Claude Opus Sonnet 4"},
			want:  TierOpus,
		},
		{
			name:  "sonnet takes precedence over haiku",
			model: Model{DisplayName: "Claude Sonnet Haiku 4"},
			want:  TierSonnet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyModel(tt.model)
			if got != tt.want {
				t.Errorf("classifyModel(%v) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}
