# k8s-diag-agent

A **read-only incident-diagnosis agent for Kubernetes**, written in Go.

When Alertmanager fires an alert (a crash-looping pod, an image that cannot be pulled, a missing ConfigMap), the
agent investigates it the way an on-call engineer would: it looks at the pod, its events, rollout history,
logs and metrics, and answers with a **root cause, the evidence for it and a suggested next step**. It can look
but never touch: the model only gets read-only tools, and every suggested command is marked as needing a human.

```
Alertmanager --webhook--> alert receiver --> orchestrator <--> language model
                                                  |
                                                  | MCP (stdio)
                                                  v
                                          MCP server: 10 read-only tools
                                          (Kubernetes API, Prometheus, Loki)
```

The repository also contains the **experiment lab** the agent is developed and measured in: a local `kind`
cluster with Prometheus, Alertmanager, Grafana and Loki, a demo application, repeatable fault scenarios with a
known ground truth, and a runner (`labctl`) that injects a fault, waits for the alert, runs the agent, judges
the answer and resets the cluster.

## Highlights

- **Asks what the model actually adds.** Every fault is diagnosed by three methods on the same live cluster state:
  a **rule baseline** (7 rules over a fixed read-only sequence, no model), a **pipeline** (the same fixed sequence,
  then one model call) and the **agent** (the model chooses its own tools). See [docs/architecture.md](docs/architecture.md).
- **Strict judge.** An answer is correct only if the fault category, the component to fix **and** the change that
  triggered the fault are all right ("OOMKilled" alone is a symptom). See [docs/scenarios.md](docs/scenarios.md).
- **Five scenarios, two of them hard.** S1-S3 are known classes a rule recognises; in S4 and S5 every pod is healthy
  and the cause is only visible by combining metrics, logs and the rollout history, in S5 in another workload.
- **Interchangeable models.** The on-device Apple model through [apfel](https://github.com/Arthur-Ficial/apfel)
  (macOS 26+ on Apple Silicon only; free, offline, 4096-token window), hosted **Gemini**, or any
  **OpenAI-compatible** server (Ollama, llama.cpp, vLLM) on any OS. See [docs/local-model.md](docs/local-model.md).
- **MCP tools, not a shell.** The model reaches the cluster only through an MCP server with 10 read-only tools.
- **Guarded loop.** Tool allow-list, call and time budgets, duplicate-call filter, tool output treated as
  untrusted data, and a validated answer (JSON schema, every evidence item must cite a call that really ran, every
  command needs human approval).
- **Reaches the on-call engineer.** The report is posted to a chat webhook (Slack-style `{"text": ...}`,
  `NOTIFY_WEBHOOK_URL`); in experiments a local stand-in channel records every message. `make watch` is the
  automatic path: new alert at the receiver -> agent (one at a time, 30 min duplicate filter) -> message.

## Quick start

Without a cluster (unit tests only, about 15 s):

```bash
make test
```

With the full lab (Docker, kind, kubectl, helm and Go are needed; see [docs/running.md](docs/running.md)):

```bash
make up                                     # kind cluster + monitoring + logging + receiver + demo app (6-10 min)
make check                                  # 11 environment checks; wait ~3 min after `make up` first
make scenario S=s1-oomkilled                # fault -> alert -> fixed evidence + rule baseline -> reset (no model)
make eval S=s4-latency-bad-config N=3       # rules + pipeline + agent on the local model, same fault each run
make eval-gemini S=s5-traffic-drop-upstream # pipeline + agent with Gemini (GEMINI_API_KEY in .env)
make watch                                  # automatic: every new alert -> agent -> chat message
build/labctl summary --since 20261112       # per-scenario, per-method results table
make down                                   # delete the cluster
```

Each run leaves `results/<scenario>/<timestamp>/` with the alert, the fixed evidence and one directory per method
(`diagnosis.json`, `metrics.json`, `trace.jsonl`, `report.md`, `notification.json`). A diagnosis looks like this
(abridged, a real S4 answer):

```json
{
  "root_cause": {"category": "ConfigurationError", "component": "demo/deployment/frontend",
                 "explanation": "In revision 6, --backend-url was changed from ...:9898/echo to ...:9899/echo, causing backend requests to fail ..."},
  "confidence": "high",
  "evidence": [{"tool_call": "A5", "source": "rollout history", "fact": "template diff: --backend-url ...9898 -> ...9899"}],
  "next_steps": [{"action": "Roll back the frontend", "command": "kubectl -n demo rollout undo deployment/frontend", "requires_human_approval": true}]
}
```

## Repository layout

```
cmd/
  agent/            the agent: alert -> model (+ MCP tools, or --mode pipeline) -> diagnosis, report, chat message
  mcp-server/       MCP server (stdio) with the 10 read-only tools
  mcp-probe/        MCP client that lists the tools and runs a scripted investigation (no model)
  labctl/           lab CLI: check, scenario, evidence, summary, alerts, judge, rules, watch
  alert-receiver/   Alertmanager webhook receiver (runs in the cluster, standard library only)
internal/
  agent/            orchestrator, pipeline, MCP client, Gemini / apfel / OpenAI-compatible clients, prompts
  rules/            the rule baseline (R1-R7)
  notify/           report text and chat webhook
  watch/            queue and duplicate filter for the automatic path
  tools/            the read-only tools, input validation, output cap
  mcpserver/        MCP server around the tools (official Go SDK)
  lab/ evidence/    isolated cluster access (client-go, PromQL, LogQL) and the fixed evidence sequence
  scenario/ summary/ envcheck/ receiver/ alertmsg/
docs/               architecture, running, local model, scenarios, results, diagnosis JSON schema
env/                kind config, pinned Helm values, manifests (demo app, alert rules, receiver + RBAC)
scenarios/          five fault definitions with ground truth (S1-S3 known classes, S4-S5 multi-source)
results/            recorded runs
Dockerfile.receiver FROM scratch image for the receiver
```

## Documentation

| | |
|---|---|
| [docs/architecture.md](docs/architecture.md) | components, investigation loop, tools, safety layers, answer schema |
| [docs/running.md](docs/running.md) | requirements, every make target and `labctl` command, configuration, output files, troubleshooting |
| [docs/local-model.md](docs/local-model.md) | the apfel model: why, limits (4096 tokens) and how the client works around them |
| [docs/scenarios.md](docs/scenarios.md) | scenario file format, the five scenarios, the three-part judge, adding one |
| [docs/results.md](docs/results.md) | what has been measured so far, with the caveats |

## Status and limits

This is a prototype under evaluation; sample sizes are small (see [docs/results.md](docs/results.md)). The hosted
Gemini free tier allows about 20 requests per model per day, which limits how many agent runs can be made per day.
The local 3B model is too weak for the strict judge; it remains the private, free option. The agent is run by the
scenario runner or `labctl watch` on the host; deploying it inside the cluster next to the receiver is still to do.
The lab's cluster name and alert label use the word `kursinis` (the Lithuanian word for "course paper", where this
started); it is kept so that recorded results stay consistent.
