package internal

import "strings"

// StdinData represents the top-level JSON structure piped by Claude Code every ~300ms.
// It contains session metrics, context window usage, cost data, and workspace information.
// Field set tracks the Claude Code 2.1 statusline schema (https://code.claude.com/docs/en/statusline).
type StdinData struct {
	SessionID         string        `json:"session_id"`
	SessionName       string        `json:"session_name"`
	TranscriptPath    string        `json:"transcript_path"`
	CWD               string        `json:"cwd"`
	Version           string        `json:"version"`
	Model             Model         `json:"model"`
	Workspace         Workspace     `json:"workspace"`
	Cost              Cost          `json:"cost"`
	ContextWindow     ContextWindow `json:"context_window"`
	Exceeds200KTokens bool          `json:"exceeds_200k_tokens"`
	OutputStyle       *OutputStyle  `json:"output_style"`
	Vim               *Vim          `json:"vim"`
	Agent             *Agent        `json:"agent"`
	Effort            *Effort       `json:"effort"`
	Thinking          *Thinking     `json:"thinking"`
	RateLimits        *RateLimits   `json:"rate_limits"`
	PR                *PullRequest  `json:"pr"`
	Worktree          *Worktree     `json:"worktree"`
}

// Model represents the AI model being used for the session.
type Model struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// Workspace represents the working directory context for the session.
type Workspace struct {
	CurrentDir  string    `json:"current_dir"`
	ProjectDir  string    `json:"project_dir"`
	AddedDirs   []string  `json:"added_dirs"`
	GitWorktree string    `json:"git_worktree"`
	Repo        *RepoInfo `json:"repo"`
}

// RepoInfo identifies the git repository hosting the workspace, when an origin
// remote is present.
type RepoInfo struct {
	Host  string `json:"host"`
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

// Cost represents the cumulative session cost and duration metrics.
type Cost struct {
	TotalCostUSD       float64 `json:"total_cost_usd"`
	TotalDurationMS    int64   `json:"total_duration_ms"`
	TotalAPIDurationMS int64   `json:"total_api_duration_ms"`
	TotalLinesAdded    int     `json:"total_lines_added"`
	TotalLinesRemoved  int     `json:"total_lines_removed"`
}

// ContextWindow represents the context window usage and limits for the current session.
type ContextWindow struct {
	TotalInputTokens    int           `json:"total_input_tokens"`
	TotalOutputTokens   int           `json:"total_output_tokens"`
	ContextWindowSize   int           `json:"context_window_size"`
	UsedPercentage      *float64      `json:"used_percentage"`
	RemainingPercentage *float64      `json:"remaining_percentage"`
	CurrentUsage        *CurrentUsage `json:"current_usage"`
}

// CurrentUsage represents the token breakdown for the current API call.
type CurrentUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// OutputStyle represents the current output style configuration.
type OutputStyle struct {
	Name string `json:"name"`
}

// Vim represents the current vim mode if enabled in Claude Code.
type Vim struct {
	Mode string `json:"mode"`
}

// Agent represents the active agent teammate if in team mode.
type Agent struct {
	Name string `json:"name"`
}

// Effort represents the reasoning effort level, present when the model supports
// the reasoning effort parameter.
type Effort struct {
	Level string `json:"level"` // e.g. "high"
}

// Thinking represents whether extended thinking is enabled.
type Thinking struct {
	Enabled bool `json:"enabled"`
}

// PullRequest represents an open pull request associated with the branch.
type PullRequest struct {
	Number      int    `json:"number"`
	URL         string `json:"url"`
	ReviewState string `json:"review_state"` // e.g. "pending"; may be absent
}

// Worktree represents an active --worktree session.
type Worktree struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	Branch         string `json:"branch"` // may be absent for hook-based worktrees
	OriginalCWD    string `json:"original_cwd"`
	OriginalBranch string `json:"original_branch"` // may be absent
}

// RateLimits holds Claude.ai subscription rate-limit usage. Present only for
// subscribers, after the first API response. Each window can be independently absent.
type RateLimits struct {
	FiveHour *RateLimitWindow `json:"five_hour"`
	SevenDay *RateLimitWindow `json:"seven_day"`
}

// RateLimitWindow is a single rate-limit window. UsedPercentage is 0-100;
// ResetsAt is Unix epoch seconds.
type RateLimitWindow struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64   `json:"resets_at"`
}

// RenderContext bundles all inputs for Render. Optional sources (Git, Usage,
// Tools, Account) are nil-safe — Render skips them when nil.
type RenderContext struct {
	Data    *StdinData
	Metrics Metrics
	Git     *GitInfo
	Usage   *UsageData
	Tools   *ToolInfo
	Account *AccountInfo
	Config  Config
}

// ModelTier classifies a model by its performance/cost tier.
type ModelTier int

const (
	TierUnknown ModelTier = iota
	TierHaiku
	TierSonnet
	TierOpus
)

func classifyModel(m Model) ModelTier {
	name := m.DisplayName
	if name == "" {
		name = m.ID
	}
	lower := strings.ToLower(name)
	if strings.Contains(lower, "opus") {
		return TierOpus
	}
	if strings.Contains(lower, "sonnet") {
		return TierSonnet
	}
	if strings.Contains(lower, "haiku") {
		return TierHaiku
	}
	return TierUnknown
}
