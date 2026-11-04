#!/usr/bin/env bash
# Samples GPU, CPU and memory into a CSV while benchmarks run, so temperatures, VRAM, RAM and swap can be
# checked against guardrails afterwards (scripts/bench.sh reads it when TELEMETRY points at the file).
# Usage: scripts/telemetry.sh <out.csv> [interval-seconds]   (needs nvidia-smi; sensors for the CPU)
set -u
out=$1
interval=${2:-10}
[ -s "$out" ] || echo "time,gpu_c,gpu_util,vram_mib,power_w,hw_slowdown,sw_slowdown,cpu_c,mem_avail_mib,swap_used_mib,load1" > "$out"
while true; do
  gpu=$(nvidia-smi --query-gpu=temperature.gpu,utilization.gpu,memory.used,power.draw,clocks_event_reasons.hw_thermal_slowdown,clocks_event_reasons.sw_thermal_slowdown \
    --format=csv,noheader,nounits | head -1 | tr -d ' ')
  cpu=$(sensors 2>/dev/null | awk '/^Package id 0:/ {gsub(/[^0-9.]/, "", $4); print $4; exit}')
  mem=$(free -m | awk '/^Mem:/ {avail=$7} /^Swap:/ {swap=$3} END {print avail "," swap}')
  echo "$(date +%s),${gpu:-?,?,?,?,?,?},${cpu:-?},$mem,$(cut -d' ' -f1 /proc/loadavg)" >> "$out"
  sleep "$interval"
done
