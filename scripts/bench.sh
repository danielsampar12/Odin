#!/usr/bin/env bash
# Benchmarks the Odin + OpenCode + Ollama stack on medium Exercism Go exercises (difficulty 4-7 of 10).
#
# Each run: a fresh git repo with the exercise stub, tests and instructions (no .meta/ solutions), a
# tool allowlist in opencode.json, `odin init`, `odin opencode generate`, one non-interactive
# `opencode run`, then `go test` against the original test files. Runs go one at a time (one GPU).
#
# Usage: scripts/bench.sh [held-out|dev|all|slugify|<exercise>...]   (default: held-out)
# Env:   RUNS=3          runs per task
#        AGENT=          OpenCode agent for `opencode run --agent` (e.g. baldur); empty = OpenCode's default
#        TIMEOUT=600     seconds per `opencode run`
#        BENCH_DIR=      where repos, logs and results.tsv go (default: a new dir under ~/.cache/odin-bench)
#        EXERCISM_DIR=   checkout of github.com/exercism/go (cloned and pinned automatically)
#        TELEMETRY=      CSV written by scripts/telemetry.sh; enables the thermal pause and the guardrail report
# Needs odin, opencode, ollama, go, git and jq on PATH, a running Ollama with the configured model,
# and `odin start --exec` run once on this machine (and again after upgrading odin) so the context model
# exists with the current parameters.
set -u

EXERCISM_COMMIT=6d9eed09379935e9721d6a57ba5675f4f803fcb3
# Fixed split, chosen with random.Random(20261028).sample() over active exercises of difficulty 4-7.
DEV=(poker circular-buffer save-the-cow binary-search-tree)
HELD_OUT=(robot-simulator dot-dsl meetup tree-building affine-cipher paasio simple-linked-list rectangles)

RUNS=${RUNS:-3}
AGENT=${AGENT:-}
TIMEOUT=${TIMEOUT:-600}
BENCH_DIR=${BENCH_DIR:-$HOME/.cache/odin-bench/$(date +%Y%m%d-%H%M%S)}
EXERCISM_DIR=${EXERCISM_DIR:-$HOME/.cache/odin-bench/exercism-go}
OUT="$BENCH_DIR/results.tsv"

case "${1:-held-out}" in
  held-out) TASKS=("${HELD_OUT[@]}") ;;
  dev) TASKS=("${DEV[@]}") ;;
  all) TASKS=("${HELD_OUT[@]}" "${DEV[@]}" slugify) ;;
  *) TASKS=("$@") ;;
esac

for tool in odin opencode ollama go git jq; do
  command -v "$tool" >/dev/null || { echo "bench: $tool is not on PATH" >&2; exit 1; }
done

if [ ! -d "$EXERCISM_DIR/.git" ]; then
  git clone -q https://github.com/exercism/go.git "$EXERCISM_DIR" || exit 1
fi
git -C "$EXERCISM_DIR" cat-file -e "$EXERCISM_COMMIT" 2>/dev/null || git -C "$EXERCISM_DIR" fetch -q origin || exit 1
git -C "$EXERCISM_DIR" -c advice.detachedHead=false checkout -q "$EXERCISM_COMMIT" || exit 1

mkdir -p "$BENCH_DIR" && BENCH_DIR=$(cd "$BENCH_DIR" && pwd) && OUT="$BENCH_DIR/results.tsv" # runs cd into each repo
echo "odin: $(odin version) ($(command -v odin)); opencode $(opencode --version); agent: ${AGENT:-default}" | tee "$BENCH_DIR/versions.txt"
[ -f "$OUT" ] || printf 'task\trun\tpass\tseconds\texit\ttools\ttool_errors\ttests_edited\tclaimed_pass\tfalse_claim\tprocessor\tcontext\tstarted\n' > "$OUT"
# One model on the GPU at a time: unload whatever another config left loaded.
ollama ps | awk 'NR > 1 {print $1}' | while read -r loaded; do ollama stop "$loaded"; done
BENCH_START=$(date +%s)
SWAP_START=$(free -m | awk '/^Swap:/ {print $3}')

