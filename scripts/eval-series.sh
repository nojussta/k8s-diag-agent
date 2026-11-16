#!/usr/bin/env bash
# The evaluation series of 2026-11-16 (hosted model, free tier). Every run
# injects one fault and all listed methods diagnose the same live fault state.
#
#   scripts/eval-series.sh cross      S5 and S6, N runs each (default 6): pipeline, agent, agent with the earlier prompt (v1)
#   scripts/eval-series.sh injection  S7, N runs (default 3): pipeline and agent
#   scripts/eval-series.sh incluster  S5 and S7 once each, diagnosed by the agent deployed in the cluster
#   scripts/eval-series.sh known      S1-S4, one fault each, the agent three times on that fault (regression check)
#   scripts/eval-series.sh storm      20 alerts at once against the in-cluster queue
#
# A part that was interrupted can be started again: runs of today that already hold the
# part's first method are counted, and only the missing ones are made.
# The series stops when the provider's daily quota is used up.
set -uo pipefail
cd "$(dirname "$0")/.."
export KUBECONFIG="$PWD/.kube/config"
export GEMINI_MODEL="${GEMINI_MODEL:-gemini-3.1-flash-lite}"
export GEMINI_MIN_INTERVAL="${GEMINI_MIN_INTERVAL:-4.2s}"
AGENT='build/agent --alert {alert} --out {out} --notify {notify} --provider gemini'
PIPE="$AGENT --mode pipeline --evidence {evidence}"
LOG=results/eval-series-$(date +%Y%m%d).log

quota_left() { ! grep -rqs "daily quota" "$1"/*/output.txt; }

# done_runs SCENARIO METHOD: today's detected runs of the scenario that already hold the method
done_runs() {
  grep -ls "\"$2\": {" results/"$1"/$(date +%Y%m%d)-*/run.json 2>/dev/null | wc -l | tr -d ' '
}

run() { # scenario, total runs wanted, marker method, labctl arguments...
  local s=$1 n=$2 marker=$3; shift 3
  local have; have=$(done_runs "$s" "$marker")
  for i in $(seq $((have + 1)) "$n"); do
    echo "$(date +%T) $s run $i/$n" | tee -a "$LOG"
    build/labctl scenario "$s" "$@" >>"$LOG" 2>&1 || echo "$(date +%T) $s run $i failed" | tee -a "$LOG"
    local last; last=$(ls -d results/"$s"/*/ | tail -1)
    grep -h "method .*: " "$LOG" | grep -v running | tail -4
    if ! quota_left "$last"; then echo "daily quota used up; stopping" | tee -a "$LOG"; exit 3; fi
  done
}

make build >/dev/null
case "${1:-}" in
  cross)
    for s in s5-traffic-drop-upstream s6-latency-slow-dependency; do
      run $s "${N:-6}" agent-gemini --method pipeline-gemini="$PIPE" --method agent-gemini="$AGENT" \
        --method agent-gemini-v1="AGENT_PROMPT=v1 $AGENT"
    done ;;
  injection)
    run s7-log-injection "${N:-3}" agent-gemini --method pipeline-gemini="$PIPE" --method agent-gemini="$AGENT" ;;
  known)
    # One fault, three investigations of the same live state: the fault is deterministic, the model is not.
    for s in s1-oomkilled s2-bad-image s3-missing-config s4-latency-bad-config; do
      run $s 1 agent-gemini --method agent-gemini="$AGENT" --method agent-gemini-r2="$AGENT" --method agent-gemini-r3="$AGENT"
    done ;;
  incluster)
    make agent-deploy >>"$LOG" 2>&1   # the image of the current code
    make agent-on >>"$LOG" 2>&1
    for s in s5-traffic-drop-upstream s7-log-injection; do
      run $s 1 agent-incluster --incluster
    done
    make agent-off >>"$LOG" 2>&1 ;;
  storm)
    make agent-on BUDGET="${BUDGET:-5}" >>"$LOG" 2>&1
    build/labctl storm "${N:-20}" | tee -a "$LOG"
    make agent-off >>"$LOG" 2>&1 ;;
  *) sed -n '2,12p' "$0"; exit 2 ;;
esac
