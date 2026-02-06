package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ai-scream/howl/internal"
)

func main() {
	var data internal.StdinData
	if err := json.NewDecoder(os.Stdin).Decode(&data); err != nil {
		fmt.Fprint(os.Stderr, "howl: stdin parse error")
		os.Exit(1)
	}

	metrics := internal.ComputeMetrics(&data)

	dir := data.Workspace.ProjectDir
	if dir == "" {
		dir = data.CWD
	}
	git := internal.GetGitInfo(dir)

	// Try to fetch usage quota (optional)
	usage := internal.GetUsage(data.SessionID)

	// Parse transcript for tools/agents (optional)
	toolInfo := internal.ParseTranscript(data.TranscriptPath)

	lines := internal.Render(&data, metrics, git, usage, toolInfo)

	// Output each line individually with:
	// 1. RESET prefix to clear ANSI state
	// 2. Spaces replaced with NBSP (\u00A0) to prevent Claude Code from stripping them
	for _, line := range lines {
		line = strings.ReplaceAll(line, " ", "\u00A0")
		fmt.Println(internal.Reset + line)
	}
}
