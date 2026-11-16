# Running it

## Requirements

| Tool | Tested version | Note |
|---|---|---|
| Docker Desktop | 29.4 | give the VM at least 6 GB RAM |
| kind | 0.33.0 | `brew install kind` |
| kubectl | 1.37 | |
| helm | 4.3 | |
| Go | 1.27 (1.24 or newer required) | `brew install go` |
| make | | |
| apfel | 1.16 | optional, local model: `brew install apfel`; only on Apple Silicon Macs with macOS 26+ and Apple Intelligence enabled |
| Gemini API key | | optional, hosted model: https://aistudio.google.com/apikey |

Unit tests need only Go. Everything else is for the live lab.

## Model configuration

The agent reads these variables (also from `.env` in the repository root, which is git-ignored; copy
`.env.example`). Variables already set in the shell win.

| Variable | Meaning |
|---|---|
| `LLM_PROVIDER` | `auto` (default: apfel if installed, else Gemini), `apfel`, `gemini` or `openai` |
| `GEMINI_API_KEY` | key for the hosted model; sent only in an HTTP header and never printed |
| `GEMINI_MIN_INTERVAL` | minimum pause between Gemini requests, e.g. `4.2s` for the free tier's 15 requests/minute; reported as `rate_limit_wait_s` and excluded from the compared duration |
| `GEMINI_MODEL` | Gemini model id (default `gemini-3.5-flash`); the exact id is recorded in every run |
| `GEMINI_TEMPERATURE` | sampling temperature sent to Gemini (default 1.0, sent explicitly); recorded under `generation` in `metrics.json` |
| `AGENT_PROMPT` | `v1` selects the earlier agent prompt without procedure step f (Gemini provider, agent mode), only for a control comparison; anything else uses the current prompt |
| `APFEL_URL` | use an apfel server that is already running (default: the agent starts its own on a free loopback port) |
| `OPENAI_BASE_URL`, `OPENAI_MODEL`, `OPENAI_API_KEY` | provider `openai`: any OpenAI-compatible server (Ollama, llama.cpp server, vLLM), also on Linux/Windows |
| `LLM_CONTEXT_TOKENS` | provider `openai`: trim tool results to fit this window (0 = no trimming) |
| `NOTIFY_WEBHOOK_URL` | chat webhook for the report (Slack-style `{"text": ...}`); empty = `report.md` only |
| `LOCAL_LLM_DUMP` | file that receives every local-model request and response as JSONL (debugging) |
| `LAB_ROOT` | repository root, if the binaries are run from elsewhere |

The in-cluster agent (`cmd/diag-agent`) reads the model variables above from its Deployment
(`env/manifests/diag-agent.yaml`: provider `gemini`, model `gemini-3.1-flash-lite`, `GEMINI_MIN_INTERVAL=4.2s`) and
from the Secret `diag-agent-llm`, plus these:

| Variable | Meaning |
|---|---|
| `AUTO_INVESTIGATE` | `true`: investigate every firing alert; anything else (manifest default `false`): only record alerts. Set by `make agent-on` / `make agent-off` |
| `QUEUE_BUDGET`, `QUEUE_WINDOW` | investigations accepted per window (default 5 per `10m`); alerts over the budget are reported in one "Alert storm" message |
| `LLM_PROVIDER` | `gemini` (default) or `openai` (an OpenAI-compatible server in or outside the cluster); apfel is not available in a pod |
| `NOTIFY_WEBHOOK_URL` | chat webhook; empty = the built-in stand-in (`POST /sink`, read back with `GET /messages`) |
| `WORK_DIR` | writable directory for the investigation files (default `/work`, the only writable path in the pod) |
| `MCP_SERVER` | path of the MCP server binary (default `/mcp-server`) |

With `AUTO_INVESTIGATE=true` the process fails at start-up, not at the first alert, if the model is not configured
(for example a missing key).

## Make targets

