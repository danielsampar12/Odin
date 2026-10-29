#!/usr/bin/env bash
# Cheap code-quality signals for a finished scripts/bench.sh run, since passing tests says nothing about the
# code itself: go vet, gofmt, and how many lines of non-test Go the agent changed. Each run's repo still has
# the task commit, so this reads the agent's change from git.
# Usage: scripts/bench-quality.sh <BENCH_DIR>   (writes <BENCH_DIR>/quality.tsv and prints a summary)
set -u
export LC_ALL=C
dir=$(cd "$1" && pwd) || exit 1
out=$dir/quality.tsv
printf 'task\trun\tpass\tvet_ok\tgofmt_ok\tgo_lines\n' > "$out"
tail -n +2 "$dir/results.tsv" | while IFS=$'\t' read -r task run pass _; do
  repo=$dir/$task-$run
  [ -d "$repo/.git" ] || continue
  (
    cd "$repo" || exit
    vet=no gofmt_ok=no
    go vet ./... > /dev/null 2>&1 && vet=yes
    # Agent-written Go outside the tests: changed tracked files plus new files.
    files=$( { git diff --name-only --diff-filter=d HEAD -- '*.go'; git ls-files -o --exclude-standard -- '*.go'; } | grep -v '_test\.go$' | sort -u)
    # gofmt exits non-zero on a file that does not parse, which counts as unformatted.
    if [ -z "$files" ] || { unformatted=$(echo "$files" | xargs gofmt -l 2> /dev/null) && [ -z "$unformatted" ]; }; then
      gofmt_ok=yes
    fi
    lines=$(git diff --numstat HEAD -- '*.go' ':!*_test.go' | awk '{s+=$1+$2} END{print s+0}')
    for f in $(git ls-files -o --exclude-standard -- '*.go' | grep -v '_test\.go$'); do lines=$((lines + $(wc -l < "$f"))); done
    printf '%s\t%s\t%s\t%s\t%s\t%s\n' "$task" "$run" "$pass" "$vet" "$gofmt_ok" "$lines"
  ) >> "$out"
done
pass=$(awk -F'\t' '$3=="yes"' "$out")
[ -n "$pass" ] || { echo "no passing runs"; exit 0; }
n=$(echo "$pass" | wc -l)
median=$(echo "$pass" | cut -f6 | sort -n | awk '{a[NR]=$1} END {print (NR % 2) ? a[(NR+1)/2] : (a[NR/2] + a[NR/2+1]) / 2}')
echo "passing runs $n: go vet clean $(echo "$pass" | awk -F'\t' '$4=="yes"' | wc -l)/$n," \
  "gofmt clean $(echo "$pass" | awk -F'\t' '$5=="yes"' | wc -l)/$n, median non-test Go lines changed $median"
