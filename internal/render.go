package internal

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ANSI escape codes for terminal color formatting.
// Reset clears all formatting, bold/dim control intensity,
// colors follow the standard 3-bit palette (30-37),
// and extended 8-bit colors use 38;5;N format.
const (
	Reset   = "\033[0m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	boldRed = "\033[1;31m"
	boldYlw = "\033[1;33m" // gold for opus
	orange  = "\033[38;5;208m"
	grey    = "\033[38;5;245m"
)

// Render produces lines for the statusline display.
// Normal mode: 2-4 lines (depending on active features). Danger mode (configurable, default 85%+): 2 dense lines.
func Render(rc RenderContext) []string {
	if rc.Config.Thresholds == (Thresholds{}) {
		rc.Config.Thresholds = DefaultThresholds()
	}
	if rc.Metrics.ContextPercent >= rc.Config.Thresholds.ContextDanger {
		return renderDangerMode(rc)
	}
	return renderNormalMode(rc)
}

func renderNormalMode(rc RenderContext) []string {
	d, m, git, usage, tools, account, cfg := rc.Data, rc.Metrics, rc.Git, rc.Usage, rc.Tools, rc.Account, rc.Config
	t := cfg.Thresholds

	// Determine if quota bars will be shown
	hasQuotaBars := cfg.Features.Quota && usage != nil

	// Line 1: model(+context size) | account | git | speed | cost | duration
	// When no quota bars, context bar is inlined into L1 (minimal/developer compact)
	line1 := make([]string, 0, 8)
	if hasQuotaBars {
		line1 = append(line1, renderModelBadge(d.Model, d.ContextWindow.ContextWindowSize))
	} else {
		line1 = append(line1, renderModelBadge(d.Model, 0))
		line1 = append(line1, renderContextBar(m.ContextPercent, d.ContextWindow, t))
	}
	if cfg.Features.Account && account != nil && account.EmailAddress != "" {
		line1 = append(line1, renderAccount(account))
	}
	if cfg.Features.Git && git != nil && git.Branch != "" {
		line1 = append(line1, renderGitCompact(git))
	}
	if cfg.Features.OutputTokens {
		if s := renderOutputTokens(d.ContextWindow.CurrentUsage); s != "" {
			line1 = append(line1, s)
		}
	}
	if costStr := renderCost(d.Cost.TotalCostUSD, t); costStr != "" {
		line1 = append(line1, costStr)
	}
	line1 = append(line1, renderDuration(d.Cost.TotalDurationMS))

	// Line 2: context bar | 5h quota bar | 7d quota bar (only when quota bars exist)
	var line2 []string
	if hasQuotaBars {
		line2 = make([]string, 0, 3)
		line2 = append(line2, renderContextBar(m.ContextPercent, d.ContextWindow, t))
		if s := renderQuotaWindow(usage.FiveHour, "5h", t); s != "" {
			line2 = append(line2, s)
		}
		if s := renderQuotaWindow(usage.SevenDay, "7d", t); s != "" {
			line2 = append(line2, s)
		}
	}

	// Line 3: line changes | cache | wait | cost velocity (+ vim/agent conditionally)
	line3 := make([]string, 0, 7)
	if cfg.Features.LineChanges {
		if lines := renderLineChanges(d.Cost); lines != "" {
			line3 = append(line3, lines)
		}
	}
	if cfg.Features.CacheEfficiency && m.CacheEfficiency != nil && *m.CacheEfficiency > 0 {
		line3 = append(line3, renderCacheEfficiencyLabeled(*m.CacheEfficiency, t, d.ContextWindow.CurrentUsage))
	}
	if cfg.Features.APIWaitRatio && m.APIWaitRatio != nil && *m.APIWaitRatio > 0 {
		line3 = append(line3, renderAPIRatioLabeled(*m.APIWaitRatio, t))
	}
	if cfg.Features.CostVelocity && m.CostPerMinute != nil {
		line3 = append(line3, renderCostVelocityLabeled(*m.CostPerMinute, t))
	}
	if cfg.Features.VimMode && d.Vim != nil && d.Vim.Mode != "" {
		line3 = append(line3, renderVimCompact(d.Vim.Mode))
	}
	if cfg.Features.AgentName && d.Agent != nil && d.Agent.Name != "" {
		line3 = append(line3, renderAgentCompact(d.Agent.Name))
	}
	if cfg.Features.Effort {
		if s := renderEffort(d.Effort); s != "" {
			line3 = append(line3, s)
		}
	}
	if cfg.Features.Thinking {
		if s := renderThinking(d.Thinking); s != "" {
			line3 = append(line3, s)
		}
	}
	if cfg.Features.SessionName {
		if s := renderSessionName(d.SessionName); s != "" {
			line3 = append(line3, s)
		}
	}
	if cfg.Features.PullRequest {
		if s := renderPR(d.PR); s != "" {
			line3 = append(line3, s)
		}
	}
	if cfg.Features.Worktree {
		if s := renderWorktreeName(d.Worktree); s != "" {
			line3 = append(line3, s)
		}
	}
	if v := renderVersion(d.Version); v != "" {
		line3 = append(line3, v)
	}

	// Line 4: tools and agents with shared width budget
	line4 := make([]string, 0, 3)

	// Pre-render agents to calculate remaining budget for tools
	var agentStr string
	if cfg.Features.Agents && tools != nil && len(tools.Agents) > 0 {
		agentStr = renderAgents(tools.Agents)
	}

	toolBudget := terminalColumns()
	if agentStr != "" {
		toolBudget -= visibleLen(agentStr) + 3 // 3 for " | " separator
	}
	if toolBudget < 20 {
		toolBudget = 20
	}

	if cfg.Features.Tools && tools != nil && len(tools.Tools) > 0 {
		line4 = append(line4, renderTools(tools.Tools, toolBudget))
	}
	if agentStr != "" {
		line4 = append(line4, agentStr)
	}

	lines := []string{joinParts(line1)}
	if len(line2) > 0 {
		lines = append(lines, joinParts(line2))
	}
	if len(line3) > 0 {
		lines = append(lines, joinParts(line3))
	}
	if len(line4) > 0 {
		lines = append(lines, joinParts(line4))
	}
	return lines
}

func renderDangerMode(rc RenderContext) []string {
	d, m, git, usage, t := rc.Data, rc.Metrics, rc.Git, rc.Usage, rc.Config.Thresholds
	// L1: model | 🔴 danger context bar (remaining+ETA) | 5h quota bar | 7d quota bar
	line1 := make([]string, 0, 4)
	line1 = append(line1, renderModelBadge(d.Model, 0))
	line1 = append(line1, renderContextBarDanger(m.ContextPercent, d.ContextWindow, d.Cost.TotalDurationMS, t))

	if usage != nil {
		if s := renderQuotaWindow(usage.FiveHour, "5h", t); s != "" {
			line1 = append(line1, s)
		}
		if s := renderQuotaWindow(usage.SevenDay, "7d", t); s != "" {
			line1 = append(line1, s)
		}
	}

	// L2: workspace/git | Δchanges | In:XK Out:XK | C:X% | speed | $cost $cost/h | duration
	line2 := make([]string, 0, 8)

	if ws := renderWorkspace(d); ws != "" {
		if git != nil && git.Branch != "" {
			line2 = append(line2, ws+renderGitCompact(git))
		} else {
			line2 = append(line2, ws)
		}
	}

	if lines := renderLineChanges(d.Cost); lines != "" {
		line2 = append(line2, lines)
	}

	if cu := d.ContextWindow.CurrentUsage; cu != nil {
		line2 = append(line2, renderTokenIO(cu))
	}

	if m.CacheEfficiency != nil {
		line2 = append(line2, renderCacheEfficiencyCompact(*m.CacheEfficiency, t))
	}

	if costStr := renderCost(d.Cost.TotalCostUSD, t); costStr != "" {
		costParts := costStr
		if m.CostPerMinute != nil {
			hourly := *m.CostPerMinute * 60
			costParts += " " + fmt.Sprintf("%s$%.1f/h%s", yellow, hourly, Reset)
		}
		line2 = append(line2, costParts)
	}

	line2 = append(line2, renderDuration(d.Cost.TotalDurationMS))

	return []string{joinParts(line1), joinParts(line2)}
}

func renderModelBadge(m Model, contextSize int) string {
	name := m.DisplayName
	if name == "" {
		name = m.ID
	}
	if name == "" {
		name = "?"
	}

	// Detect provider
	suffix := ""
	lowerID := strings.ToLower(m.ID)
	if strings.Contains(lowerID, "anthropic.claude-") {
		suffix = " BR" // Bedrock
	} else if strings.Contains(lowerID, "publishers/anthropic") {
		suffix = " VX" // Vertex
	}

	// Append context window size (only for large windows like 1M+)
	if contextSize >= 1000000 {
		suffix += " (" + formatTokenCount(contextSize) + " context)"
	}

	var color string
	switch classifyModel(m) {
	case TierOpus:
		color = boldYlw
	case TierSonnet:
		color = cyan
	case TierHaiku:
		color = green
	default:
		color = dim
	}

	return fmt.Sprintf("%s[%s%s]%s", color, name, suffix, Reset)
}

func renderContextBar(percent int, cw ContextWindow, t Thresholds) string {
	const width = 10
	filled := min(width*percent/100, width)
	bar := buildBar(filled, width)

	color := contextColor(percent, t)
	prefix := ""
	if percent >= t.ContextDanger {
		prefix = "🔴 "
	} else if percent >= t.ContextWarning {
		prefix = "⚠ "
	}

	// Calculate absolute token usage from percentage (more accurate)
	// The percentage is provided by Claude Code and accounts for all token types
	totalTokens := cw.ContextWindowSize
	usedTokens := totalTokens * percent / 100

	// Pad used tokens to match total's width for fixed-width display
	total := formatTokenCount(totalTokens)
	used := fmt.Sprintf("%*s", len(total), formatTokenCount(usedTokens))

	return fmt.Sprintf("%s%s%s%s %3d%% (%s/%s)", prefix, color, bar, Reset, percent, used, total)
}

// renderContextBarDanger renders context bar with remaining tokens and ETA for danger mode.
func renderContextBarDanger(percent int, cw ContextWindow, durationMS int64, t Thresholds) string {
	const width = 10
	filled := min(width*percent/100, width)
	bar := buildBar(filled, width)

	color := contextColor(percent, t)
	prefix := "🔴 "

	// Remaining tokens
	totalTokens := cw.ContextWindowSize
	usedTokens := totalTokens * percent / 100
	remainTokens := totalTokens - usedTokens

	// ETA: estimate minutes until context is full
	eta := ""
	durationMin := float64(durationMS) / msPerMinute
	if percent > 0 && durationMin > 0 {
		remainPct := float64(100 - percent)
		etaMin := remainPct * durationMin / float64(percent)
		if etaMin < 1 {
			eta = " ~<1m"
		} else if etaMin < 60 {
			eta = fmt.Sprintf(" ~%.0fm", etaMin)
		} else {
			h := int(etaMin) / 60
			m := int(etaMin) % 60
			if m == 0 {
				eta = fmt.Sprintf(" ~%dh", h)
			} else {
				eta = fmt.Sprintf(" ~%dh%dm", h, m)
			}
		}
	}

	return fmt.Sprintf("%s%s%s%s %3d%% (%s left%s)", prefix, color, bar, Reset, percent,
		formatTokenCount(remainTokens), eta)
}

// renderTokenIO renders only In/Out token counts (without cache, used in danger mode).
func renderTokenIO(cu *CurrentUsage) string {
	return fmt.Sprintf("%sIn:%s%s %sOut:%s%s",
		grey, formatTokenCount(cu.InputTokens), Reset,
		grey, formatTokenCount(cu.OutputTokens), Reset)
}

func contextColor(p int, t Thresholds) string {
	switch {
	case p >= t.ContextDanger:
		return boldRed
	case p >= t.ContextWarning:
		return orange
	case p >= t.ContextModerate:
		return yellow
	default:
		return green
	}
}

func renderCost(usd float64, t Thresholds) string {
	if usd < 0.001 {
		return ""
	}
	var color string
	switch {
	case usd >= t.SessionCostHigh:
		color = boldRed
	case usd >= t.SessionCostMedium:
		color = yellow
	default:
		color = Reset
	}
	if usd < 10 {
		return fmt.Sprintf("%s$%.2f%s", color, usd, Reset)
	}
	return fmt.Sprintf("%s$%.1f%s", color, usd, Reset)
}

func renderDuration(ms int64) string {
	minutes := ms / msPerMinute
	if minutes < 1 {
		return grey + "<1m" + Reset
	}
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	h := minutes / 60
	rem := minutes % 60
	return fmt.Sprintf("%dh%dm", h, rem)
}

func renderWorkspace(d *StdinData) string {
	dir := d.Workspace.ProjectDir
	if dir == "" {
		dir = d.Workspace.CurrentDir
	}
	if dir == "" {
		return ""
	}
	// Extract last path component
	name := dir
	for i := len(dir) - 1; i >= 0; i-- {
		if dir[i] == '/' {
			name = dir[i+1:]
			break
		}
	}
	if name == "" {
		return ""
	}
	return grey + name + "/" + Reset
}

func renderAccount(account *AccountInfo) string {
	// Display full email in grey/dim color
	return grey + account.EmailAddress + Reset
}

func renderGitCompact(g *GitInfo) string {
	if g.Branch == "" {
		return ""
	}
	dirty := ""
	if g.Dirty {
		dirty = "*"
	}
	return fmt.Sprintf("%s%s%s%s", magenta, g.Branch, dirty, Reset)
}

func renderLineChanges(c Cost) string {
	if c.TotalLinesAdded == 0 && c.TotalLinesRemoved == 0 {
		return ""
	}
	add := formatCount(c.TotalLinesAdded)
	del := formatCount(c.TotalLinesRemoved)
	return fmt.Sprintf("%sΔ%s%s+%s%s/%s-%s%s", grey, Reset, green, add, Reset, red, del, Reset)
}

func cacheColor(pct int, t Thresholds) string {
	switch {
	case pct >= t.CacheExcellent:
		return green
	case pct >= t.CacheGood:
		return yellow
	default:
		return red
	}
}

func renderCacheEfficiencyCompact(pct int, t Thresholds) string {
	return fmt.Sprintf("%sC%d%%%s", cacheColor(pct, t), pct, Reset)
}

func renderCacheEfficiencyLabeled(pct int, t Thresholds, cu *CurrentUsage) string {
	base := fmt.Sprintf("%sCache:%s%d%%%s", grey, cacheColor(pct, t), pct, Reset)
	if cu == nil {
		return base
	}
	w := formatTokenCount(cu.CacheCreationInputTokens)
	r := formatTokenCount(cu.CacheReadInputTokens)
	return fmt.Sprintf("%s%s(W:%s/R:%s)%s", base, grey, w, r, Reset)
}

func apiRatioColor(pct int, t Thresholds) string {
	switch {
	case pct >= t.WaitHigh:
		return red
	case pct >= t.WaitMedium:
		return yellow
	default:
		return green
	}
}

func renderAPIRatioLabeled(pct int, t Thresholds) string {
	return fmt.Sprintf("%sWait:%s%d%%%s", grey, apiRatioColor(pct, t), pct, Reset)
}

func renderCostVelocityLabeled(perMin float64, t Thresholds) string {
	var color string
	switch {
	case perMin >= t.CostVelocityHigh:
		color = boldRed
	case perMin >= t.CostVelocityMedium:
		color = yellow
	default:
		color = green
	}
	return fmt.Sprintf("%sCost:%s$%.2f/m%s", grey, color, perMin, Reset)
}

// renderOutputTokens shows the output token count of the current/last API
// response — a truthful replacement for the removed tok/s speed metric.
// Returns "" when there is nothing to show.
func renderOutputTokens(cu *CurrentUsage) string {
	if cu == nil || cu.OutputTokens <= 0 {
		return ""
	}
	return grey + "Out:" + formatTokenCount(cu.OutputTokens) + Reset
}

func renderVimCompact(mode string) string {
	var color string
	m := strings.ToLower(mode)
	switch m {
	case "normal":
		color = blue
		return color + "Normal" + Reset
	case "insert":
		color = green
		return color + "Insert" + Reset
	case "visual":
		color = magenta
		return color + "Visual" + Reset
	default:
		return ""
	}
}

func renderVersion(version string) string {
	if version == "" {
		return ""
	}
	return grey + "v" + version + Reset
}

// Optional CC 2.1 field renderers. Each returns "" when its source is absent so
// callers append conditionally (OCP: add a field by adding a function).

func renderEffort(e *Effort) string {
	if e == nil || e.Level == "" {
		return ""
	}
	return dim + "E:" + e.Level + Reset
}

func renderThinking(th *Thinking) string {
	if th == nil || !th.Enabled {
		return ""
	}
	return cyan + "Think" + Reset
}

func renderSessionName(name string) string {
	if name == "" {
		return ""
	}
	runes := []rune(name)
	if len(runes) > 12 {
		name = string(runes[:12])
	}
	return dim + name + Reset
}

func renderPR(pr *PullRequest) string {
	if pr == nil || pr.Number == 0 {
		return ""
	}
	s := fmt.Sprintf("%sPR#%d", cyan, pr.Number)
	if pr.ReviewState != "" {
		s += " " + pr.ReviewState
	}
	return s + Reset
}

func renderWorktreeName(wt *Worktree) string {
	if wt == nil || wt.Name == "" {
		return ""
	}
	name := wt.Name
	runes := []rune(name)
	if len(runes) > 10 {
		name = string(runes[:10])
	}
	return cyan + "wt:" + name + Reset
}

// terminalColumns returns the terminal width from the COLUMNS env var that
// Claude Code provides to statusline commands (v2.1.153+), clamped to a sane
// range. Falls back to maxToolLineWidth when unset/invalid (tests, non-Claude
// execution) so behavior is unchanged where COLUMNS is absent.
func terminalColumns() int {
	n, err := strconv.Atoi(os.Getenv("COLUMNS"))
	if err != nil {
		return maxToolLineWidth
	}
	return min(max(n, 40), 240)
}

func renderAgentCompact(name string) string {
	runes := []rune(name)
	if len(runes) > 8 {
		name = string(runes[:8])
	}
	return cyan + "@" + name + Reset
}

const (
	maxToolNameLen   = 12
	maxToolLineWidth = 80
)

// visibleLen computes the display width of a string, excluding ANSI escape sequences.
// This implementation assumes all ANSI codes follow the CSI SGR (Select Graphic Rendition)
// pattern: \033[...m. This is valid for this codebase, which uses only standard SGR codes
// (colors, bold, dim) and 8-bit extended colors (\033[38;5;Nm). Other escape sequence
// types (like cursor movement or OSC sequences) are not used and would require different parsing.
func visibleLen(s string) int {
	inEscape := false
	count := 0
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		count++
	}
	return count
}

