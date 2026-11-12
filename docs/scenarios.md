# Scenarios

A scenario is a repeatable fault with a known cause. `labctl scenario <id>` injects it into the demo application,
waits for the matching alert, collects the fixed evidence (A1-A11), applies the rule-based baseline, runs every
configured method (`--method NAME=CMD`) against the same live fault, judges every answer and resets the cluster.

| ID | Fault | Alert | Where the cause is visible | Rule baseline |
|---|---|---|---|---|
| `s1-oomkilled` | a new backend revision needs more memory than its limit, so it is OOMKilled and crash-loops | `LabPodCrashLooping` | pod status (OOMKilled) + rollout history | recognises it (R1) |
| `s2-bad-image` | the frontend is rolled out with a non-existent image tag (6.15.99) | `LabPodNotReady` | pod status + rollout history | recognises it (R2) |
| `s3-missing-config` | the frontend references a ConfigMap that does not exist | `LabPodNotReady` | pod status + rollout history | recognises it (R3) |
| `s4-latency-bad-config` | a frontend release points `--backend-url` at port 9899, which the backend Service does not expose; nothing crashes | `LabHighLatency` (frontend) | metrics (latency) + frontend logs ("backend call failed") + rollout history; all pods healthy | does not |
| `s5-traffic-drop-upstream` | the same fault as S4, seen from the backend | `LabTrafficDrop` (backend) | in another workload (the frontend that calls the backend); the alerted backend is healthy | does not |

S1-S3 are known failure classes: a simple rule on the pod status finds them. S4 and S5 were added so that the
cause can only be found by combining the rollout history, metrics and logs (S4), and by following a dependency to
a workload other than the alerted one (S5). The fixed sequence inspects only the alerted object, so S5 tests what
an agent that chooses its own tools adds.

The demo application is `loadgen -> frontend -> backend` (podinfo), in the `demo` namespace, with the alert rules
in `env/manifests/alert-rules.yaml` (three pod-level rules and two service-level RED rules). Only alerts labelled
`lab="kursinis"` are routed to the receiver.

## File format

`scenarios/<id>/scenario.yaml`:

```yaml
id: s2-bad-image
title: "..."
status: implemented            # implemented | planned
story: >                       # why it happens, in prose
inject:                        # shell commands run in order (kubectl against the lab cluster only)
  - kubectl -n demo set image deployment/frontend frontend=ghcr.io/stefanprodan/podinfo:6.15.99
reset:                         # commands that restore the healthy baseline
  - kubectl apply -f env/manifests/demo-app.yaml
expected_alert:
  alertname: LabPodNotReady    # the alert the run waits for
  labels: {namespace: demo}
  timeout_seconds: 420
ground_truth:
  category: ImagePullError     # one of the schema categories
  component: demo/deployment/frontend (container frontend)
  root_cause: >                # prose, for humans
  key_evidence: [...]          # what a good answer should have found
  match:                       # what the automatic judge checks (below)
    category_any: ["ImagePullError"]
    component_any: ["frontend"]
    component_not: []          # optional: components that must NOT be named (e.g. the healthy alerted one)
    change_any: ["6.15.99"]
  acceptable_actions: [...]    # what a good next step looks like
```

## How a verdict is decided

`labctl judge` (and the runner, automatically) calls `scenario.Judge`. A diagnosis is **correct only when all three
parts are right**:

- **category**: `root_cause.category` is exactly one of `match.category_any` (default: `ground_truth.category`);
- **component**: `root_cause.component` contains a word from `match.component_any` and none from `component_not`
  (the workload that has to be fixed, which in S5 is not the alerted one);
- **change**: `root_cause.explanation` or `summary` names the change that triggered the fault (a word from
  `match.change_any`, e.g. `stress-memory`, `6.15.99`, `9899`).

The category alone is a symptom: "OOMKilled" without the release that raised the memory need is not a root cause.
An earlier version of the judge accepted the category or a keyword anywhere in the text; re-judged with the strict
rule, the three local-model S1 runs of 2026-11-11 go from 3/3 to 0/3, because none of them looked at the rollout
history. `labctl judge <run-dir>...` re-judges recorded runs after the ground truth changes, and
`labctl rules <run-dir>...` re-applies the rule baseline after the rules change. The judge matches keywords, so
every mismatch is also reviewed by hand; it does not check that quoted facts are true.

## Adding a scenario

1. Create `scenarios/<id>/scenario.yaml` with commands that inject the fault and reset it. Run them by hand first.
2. Make sure an alert rule fires for it (`env/manifests/alert-rules.yaml`, label `lab: kursinis`), or add a rule and
   run `make rules`.
3. Choose `category_any`, `component_any` and `change_any` so that a correct answer matches and a wrong one does not.
   The `TestShippedScenariosDeclareWhatTheJudgeNeeds` test fails if a scenario leaves them empty.
4. `make scenario S=<id>` to check inject, alert and reset without a model, then `make eval S=<id> N=3`.
