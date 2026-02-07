package internal

import "testing"

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

func TestToLower(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "already lowercase",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "all uppercase",
			input: "HELLO WORLD",
			want:  "hello world",
		},
		{
			name:  "mixed case",
			input: "HeLLo WoRLd",
			want:  "hello world",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "single character lowercase",
			input: "a",
			want:  "a",
		},
		{
			name:  "single character uppercase",
			input: "Z",
			want:  "z",
		},
		{
			name:  "numbers and symbols unchanged",
			input: "ABC123!@#",
			want:  "abc123!@#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toLower(tt.input)
			if got != tt.want {
				t.Errorf("toLower(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		sub  string
		want bool
	}{
		{
			name: "found at start",
			s:    "hello world",
			sub:  "hello",
			want: true,
		},
		{
			name: "found in middle",
			s:    "hello world",
			sub:  "lo wo",
			want: true,
		},
		{
			name: "found at end",
			s:    "hello world",
			sub:  "world",
			want: true,
		},
		{
			name: "not found",
			s:    "hello world",
			sub:  "golang",
			want: false,
		},
		{
			name: "empty substring returns true",
			s:    "hello",
			sub:  "",
			want: true,
		},
		{
			name: "empty string with empty sub",
			s:    "",
			sub:  "",
			want: true,
		},
		{
			name: "substring longer than string",
			s:    "hi",
			sub:  "hello",
			want: false,
		},
		{
			name: "exact match",
			s:    "exact",
			sub:  "exact",
			want: true,
		},
		{
			name: "single character found",
			s:    "hello",
			sub:  "e",
			want: true,
		},
		{
			name: "single character not found",
			s:    "hello",
			sub:  "x",
			want: false,
		},
		{
			name: "case sensitive - different case",
			s:    "Hello",
			sub:  "hello",
			want: false,
		},
		{
			name: "empty string with non-empty sub",
			s:    "",
			sub:  "test",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.s, tt.sub)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.sub, got, tt.want)
			}
		})
	}
}