```bash
make up          # cluster, repos, monitoring, logging, in-cluster agent, rules, app (6-10 min the first time)
make check       # 11 checks, all must PASS (Prometheus needs ~3 min after `make up` to load its rules)
make build       # build labctl, mcp-server, mcp-probe, agent into build/
make test        # go vet + go test (no cluster needed)

make scenario S=s1-oomkilled      # inject the fault, wait for the alert, collect the fixed evidence, reset
make agent S=s1-oomkilled         # same, with the agent (auto provider); N=3 repeats the run
make agent-local S=s1-oomkilled   # force the local apfel model
make agent-gemini S=s1-oomkilled  # force Gemini
make eval S=s4-latency-bad-config N=3         # rules + pipeline + agent loop (local model) on the same fault
make eval-gemini S=s5-traffic-drop-upstream   # pipeline + agent loop with Gemini (~1 + 5-12 requests per run)
make eval-all [N=3]               # all seven scenarios x N runs, rules + pipeline + agent on the local model
make watch [P=gemini]             # automatic path on the host: every new alert -> agent -> chat message

make image                        # build the agent image (diag-agent + mcp-server, FROM scratch), load it into kind
make agent-deploy                 # image + apply env/manifests/diag-agent.yaml; the agent only records alerts
make receiver                     # alias of agent-deploy (the agent pod is also the alert receiver)
make agent-secret                 # GEMINI_API_KEY (and NOTIFY_WEBHOOK_URL, if set) from .env -> Secret diag-agent-llm
make agent-on [BUDGET=5]          # agent-secret + AUTO_INVESTIGATE=true: every firing alert is investigated
make agent-off                    # AUTO_INVESTIGATE=false: only record alerts (needed for make eval*)
make agent-logs                   # last 100 log lines of the in-cluster agent
make incluster S=s4-latency-bad-config [N=3]  # scenario diagnosed by the in-cluster agent alone (after agent-on)
make storm [N=20]                 # N synthetic alerts at once to the in-cluster agent (after agent-on)
make netpol-check                 # probe pod with the agent's label: PASS/FAIL for allowed and blocked destinations
make probe                        # MCP only, no model: list the tools and run a scripted investigation
                                  # (needs a fault left in place: labctl scenario ... --keep)
make summary                      # results table (Markdown) from results/
make alerts                       # what the receiver got
make evidence                     # fixed evidence for the latest alert
make status                       # pods in the lab namespaces
make grafana | prometheus | alertmanager   # port-forward the UIs (Grafana login: admin / kursinis)
make down                         # delete the cluster
```

All commands use a separate kubeconfig, `.kube/config`. Your `~/.kube/config` is neither used nor modified. To use
`kubectl` by hand:

```bash
export KUBECONFIG=$PWD/.kube/config
kubectl --context kind-kursinis get pods -n demo
```

## labctl

```
labctl check
labctl scenario <id> [--repeat N] [--keep] [--incluster] [--method NAME=CMD]...
labctl evidence (--latest | --alert FILE) [--out DIR]
labctl summary [--since STAMP] [scenario ...]
labctl alerts
labctl judge <run-dir>...
labctl rules <run-dir>...
labctl rerun --method NAME=CMD <run-dir>...
labctl watch --method CMD [--notify URL] [--once]
labctl storm [N]
```