func truncateToolName(name string, maxLen int) string {
	runes := []rune(name)
	if len(runes) <= maxLen {
		return name
	}
	return string(runes[:maxLen-1]) + "…"
}

func renderTools(tools map[string]int, maxWidth int) string {
	if len(tools) == 0 {
		return ""
	}
	// Sort by count and show top 5
	type entry struct {
		name  string
		count int
	}
	entries := make([]entry, 0, len(tools))
	for name, count := range tools {
		entries = append(entries, entry{name, count})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].count > entries[j].count
	})
	if len(entries) > 5 {
		entries = entries[:5]
	}

	var result string
	shown := 0
	for _, e := range entries {
		name := truncateToolName(e.name, maxToolNameLen)
		part := fmt.Sprintf("%s%s%s(%d)", blue, name, Reset, e.count)

		candidate := result
		if shown > 0 {
			candidate += " "
		}
		candidate += part

		// Always show at least 1 tool; after that, check budget
		if shown > 0 && visibleLen(candidate) > maxWidth {
			remaining := len(entries) - shown
			if remaining > 0 {
				result += fmt.Sprintf(" +%d", remaining)
			}
			break
		}

		if shown > 0 {
			result += " "
		}
		result += part
		shown++
	}
	return result
}

func renderAgents(agents []string) string {
	if len(agents) == 0 {
		return ""
	}
	// Show first 2 running agents
	display := agents
	if len(display) > 2 {
		display = display[:2]
	}
	result := fmt.Sprintf("%s▶%s", yellow, Reset)
	for i, a := range display {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%s%s%s", cyan, a, Reset)
	}
	return result
}