# Pauses before a run when the previous one got the GPU to 83 C, set a thermal-slowdown flag or got the
# CPU package to 90 C, until both are below 70 C.
cool_down() {
  [ -n "${TELEMETRY:-}" ] && [ -f "$TELEMETRY" ] || return 0
  awk -F, -v since="${1:-0}" 'NR > 1 && $1 >= since && ($2 >= 83 || $6 == "Active" || $7 == "Active" || ($8 != "?" && $8 >= 90)) {hot = 1} END {exit !hot}' "$TELEMETRY" || return 0
  echo "$(date +%T) thermal guardrail: pausing until GPU and CPU are below 70 C" | tee -a "$BENCH_DIR/guardrails.log"
  while true; do
    local gpu cpu
    gpu=$(nvidia-smi --query-gpu=temperature.gpu --format=csv,noheader,nounits | head -1)
    cpu=$(sensors 2>/dev/null | awk '/^Package id 0:/ {gsub(/[^0-9.]/, "", $4); print int($4); exit}')
    [ "${gpu:-0}" -lt 70 ] && [ "${cpu:-0}" -lt 70 ] && break
    sleep 10
  done
  echo "$(date +%T) thermal guardrail: resumed" >> "$BENCH_DIR/guardrails.log"
}

# Grants what a user would approve for a coding task (the Go toolchain and read-only commands) without
# OpenCode's interactive prompts; nothing that deletes or moves files. OpenCode checks each part of
# `cd repo && go test ./... | head` separately, so cd and head are allowed too.
ALLOWLIST='{
  "$schema": "https://opencode.ai/config.json",
  "permission": {
    "edit": "allow",
    "webfetch": "deny",
    "external_directory": "deny",
    "bash": {
      "*": "deny",
      "go test*": "allow",
      "go vet*": "allow",
      "go build*": "allow",
      "go run*": "allow",
      "go doc*": "allow",
      "go env*": "allow",
      "gofmt*": "allow",
      "cd*": "allow",
      "pwd*": "allow",
      "ls*": "allow",
      "cat*": "allow",
      "grep*": "allow",
      "head*": "allow",
      "tail*": "allow",
      "wc*": "allow",
      "sort*": "allow",
      "diff*": "allow"
    }
  }
}'

SLUGIFY_TASK='Add a function Slugify(s string) string in slug.go (package sample) that lowercases s, replaces runs of non-alphanumeric characters with a single hyphen, and trims leading/trailing hyphens. Add a table-driven test in slug_test.go. Then run `go test ./...` and make sure it passes.'

setup_repo() { # task dir -> prints the prompt
  local task=$1 dir=$2
  mkdir -p "$dir"
  if [ "$task" = slugify ]; then
    # The Stage 0 sample repo.
    printf 'module example.com/sample\n\ngo 1.22\n' > "$dir/go.mod"
    printf 'package sample\n\n// Reverse returns s with its runes in reverse order.\nfunc Reverse(s string) string {\n\tr := []rune(s)\n\tfor i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {\n\t\tr[i], r[j] = r[j], r[i]\n\t}\n\treturn string(r)\n}\n' > "$dir/reverse.go"
    printf 'package sample\n\nimport "testing"\n\nfunc TestReverse(t *testing.T) {\n\ttests := []struct{ in, want string }{\n\t\t{"", ""},\n\t\t{"abc", "cba"},\n\t\t{"h\xc3\xa9llo", "oll\xc3\xa9h"},\n\t}\n\tfor _, tt := range tests {\n\t\tif got := Reverse(tt.in); got != tt.want {\n\t\t\tt.Errorf("Reverse(%%q) = %%q, want %%q", tt.in, got, tt.want)\n\t\t}\n\t}\n}\n' > "$dir/reverse_test.go"
    echo "$SLUGIFY_TASK"
    return
  fi
  local src="$EXERCISM_DIR/exercises/practice/$task"
  [ -d "$src" ] || { echo "bench: unknown exercise $task" >&2; return 1; }
  # Keep stub, tests and instructions; drop the reference solution and approach write-ups.
  (cd "$src" && tar --exclude=.meta --exclude=.approaches --exclude=.articles -cf - .) | tar -xf - -C "$dir"
  local solution
  solution=$(jq -r '.files.solution | join(", ")' "$src/.meta/config.json")
  echo "Solve the programming exercise in this repository. It is described in .docs/instructions.md (and .docs/instructions.append.md if that file exists), and the *_test.go files define the expected behaviour. Implement the solution in $solution. Do not modify the test files. Run \`go test ./...\` and keep fixing the code until all tests pass."
}