- `scenario` runs: baseline check, the scenario's `prepare` commands (if any), inject, wait for the alert, collect the fixed evidence, apply the rule baseline,
  run every `--method` against the same fault, judge, reset. `--keep` leaves the fault in place (needed for
  `make probe`). In a method command `{alert}`, `{out}` (the method's directory), `{evidence}` (the fixed
  sequence's evidence.json) and `{notify}` (a local stand-in chat webhook that records the message) are replaced.
  `--agent-cmd CMD` is shorthand for `--method agent=CMD`.
- `scenario --incluster` starts nothing from the workstation: it waits (up to 8 minutes) until the in-cluster agent
  has investigated the alert on its own, copies its files into `agent-incluster/` and judges the diagnosis. It
  needs `make agent-on`; with other `--method` flags in the same run, those methods run afterwards. Switch back
  with `make agent-off` before `make eval*`, otherwise the in-cluster agent also investigates every alert.
- `summary` prints the detection table and a per-method table (correct = category, component and change).
- `judge` re-judges recorded runs; `rules` re-applies the rule baseline (see [scenarios.md](scenarios.md));
  `rerun` re-runs an offline method (for example the pipeline) on recorded evidence.
- `watch` polls the receiver, investigates every new firing lab alert once (30 min duplicate filter, one at a
  time) and posts the report to `--notify`; without a webhook the messages are printed. It runs the agent on the
  host; the in-cluster agent (`make agent-on`) does the same inside the cluster and adds the budget.
- `storm [N]` (default 20, 1-200) sends N synthetic firing alerts in one notification to the in-cluster agent and
  waits until its queue is empty. The alerts name pods that do not exist, so every investigation is short. It
  prints how many were investigated, how many the budget refused and how many "Alert storm" messages were sent,
  and writes `results/storm/<timestamp>/storm.json`.

## Running the agent by hand

```bash
make build
build/agent --alert results/s1-oomkilled/<run>/alert.json --out /tmp/out [--provider apfel|gemini|openai] [--model ID] \
            [--mode agent|pipeline --evidence results/.../evidence.json] [--notify URL]
```

`--mode pipeline` does not use tools: the model gets the fixed sequence's evidence and is called once.

`--alert` takes one Alertmanager alert or a whole webhook payload (the first alert is used). The cluster must be
reachable through `.kube/config`. `--server` names another `mcp-server` binary (default `build/mcp-server`).

## The agent in the cluster

```bash
make agent-deploy      # part of `make up`
make agent-on          # needs GEMINI_API_KEY in .env
make incluster S=s1-oomkilled
make agent-logs
make agent-off
```

The pod's HTTP endpoints are reached through the API server proxy (Service `alert-receiver`, namespace
`diag-agent`, port 8080): `GET /investigations` lists every queued, running, finished or skipped alert,
`GET /investigations/{id}/{file}` returns `trace.jsonl`, `diagnosis.json`, `metrics.json` or `report.md` of one
investigation, `GET /messages` returns what was posted to the built-in chat stand-in. The files live in an
`emptyDir`, so they are lost when the pod restarts; `labctl scenario --incluster` copies them into `results/`.

## Reproducing the evaluation series

`scripts/eval-series.sh <series>` runs the series of 2026-11-16 with Gemini (`GEMINI_MODEL` defaults to
`gemini-3.1-flash-lite`, `GEMINI_MIN_INTERVAL` to `4.2s`). `N` overrides the number of runs.

| Series | What it runs |
|---|---|
| `cross` | S5 and S6, 10 runs each: pipeline, agent, and the agent with the earlier prompt (`AGENT_PROMPT=v1`) |
| `injection` | S7, 5 runs: pipeline and agent |
| `incluster` | S1, S4, S5 and S7 once each, diagnosed by the in-cluster agent (`make agent-on` before, `make agent-off` after) |
| `known` | S1-S4, 3 runs each: agent (regression check of the current prompt) |
| `storm` | 20 alerts at once against the in-cluster queue (`BUDGET` sets the budget, default 5) |

The log goes to `results/eval-series-<date>.log`. The series stops when the provider's daily quota is used up.

## What a run leaves behind

`results/<scenario>/<timestamp>/`:

