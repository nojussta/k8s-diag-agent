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

- **Two interchangeable models.** A **local on-device model** through [apfel](https://apfel.franzai.com/) (the
  Apple model that ships with macOS 26+; free, offline, nothing leaves the machine) and hosted **Gemini**.
  `--provider auto` prefers the local one. See [docs/local-model.md](docs/local-model.md).
- **MCP tools, not a shell.** The model reaches the cluster only through an MCP server with 10 read-only tools.
  There is no tool that runs commands or writes anything.
- **Guarded loop.** Tool allow-list, call and time budgets, duplicate-call filter, tool output treated as
  untrusted data, and a validated answer (JSON schema, enum categories, every evidence item must cite a tool
  call that really ran, every command needs human approval).
- **Measured.** Every run writes a trace, token counts, duration and an automatic verdict against the scenario's
  ground truth, so models and prompts can be compared. See [docs/results.md](docs/results.md).

## Quick start

Without a cluster (unit tests only, about 15 s):

```bash
make test
```

With the full lab (Docker, kind, kubectl, helm and Go are needed; see [docs/running.md](docs/running.md)):

```bash
make up                          # kind cluster + monitoring + logging + receiver + demo app (6-10 min)
make check                       # 11 environment checks; wait ~3 min after `make up` first
make scenario S=s1-oomkilled     # fault -> alert -> fixed evidence collection -> reset (no model)
make agent-local S=s1-oomkilled  # same, with the agent on the local apfel model (N=3 repeats it)
make agent-gemini S=s1-oomkilled # same, with Gemini (GEMINI_API_KEY in .env)
make summary                     # results table
make down                        # delete the cluster
```

Each agent run leaves `results/<scenario>/<timestamp>/` with `diagnosis.json`, `metrics.json`, `trace.jsonl`
(every tool call), the alert and the evidence. A diagnosis looks like this (abridged):

```json
{
  "summary": "Pod backend-... is crash looping because the container exceeds its memory limit.",
  "root_cause": {"category": "OOMKilled", "component": "demo/backend", "explanation": "..."},
  "confidence": "high",
  "evidence": [{"tool_call": "t1", "source": "get_pod_status", "fact": "lastState: OOMKilled, exit code 137"}],
  "next_steps": [{"action": "Roll back the last rollout", "command": "kubectl rollout undo ...", "requires_human_approval": true}],
  "unknowns": []
}
```

## Repository layout

```
cmd/
  agent/            the agent: alert -> model + MCP tools -> diagnosis.json, metrics.json, trace.jsonl
  mcp-server/       MCP server (stdio) with the 10 read-only tools
  mcp-probe/        MCP client that lists the tools and runs a scripted investigation (no model)
  labctl/           lab CLI: check, scenario, evidence, summary, alerts, judge
  alert-receiver/   Alertmanager webhook receiver (runs in the cluster, standard library only)
internal/
  agent/            orchestrator, MCP client, Gemini client, local (OpenAI-compatible / apfel) client, prompts
  tools/            the read-only tools, input validation, output cap
  mcpserver/        MCP server around the tools (official Go SDK)
  lab/ evidence/    isolated cluster access (client-go, PromQL, LogQL) and the fixed evidence sequence
  scenario/ summary/ envcheck/ receiver/ alertmsg/
docs/               architecture, running, local model, scenarios, results, diagnosis JSON schema
env/                kind config, pinned Helm values, manifests (demo app, alert rules, receiver + RBAC)
scenarios/          fault definitions with ground truth (S1 OOMKilled, S2 bad image, S3 missing ConfigMap)
results/            recorded runs
Dockerfile.receiver FROM scratch image for the receiver
```

## Documentation

| | |
|---|---|
| [docs/architecture.md](docs/architecture.md) | components, investigation loop, tools, safety layers, answer schema |
| [docs/running.md](docs/running.md) | requirements, every make target and `labctl` command, configuration, output files, troubleshooting |
| [docs/local-model.md](docs/local-model.md) | the apfel model: why, limits (4096 tokens) and how the client works around them |
| [docs/scenarios.md](docs/scenarios.md) | scenario file format, the three scenarios, how verdicts are decided, adding one |
| [docs/results.md](docs/results.md) | what has been measured so far, with the caveats |

## Status and limits

This is a prototype. Three scenarios exist and the sample sizes are small (S1 three runs per model, S2 and S3
one run), so the results show that the chain works, not how accurate it is. The local 3B model is clearly weaker
than hosted models: it solved S1 and S3 and missed S2. The lab's cluster name and alert label use the word
`kursinis` (the Lithuanian word for "course paper", where this started); it is kept so that recorded results
stay consistent.
