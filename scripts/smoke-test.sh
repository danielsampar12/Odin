#!/usr/bin/env bash
# First-run smoke test: builds odin, then runs the user-facing flow in an empty HOME with a PATH that
# has no Ollama or OpenCode, the way a new user's machine looks before installing them.
# Usage: scripts/smoke-test.sh (needs Go and git; honours TMPDIR)
set -euo pipefail

repo=$(cd "$(dirname "$0")/.." && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
(cd "$repo" && go build -o "$work/bin/odin" ./cmd/odin)

export HOME="$work/home" PATH="$work/bin:/usr/bin:/bin"
mkdir -p "$HOME"

fail() { echo "FAIL: $*" >&2; exit 1; }
run() { out=$(odin "$@" 2>&1) || fail "odin $* exited non-zero: $out"; }
expect_failure() { if out=$(odin "$@" 2>&1); then fail "odin $* should have failed: $out"; fi; }
contains() { grep -qF -- "$1" <<<"$out" || fail "output is missing '$1': $out"; }

run version
contains "odin v"
run doctor
contains "Odin doctor"
run setup --dry-run
contains "No changes were made"
[ ! -e "$HOME/.odin" ] || fail "setup --dry-run wrote $HOME/.odin"
run model recommend
contains "Recommended coding model:"

project="$work/project"
mkdir -p "$project" && cd "$project" && git init -q
run init
contains "Created"
files=(.odin/config.toml .odin/rules.md AGENTS.md)
for f in "${files[@]}"; do [ -f "$f" ] || fail "init did not create $f"; done
grep -q 'provider = "none"' .odin/config.toml || fail "memory should default to none"
before=$(cat "${files[@]}" | sha256sum)
run init
contains "already exists"
[ "$before" = "$(cat "${files[@]}" | sha256sum)" ] || fail "a second init changed project files"

run opencode generate
contains "Created"
grep -q '".odin/rules.md"' .odin/generated/opencode.jsonc || fail "generated config does not load .odin/rules.md"
grep -qE '"model": "ollama/.+-ctx(16|32)k"' .odin/generated/opencode.jsonc || fail "generated config does not use a context model"
run opencode generate
contains "Kept existing"

expect_failure start --exec
contains "cannot launch OpenCode"
echo '{}' > .odin/generated/opencode.jsonc
expect_failure opencode generate
contains "refusing to overwrite"

echo "smoke test passed"