func formatCount(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000.0)
	}
	return fmt.Sprintf("%d", n)
}

func formatTokenCount(tokens int) string {
	if tokens >= 1_000_000 {
		m := float64(tokens) / 1_000_000.0
		if m == float64(int(m)) {
			return fmt.Sprintf("%dM", int(m))
		}
		return fmt.Sprintf("%.1fM", m)
	}
	return fmt.Sprintf("%.0fK", float64(tokens)/1000.0)
}

// renderQuotaWindow renders one quota window, or "" when the window is absent.
// Single code path for 5h/7d across normal and danger modes (DRY); presence is
// the pointer, never a zero value.
func renderQuotaWindow(w *UsageWindow, label string, t Thresholds) string {
	if w == nil {
		return ""
	}
	return renderQuotaBar(w.RemainingPercent, w.ResetsAt, label, t)
}

func renderQuotaBar(remainPct float64, resetTime time.Time, label string, t Thresholds) string {
	const width = 10
	filled := max(0, min(int(remainPct)*width/100, width))
	bar := buildBar(filled, width)

	color := quotaColor(remainPct, t)
	now := time.Now()
	var until string
	var windowLen time.Duration
	if label == "5h" {
		until = formatTimeUntilWithMinutes(now, resetTime)
		windowLen = 5 * time.Hour
	} else {
		until = formatTimeUntil(now, resetTime)
		windowLen = 7 * 24 * time.Hour
	}

	marker := quotaPaceMarker(remainPct, resetTime, now, windowLen)
	return fmt.Sprintf("%s%s%s %3.0f%% (%s/%s)%s", color, bar, Reset, remainPct, until, label, marker)
}

