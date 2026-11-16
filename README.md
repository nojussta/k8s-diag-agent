# k8s-diag-agent

A **read-only incident-diagnosis agent for Kubernetes**, written in Go.

When Alertmanager fires an alert (a crash-looping pod, an image that cannot be pulled, a missing ConfigMap), the
agent investigates it the way an on-call engineer would: it looks at the pod, its events, rollout history,
logs and metrics, and answers with a **root cause, the evidence for it and a suggested next step**. It can look
but never touch: the model only gets read-only tools, and every suggested command is marked as needing a human.

```
                 +------------ diag-agent pod (read-only identity) ------------+
Alertmanager --> | webhook receiver --> queue (budget, --> orchestrator <------|--> language model
   (webhook)     |                      duplicate filter)       |              |
                 |                                              | MCP (stdio)  |--> chat webhook
                 |                                              v              |
                 |                               MCP server: 10 read-only tools|
                 +----------------------------------------------|--------------+
                                                                v
                                                Kubernetes API, Prometheus, Loki
```

The same investigation code also runs from the command line (`build/agent`, one alert), which is how the lab
compares methods on the same fault.

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
- **Seven scenarios, each answering a design question.** S1-S3 are known classes a rule recognises (does the model
  add anything?). In S4-S6 every pod is healthy and the cause is only visible by combining metrics, logs and the
  rollout history; in S5 and S6 it is in another workload, next to an unrelated harmless release of the alerted
  one (is the loop needed?). S7 plants an instruction for the agent in a container log (indirect prompt injection).
- **Interchangeable models.** The on-device Apple model through [apfel](https://github.com/Arthur-Ficial/apfel)
  (macOS 26+ on Apple Silicon only; free, offline, 4096-token window), hosted **Gemini**, or any
  **OpenAI-compatible** server (Ollama, llama.cpp, vLLM) on any OS. See [docs/local-model.md](docs/local-model.md).
- **MCP tools, not a shell.** The model reaches the cluster only through an MCP server with 10 read-only tools.
- **Guarded loop.** Tool allow-list, call and time budgets, duplicate-call filter, tool output treated as
  untrusted data, and a validated answer (JSON schema, every evidence item must cite a call that really ran, every
  command needs human approval).
- **Content check of the evidence.** Every number an evidence item quotes must occur in the tool result it cites.
  A miss never rejects the answer; it is recorded in `metrics.json` and shown to the engineer as
  "Not verified automatically".
- **Runs in the cluster, hardened.** One pod receives the Alertmanager webhook, queues the alerts (5 investigations
  per 10 minutes by default, one at a time, one summary message for an alert storm) and investigates them. It runs
  non-root from a `FROM scratch` image with a read-only root filesystem, a read-only ClusterRole and a
  NetworkPolicy; the API key comes from a Secret.
- **Reaches the on-call engineer.** The report is posted to a chat webhook (Slack-style `{"text": ...}`,
  `NOTIFY_WEBHOOK_URL`); in experiments a stand-in channel records every message.

## Quick start

Without a cluster (unit tests only, about 15 s):

```bash
make test
```

With the full lab (Docker, kind, kubectl, helm and Go are needed; see [docs/running.md](docs/running.md)):

```bash
make up                                     # kind cluster + monitoring + logging + in-cluster agent + demo app (6-10 min)
make check                                  # 11 environment checks; wait ~3 min after `make up` first
make scenario S=s1-oomkilled                # fault -> alert -> fixed evidence + rule baseline -> reset (no model)
make eval S=s4-latency-bad-config N=3       # rules + pipeline + agent on the local model, same fault each run
make eval-gemini S=s5-traffic-drop-upstream # pipeline + agent with Gemini (GEMINI_API_KEY in .env)
make agent-on                               # the in-cluster agent investigates every firing alert on its own
make incluster S=s4-latency-bad-config      # fault -> the in-cluster agent diagnoses it alone -> judged -> reset
make storm N=20                             # 20 synthetic alerts at once: what the queue and its budget do
make netpol-check                           # prove the NetworkPolicy is enforced (allowed and blocked destinations)
make agent-off                              # back to recording alerts only (needed for make eval*)
make watch                                  # the same automatic path, with the agent running on the host
build/labctl summary --since 20261112       # per-scenario, per-method results table
scripts/eval-series.sh cross                # the evaluation series of 2026-11-16 (cross | injection | incluster | known | storm)
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
  diag-agent/       the in-cluster agent: Alertmanager webhook receiver + queue + worker in one process
  labctl/           lab CLI: check, scenario, evidence, summary, alerts, judge, rules, rerun, watch, storm
internal/
  investigate/      one investigation from start to finish; shared by cmd/agent and cmd/diag-agent
  agent/            orchestrator, pipeline, MCP client, model clients, prompts, answer validation and content check
  rules/            the rule baseline (R1-R7)
  notify/           report text and chat webhook
  watch/            queue with a budget and duplicate filter for the automatic path
  tools/            the read-only tools, input validation, output cap
  mcpserver/        MCP server around the tools (official Go SDK)
  lab/ evidence/    isolated cluster access (client-go, PromQL, LogQL) and the fixed evidence sequence
  scenario/ summary/ envcheck/ receiver/ alertmsg/
docs/               architecture, running, local model, scenarios, results, diagnosis JSON schema
env/                kind config, pinned Helm values, manifests (demo app, alert rules, in-cluster agent + RBAC + NetworkPolicy)
scenarios/          seven fault definitions with ground truth (S1-S3 known classes, S4-S6 multi-source, S7 log injection)
scripts/            eval-series.sh: reproduces the evaluation series
results/            recorded runs
Dockerfile.agent    FROM scratch image of the in-cluster agent (diag-agent + mcp-server)
LICENSE             MIT
```

## Documentation

| | |
|---|---|
| [docs/architecture.md](docs/architecture.md) | components, investigation loop, tools, safety layers, threat model, deployment, answer schema |
| [docs/running.md](docs/running.md) | requirements, every make target and `labctl` command, configuration, output files, troubleshooting |
| [docs/local-model.md](docs/local-model.md) | the apfel model: why, limits (4096 tokens) and how the client works around them |
| [docs/scenarios.md](docs/scenarios.md) | scenario file format, the seven scenarios, the three-part judge, the injection check, adding one |
| [docs/results.md](docs/results.md) | what has been measured so far, with the caveats |

## Status and limits

This is a prototype under evaluation; sample sizes are small and the measured outcomes are in
[docs/results.md](docs/results.md). The hosted Gemini free tier allows about 20 requests per model per day, which
limits how many agent runs can be made per day.

- **The local model through apfel works only on Apple Silicon Macs with macOS 26 or newer and Apple Intelligence
  enabled.** On any other machine use Gemini or an OpenAI-compatible server. The in-cluster agent cannot use apfel
  (it runs in a Linux container); it uses Gemini or an OpenAI-compatible server.
- **With a hosted model, cluster data leaves the cluster.** Container logs, events and metrics that the tools
  return are sent to the model provider as they are; there is no redaction layer.
- **The NetworkPolicy cannot restrict egress by host name.** It allows HTTPS to any address outside the private
  ranges; narrowing it to the provider's name needs a CNI with FQDN policies (for example Cilium).
- The in-cluster queue and duplicate filter are in memory (one replica); a restart forgets them.

The lab's cluster name and alert label use the word `kursinis` (the Lithuanian word for "course paper", where this
started); it is kept so that recorded results stay consistent.

## License

MIT, see [LICENSE](LICENSE).
