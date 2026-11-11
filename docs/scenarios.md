# Scenarios

A scenario is a repeatable fault with a known cause. `labctl scenario <id>` injects it into the demo application,
waits for the matching alert, collects the evidence (and runs the agent if `--agent-cmd` is given), then resets the
cluster. The scenario's `ground_truth` is what the agent's answer is judged against.

| ID | Fault | Alert | Ground-truth category |
|---|---|---|---|
| `s1-oomkilled` | a new backend revision needs more memory than its limit, so it is OOMKilled and crash-loops | `LabPodCrashLooping` | `OOMKilled` |
| `s2-bad-image` | the frontend is rolled out with a non-existent image tag (6.15.99), the pod cannot pull it | `LabPodNotReady` | `ImagePullError` |
| `s3-missing-config` | the frontend references a ConfigMap that does not exist, the container never starts | `LabPodNotReady` | `ConfigurationError` |

The demo application is `loadgen -> frontend -> backend` (podinfo), in the `demo` namespace, with the alert rules
in `env/manifests/alert-rules.yaml`. Only alerts labelled `lab="kursinis"` are routed to the receiver.

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
  match:                       # machine-readable verdict rules (below)
    cause_any: ["ImagePullBackOff", "ErrImagePull", "image", "tag"]
    component_any: ["frontend"]
  acceptable_actions: [...]    # what a good next step looks like
```

## How a verdict is decided

`labctl judge` (and the runner, automatically) calls `scenario.Judge`:

- the **cause** matches when the answer's `root_cause.category` equals `ground_truth.category`, **or** its category,
  explanation or summary contains any word from `match.cause_any` (case-insensitive);
- the **component** matches when `root_cause.component` contains any word from `match.component_any`;
- the answer is **correct** only if both match.

This is deliberately lenient about the category label: an answer that is labelled `ApplicationCrash` but explains
that the container was OOMKilled counts as correct. To count only exact category matches, compare
`diagnosis.json` against `ground_truth.category` yourself (`docs/results.md` reports both). The judge does not check
that the quoted facts are true.

## Adding a scenario

1. Create `scenarios/<id>/scenario.yaml` with commands that inject the fault and reset it. Run them by hand first.
2. Make sure an alert rule fires for it (`env/manifests/alert-rules.yaml`, label `lab: kursinis`), or add a rule and
   run `make rules`.
3. Choose `match.cause_any` and `component_any` words that appear in a correct answer but not in a wrong one.
4. `make scenario S=<id>` to check inject, alert and reset without a model, then `make agent-local S=<id> N=3`.