// quotaPaceMarker returns a "🔥" warning when the window is projected to be
// exhausted before it resets — i.e. consumption is ahead of the even time-based
// pace. Pure function of a single stdin snapshot (no history needed): it
// compares the average burn rate so far against the rate that would last until
// reset. Returns "" when on pace, already exhausted, or signal is too weak.
func quotaPaceMarker(remainPct float64, resetsAt, now time.Time, windowLen time.Duration) string {
	if resetsAt.IsZero() || remainPct <= 0 || windowLen <= 0 {
		return ""
	}
	timeToReset := resetsAt.Sub(now)
	if timeToReset <= 0 {
		return ""
	}
	elapsed := windowLen - timeToReset
	if elapsed < windowLen/20 { // need >=5% elapsed to project meaningfully
		return ""
	}
	used := 100 - remainPct
	if used <= 0 {
		return ""
	}
	// Ahead of pace iff time-to-exhaust < time-to-reset. Cross-multiplied to
	// avoid float-duration construction: remainPct/burn < timeToReset, where
	// burn = used/elapsed  =>  remainPct*elapsed < used*timeToReset.
	if remainPct*elapsed.Seconds() < used*timeToReset.Seconds() {
		return boldRed + "🔥" + Reset
	}
	return ""
}

func formatTimeUntil(now, target time.Time) string {
	return formatTimeUntilWith(now, target, false)
}

