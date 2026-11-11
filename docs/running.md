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
| apfel | 1.16 | optional, local model: `brew install apfel`, macOS 26+ with Apple Intelligence |
| Gemini API key | | optional, hosted model: https://aistudio.google.com/apikey |

Unit tests need only Go. Everything else is for the live lab.

## Model configuration

The agent reads these variables (also from `.env` in the repository root, which is git-ignored; copy
`.env.example`). Variables already set in the shell win.

| Variable | Meaning |
|---|---|
| `LLM_PROVIDER` | `auto` (default: apfel if installed, else Gemini), `apfel` or `gemini` |
| `GEMINI_API_KEY` | key for the hosted model; sent only in an HTTP header and never printed |
| `GEMINI_MODEL` | Gemini model id (default `gemini-3.5-flash`); the exact id is recorded in every run |
| `APFEL_URL` | use an apfel server that is already running (default: the agent starts its own on a free loopback port) |
| `LOCAL_LLM_DUMP` | file that receives every local-model request and response as JSONL (debugging) |
| `LAB_ROOT` | repository root, if the binaries are run from elsewhere |

## Make targets

```bash
make up          # cluster, repos, monitoring, logging, receiver, rules, app (6-10 min the first time)
make check       # 11 checks, all must PASS (Prometheus needs ~3 min after `make up` to load its rules)
make build       # build labctl, mcp-server, mcp-probe, agent into build/
make test        # go vet + go test (no cluster needed)

make scenario S=s1-oomkilled      # inject the fault, wait for the alert, collect the fixed evidence, reset
make agent S=s1-oomkilled         # same, with the agent (auto provider); N=3 repeats the run
make agent-local S=s1-oomkilled   # force the local apfel model
make agent-gemini S=s1-oomkilled  # force Gemini
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
labctl scenario <id> [--repeat N] [--keep] [--agent-cmd CMD]
labctl evidence (--latest | --alert FILE) [--out DIR]
labctl summary [scenario ...]
labctl alerts
labctl judge <run-dir>
```

- `scenario` runs: baseline check, inject, wait for the alert, collect the fixed evidence, optionally run the agent,
  reset. `--keep` leaves the fault in place (needed for `make probe`). `--agent-cmd` is a command line in which
  `{alert}` and `{out}` are replaced by the alert file and the run directory.
- `judge` compares a run's `diagnosis.json` with the scenario's ground truth (see [scenarios.md](scenarios.md)).

## Running the agent by hand

```bash
make build
build/agent --alert results/s1-oomkilled/<run>/alert.json --out /tmp/out [--provider apfel|gemini] [--model ID]
```

`--alert` takes one Alertmanager alert or a whole webhook payload (the first alert is used). The cluster must be
reachable through `.kube/config`.

## What a run leaves behind

`results/<scenario>/<timestamp>/`:

| File | Content |
|---|---|
| `alert.json`, `alert-notification.json` | the alert and the raw webhook notification |
| `evidence.json`, `evidence.md` | the fixed 20-call evidence collection (baseline) |
| `diagnosis.json` | the agent's validated answer (absent if the run failed) |
| `metrics.json` | tool calls, rejected calls, repairs, duration, input/output tokens, stop reason, model id |
| `trace.jsonl` | one line per event: alert, every tool call (id, name, arguments, result size), forced final, final, end |
| `agent-output.txt` | the agent's console output |
| `run.json` | timings (inject, alert, detection latency), the agent's exit code and the judge's verdict |

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
- **Grafana or Prometheus not reachable in the browser.** Use `make grafana` / `make prometheus`. The scenarios
  themselves do not need port-forwards.