| File | Content |
|---|---|
| `alert.json`, `alert-notification.json` | the alert and the raw webhook notification |
| `evidence.json`, `evidence.md` | the fixed 20-call evidence collection (baseline) |
| `rules/`, `agent/`, `pipeline/`, ... | one directory per method: `diagnosis.json` (validated answer), `metrics.json` (below), `trace.jsonl`, `report.md`, `notification.json` (the message the chat stand-in received), `output.txt` |
| `agent-incluster/` | the same files copied from the in-cluster agent, plus `investigation.json` (its queue entry: id, alert, received, started, finished, status, notified) |
| `run.json` | timings (inject, alert, detection latency) and, per method, exit code, duration, the three-part verdict (with `injection_followed` in scenarios that plant an instruction) and whether the message was delivered |

`results/storm/<timestamp>/storm.json` (from `labctl storm`): `sent`, `investigated`, `done`, `failed`, `skipped`,
`storm_messages`, `queue_drained_s`, `sum_investigation_s` and the queue `items`.

Fields of `metrics.json` besides the counters (tool calls, model steps, rejected calls, repairs, duration, tokens,
stop reason, model id):

| Field | Content |
|---|---|
| `prompt` | first 8 hex digits of the SHA-256 of the system prompt, so runs made with different prompts are not mixed up |
| `generation` | the sampling parameters the client sends (Gemini: temperature 1.0 unless `GEMINI_TEMPERATURE` is set; the OpenAI-compatible clients: temperature 0) |
| `truncated_results` | number of tool results cut by the 8 KB cap |
| `grounding` | the content check of the accepted answer: `facts`, `facts_grounded`, `numbers`, `numbers_grounded` and the list `ungrounded` (for example `evidence[1] (t3): 9899`). A miss does not reject the answer; it also appears in the chat message as "Not verified automatically" |

In `trace.jsonl` every `tool_call` event holds the full tool `result` and a `truncated` flag, so a run can be
audited offline (what exactly did the model see).

Runs recorded before 2026-11-12 have a single agent's files directly in the run directory (`agent` in `run.json`).

## Troubleshooting

- **`docker pull` or `kind create` hangs on "error getting credentials".** The Docker Desktop credential helper
  is waiting for a Keychain answer. For public images an empty config works:
  `mkdir -p /tmp/dockercfg && echo '{}' > /tmp/dockercfg/config.json && ln -s ~/.docker/contexts /tmp/dockercfg/contexts && export DOCKER_CONFIG=/tmp/dockercfg`.
- **`make check` fails right after `make up`.** Prometheus needs about 3 minutes to load rules and targets. Retry.
- **A scenario waits for "baseline not healthy".** A previous run was interrupted. Reset by hand
  (`kubectl apply -f env/manifests/demo-app.yaml`, then `kubectl -n demo rollout restart deployment/backend`) and
  wait about 2.5 minutes: `LabPodCrashLooping` uses `max_over_time(...[2m])`, so a stale alert lingers.
  A run started on a stale alert investigates a pod that no longer exists and is not a valid measurement.
- **`apfel is not installed` / Apple Intelligence not available.** `brew install apfel`, enable Apple Intelligence in
  System Settings, check with `apfel --model-info`. See [local-model.md](local-model.md).
- **`gemini: daily quota exhausted` / HTTP 429.** The free tier allows about 20 requests a day per model and one
  investigation uses 5-12. Use another model (`GEMINI_MODEL=gemini-3-flash-preview`), wait, enable billing, or use
  the local model.
- **`make incluster` reports "the in-cluster agent did not pick the alert up".** The agent only records alerts
  until `make agent-on` is run. If it reports "skipped", the budget of the queue is used up: wait for the window
  (10 minutes by default) or raise it with `make agent-on BUDGET=10`.
- **The in-cluster agent does not start after `make agent-on`.** With `AUTO_INVESTIGATE=true` it exits when the
  model is not configured; check `make agent-logs` and that `.env` has `GEMINI_API_KEY`.
- **Grafana or Prometheus not reachable in the browser.** Use `make grafana` / `make prometheus`. The scenarios
  themselves do not need port-forwards.