func formatTimeUntilWithMinutes(now, target time.Time) string {
	return formatTimeUntilWith(now, target, true)
}

func formatTimeUntilWith(now, target time.Time, includeMinutes bool) string {
	if target.IsZero() {
		return "?"
	}
	diff := target.Sub(now)
	if diff < 0 {
		if includeMinutes {
			return "0h00m"
		}
		return "0d00h"
	}

	hours := int(diff.Hours())
	minutes := int(diff.Minutes()) % 60

	if includeMinutes {
		return fmt.Sprintf("%dh%02dm", hours, minutes)
	}
	days := hours / 24
	remainHours := hours % 24
	return fmt.Sprintf("%dd%02dh", days, remainHours)
}

func quotaColor(remaining float64, t Thresholds) string {
	switch {
	case remaining < t.QuotaCritical:
		return boldRed
	case remaining < t.QuotaLow:
		return red
	case remaining < t.QuotaMedium:
		return orange
	case remaining < t.QuotaHigh:
		return yellow
	default:
		return green
	}
}

// buildBar creates a fixed-width progress bar with filled (█) and empty (░) characters.
func buildBar(filled, width int) string {
	var b strings.Builder
	b.Grow(width)
	for range filled {
		b.WriteRune('█')
	}
	for range width - filled {
		b.WriteRune('░')
	}
	return b.String()
}

func joinParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += " " + grey + "|" + Reset + " " + parts[i]
	}
	return result
}