run_one() { # task run
  local task=$1 run=$2 dir="$BENCH_DIR/$1-$2" log="$BENCH_DIR/$1-$2.jsonl"
  rm -rf "$dir"
  local prompt
  prompt=$(setup_repo "$task" "$dir") || return 1
  cd "$dir" || return 1
  git init -q && git add -A && git -c user.name=bench -c user.email=bench@localhost commit -qm task
  echo "$ALLOWLIST" > opencode.json
  odin init > "$BENCH_DIR/$task-$run.odin.txt" 2>&1 && odin opencode generate >> "$BENCH_DIR/$task-$run.odin.txt" 2>&1 \
    || { echo "bench: odin failed, see $BENCH_DIR/$task-$run.odin.txt" >&2; exit 1; }
  local model
  model=$(grep -oE '"model": "ollama/[^"]+"' .odin/generated/opencode.jsonc | sed 's|.*ollama/||; s|"$||')
  ollama list | awk 'NR>1{print $1}' | grep -qx "$model" \
    || { echo "bench: Ollama model $model is missing; run \`odin start --exec\` once to create it" >&2; exit 1; }

  # Launch with the environment this odin version uses for OpenCode (OPENCODE_CONFIG and friends).
  local launch_env
  launch_env=$(odin opencode status | sed -n 's/^Planned launch command: //p' | grep -oE '[A-Z_]+=[^ ]+')
  [ -n "$launch_env" ] || { echo "bench: could not read the launch command from odin opencode status" >&2; exit 1; }

  cool_down "${LAST_START:-$BENCH_START}"
  local start code secs
  start=$(date +%s)
  LAST_START=$start
  # shellcheck disable=SC2086 # launch_env is a list of VAR=value words
  env $launch_env timeout "$TIMEOUT" opencode run --format json ${AGENT:+--agent "$AGENT"} "$prompt" > "$log" 2>&1 < /dev/null
  code=$?
  secs=$(( $(date +%s) - start ))
  local ps_line processor context
  ps_line=$(ollama ps | awk -v m="$model" '$1==m')
  processor=$(echo "$ps_line" | grep -oE '[0-9]+%(/[0-9]+%)? (GPU|CPU)(/GPU)?' | head -1)
  context=$(echo "$ps_line" | awk '{for(i=4;i<=NF;i++) if ($i ~ /^[0-9]+$/ && $i >= 1024) {print $i; exit}}')

  local tools tool_errors final claimed edited=no pass=no false_claim=no
  tools=$(grep '^{' "$log" | jq -s '[.[] | select(.type=="tool_use")] | length')
  tool_errors=$(grep '^{' "$log" | jq -s '[.[] | select(.type=="tool_use" and .part.state.status=="error")] | length')
  final=$(grep '^{' "$log" | jq -rs '[.[] | select(.type=="text") | .part.text] | last // ""')
  printf '%s\n' "$final" > "$BENCH_DIR/$task-$run.final.txt"
  claimed=no
  if printf '%s' "$final" | grep -qiE '(tests?|all).{0,40}pass' && ! printf '%s' "$final" | grep -qiE 'fail|not pass|n.t pass'; then
    claimed=yes
  fi
  # Score against the original tests: an agent that edits a test file gets it restored first.
  if [ "$task" != slugify ]; then
    git diff --quiet HEAD -- '*_test.go' || edited=yes
    [ -z "$(git ls-files -o -- '*_test.go')" ] || edited=yes # an added test file could exit 0 from TestMain
    git checkout -q HEAD -- '*_test.go' && git clean -fq -- '*_test.go'
  fi
  if [ "$task" = slugify ]; then
    [ -f slug.go ] && [ -f slug_test.go ] && timeout 300 go test ./... > "$BENCH_DIR/$task-$run.gotest.txt" 2>&1 && pass=yes
  else
    timeout 300 go test ./... > "$BENCH_DIR/$task-$run.gotest.txt" 2>&1 && pass=yes
  fi
  [ "$claimed" = yes ] && [ "$pass" = no ] && false_claim=yes
  printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$task" "$run" "$pass" "$secs" "$code" "$tools" "$tool_errors" \
    "$edited" "$claimed" "$false_claim" "${processor:-?}" "${context:-?}" "$start" >> "$OUT"
  tail -1 "$OUT"
  cd - >/dev/null || return 1
}

