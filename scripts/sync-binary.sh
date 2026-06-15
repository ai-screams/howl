#!/usr/bin/env bash
#
# sync-binary.sh — keep the installed Howl binary in step with the plugin version.
#
# Runs from the plugin's SessionStart hook. Claude Code auto-updates the plugin
# *content* (skills, scripts, plugin.json) from the marketplace, but the Howl
# *binary* at ~/.claude/hud/howl is a separately-downloaded release artifact and
# would otherwise drift. This script closes that gap.
#
# Behaviour:
#   - Fast no-op (no network) when already in sync — the common case.
#   - Only ever UPDATES an existing install; never auto-installs unprompted
#     (a user with no binary is expected to run /howl:setup explicitly).
#   - On a plugin version change, re-runs the installer in the background so
#     session start is never blocked. Fails silently; retries next session.
set -u

root="${CLAUDE_PLUGIN_ROOT:-}"
[ -n "$root" ] || exit 0

manifest="$root/.claude-plugin/plugin.json"
[ -f "$manifest" ] || exit 0

binary="$HOME/.claude/hud/howl"
# Only manage updates for an existing install — do not auto-install unprompted.
[ -x "$binary" ] || exit 0

# CLAUDE_PLUGIN_DATA is the persistent per-plugin state dir (survives updates).
# Fall back to a stable location on older Claude Code that doesn't set it.
data="${CLAUDE_PLUGIN_DATA:-$HOME/.claude/hud/.plugin-state}"
mkdir -p "$data" 2>/dev/null || exit 0
saved="$data/plugin.json"

# In sync (binary present + recorded manifest matches current) -> nothing to do.
[ -f "$saved" ] && cmp -s "$manifest" "$saved" && exit 0

# Out of sync: fetch the matching release in the background. Record the manifest
# only on success so a failed/offline run retries on the next session start.
nohup env HOWL_ROOT="$root" HOWL_MANIFEST="$manifest" HOWL_SAVED="$saved" bash -c '
  if bash "$HOWL_ROOT/scripts/install.sh" >/dev/null 2>&1; then
    cp "$HOWL_MANIFEST" "$HOWL_SAVED" 2>/dev/null || true
  fi
' >/dev/null 2>&1 &

exit 0
