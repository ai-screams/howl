# Howl Setup Skill - Agent Documentation

## Purpose

The `/howl:setup` skill installs the Howl binary and configures Claude Code to use it as the statusline HUD. It automates the entire installation process—from downloading the latest release to configuring the settings file—with built-in security verification.

## Skill Invocation

Users trigger this skill with:
```
/howl:setup
```

The skill automatically:
1. Downloads the latest Howl binary from GitHub Releases (auto-detects OS/arch: darwin/linux × amd64/arm64)
2. Verifies SHA256 checksum against the official release digest
3. Installs to `~/.claude/hud/howl` with executable permissions
4. Backs up existing `~/.claude/settings.json` before modification
5. Configures `~/.claude/settings.json` to include the Howl statusline path
6. Prompts user to restart Claude Code to activate the statusline

## Installation Flow

### Execution
The skill delegates to `scripts/install.sh`, which handles all installation logic. See `scripts/AGENTS.md` for detailed install script documentation.

### Output
Success flow:
- Binary installed: `~/.claude/hud/howl` (executable)
- Settings modified: `~/.claude/settings.json` now includes `statusLine` field
- Backup created: `~/.claude/settings.json.bak` (if settings file existed)

### Verification
After script completes, the skill verifies:
```bash
echo '{}' | ~/.claude/hud/howl
```
Expected: A default statusline output (shows all metrics with default values)

## Security

- **Checksum Verification**: Binary SHA256 hash validated against official release checksums
- **Settings Backup**: Existing settings backed up to `*.bak` before modification
- **Platform Detection**: Auto-detects OS and CPU architecture; fails safely if unsupported
- **Direct Download**: Fetches only from official GitHub Releases, no redirect chains

## Dependencies

- `curl` — Download binary and checksums from GitHub Releases
- `jq` or `python3` — JSON manipulation for settings.json (install script detects availability)
- `sha256sum` (macOS: `shasum -a 256`) — Checksum verification

## Agent Responsibilities

1. **Guide Installation** — Inform user that `/howl:setup` will download, verify, and install the binary
2. **Monitor Execution** — Watch for errors related to:
   - Network connectivity (GitHub access required)
   - Missing dependencies (curl, jq/python3)
   - Unsupported platforms (only darwin/linux × amd64/arm64 supported)
   - Permission issues (may need `chmod +x` if installation fails)
3. **Confirm Success** — After completion:
   - Verify `~/.claude/hud/howl --version` shows installed version
   - Remind user to restart Claude Code to activate the statusline
   - Check `~/.claude/settings.json` contains `"statusLine"` field
4. **Troubleshoot Failures** — If script fails:
   - Suggest manual download: https://github.com/ai-screams/howl/releases/latest
   - Verify platform support (check `uname -s` and `uname -m` output)
   - Check for `~/.claude/hud/` directory permissions
   - Fall back to manual JSON configuration if JSON tools unavailable

## Integration

The skill coordinates with:
- **scripts/install.sh** — Actual installation implementation (see scripts/AGENTS.md)
- **internal/config.go** — Configuration defaults and loading (not directly invoked by this skill)
- **~/.claude/settings.json** — Target configuration file (backed up before modification)

## User Communication

Agents should note:
1. **First-Time Install** — User will see "Downloading Howl vX.Y.Z..." with progress (latest release)
2. **Restart Required** — Statusline activates only after Claude Code restart
3. **Binary Location** — Installed to `~/.claude/hud/howl` (can be uninstalled by deleting this file)
4. **No Further Setup** — Once installed, use `/howl:configure` to customize display presets

This skill is zero-dependency and fully automated; agents only guide users through the process and interpret results.
