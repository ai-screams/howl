---
description: This skill should be used when the user asks to "install howl", "set up the statusline", "configure howl", "enable the HUD", or "install the statusline binary". It downloads the Howl binary and configures Claude Code to use it.
disable-model-invocation: true
---

# Howl Setup

Install and configure the Howl statusline HUD for Claude Code.

## What This Does

1. Download the latest Howl binary from GitHub Releases
2. Install it to `~/.claude/hud/howl`
3. Configure `~/.claude/settings.json` to use Howl as the statusline
4. Prompt the user to restart Claude Code

## Installation

Run the install script:

```bash
bash "${CLAUDE_PLUGIN_ROOT}/scripts/install.sh"
```

## Verify Installation

After the script completes, verify the binary works:

```bash
echo '{}' | ~/.claude/hud/howl
```

The output should show a statusline with default values. If the binary is not found or fails, check:

- `~/.claude/hud/howl` exists and is executable
- Run `chmod +x ~/.claude/hud/howl` if needed

## Post-Install

Inform the user:

- Binary installed to `~/.claude/hud/howl`
- Statusline configured in `~/.claude/settings.json`
- **Restart Claude Code** to activate the statusline

## Troubleshooting

If the install fails:

- Check internet connectivity (needs access to github.com)
- Verify `curl` is available
- Try manual download from: https://github.com/ai-screams/Howl/releases/latest
- Supported platforms: macOS (arm64/amd64), Linux (arm64/amd64)

If the statusline does not appear after restart:

- Check `~/.claude/settings.json` contains the `statusLine` field
- Verify the binary path is correct for this system
- Run `~/.claude/hud/howl --version` to confirm the binary works