for task in "${TASKS[@]}"; do
  for run in $(seq 1 "$RUNS"); do
    run_one "$task" "$run" || exit 1
  done
done

echo
echo "Results ($OUT):"
awk -F'\t' 'NR>1 {
  n[$1]++; p[$1]+=($3=="yes"); f[$1]+=($10=="yes"); tests[$1]=tests[$1] " " $4
  total++; passes+=($3=="yes"); falses+=($10=="yes"); all[total]=$4
}
function median(list,   a, k, i, j, t) {
  k = split(list, a, " ")
  for (i = 1; i <= k; i++) for (j = i + 1; j <= k; j++) if (a[j] + 0 < a[i] + 0) { t = a[i]; a[i] = a[j]; a[j] = t }
  return k % 2 ? a[(k + 1) / 2] : (a[k / 2] + a[k / 2 + 1]) / 2
}
END {
  printf "%-20s %6s %10s %12s\n", "task", "passed", "median s", "false claims"
  for (t in n) printf "%-20s %3d/%-2d %10s %12d\n", t, p[t], n[t], median(tests[t]), f[t]
  s = ""; for (i = 1; i <= total; i++) s = s " " all[i]
  printf "%-20s %3d/%-2d %10s %12d   pass@1 %.0f%%\n", "TOTAL", passes, total, median(s), falses, 100 * passes / total
}' "$OUT"

# Guardrails: peaks over this benchmark from the telemetry, and the reasons to reject the config.
[ -n "${TELEMETRY:-}" ] && [ -f "$TELEMETRY" ] || exit 0
oom=$(journalctl -k --since "@$BENCH_START" --no-pager 2>/dev/null | grep -ciE 'out of memory|oom-kill')
median=$(awk -F'\t' 'NR > 1 {print $4}' "$OUT" | sort -n | awk '{a[NR] = $1} END {print (NR % 2 ? a[(NR + 1) / 2] : (a[NR / 2] + a[NR / 2 + 1]) / 2)}')
awk -F, -v since="$BENCH_START" -v swap0="$SWAP_START" -v oom="${oom:-0}" -v median="$median" \
  -v pauses="$(grep -c pausing "$BENCH_DIR/guardrails.log" 2>/dev/null || echo 0)" '
NR > 1 && $1 >= since {
  if ($2 > gpu) gpu = $2; if ($4 > vram) vram = $4; if ($8 != "?" && $8 > cpu) cpu = $8
  if (minram == "" || $9 < minram) minram = $9; if ($10 > swap) swap = $10
  if ($6 == "Active" || $7 == "Active") throttled++
}
END {
  printf "Peaks: GPU %d C, CPU %.0f C, VRAM %d MiB, RAM available min %d MiB, swap max %d MiB (start %d); thermal pauses %d, throttled samples %d\n", gpu, cpu, vram, minram, swap, swap0, pauses, throttled
  if (minram < 4096) reject = reject " available RAM below 4 GiB;"
  if (swap - swap0 > 2048) reject = reject " swap grew more than 2 GiB;"
  if (oom > 0) reject = reject " a process was killed for running out of memory;"
  if (median > 600) reject = reject " median task over 10 min;"
  print (reject == "" ? "Guardrails: ok" : "Guardrails: REJECT:" reject)
}' "$TELEMETRY" | tee -a "$BENCH_DIR/guardrails.log"
