# Results

Everything below comes from the runs recorded in `results/` on a 3-node kind cluster created on 2026-11-12. Every
run injects one fault, waits for the alert, and then **all methods diagnose the same live fault state**; answers are
judged by the strict three-part judge (category, component to fix, triggering change; see
[scenarios.md](scenarios.md)). All methods were run at no cost: the local model runs on the Mac, and Gemini was used on
its free tier. Regenerate the raw tables with `build/labctl summary --since 20261112`.

## Fully correct answers

| Scenario | Rules | Pipeline, local | Agent, local | Pipeline, Gemini | Agent, Gemini |
|---|---|---|---|---|---|
| S1 OOMKilled | 6/6 | 0/3 | 0/3 | 3/3 | 3/3 |
| S2 bad image tag | 6/6 | 0/3 | 2/3 | 3/3 | 3/3 |
| S3 missing ConfigMap | 6/6 | 0/3 | 1/3 | 3/3 | 3/3 |
| S4 wrong backend port, latency alert | 0/7 | 0/4 | 0/3 | 3/3 | 3/3 |
| S5 same fault, traffic-drop alert on the healthy backend | 0/7 | 0/4 | 0/3 | 0/3 | 1/3 |
| Mean read-only calls | 20.5 | 20.6 | 5.5 | 20.4 | 11.7 |
| Mean duration, s (rate-limit waits excluded) | 0.1 | 28.0 | 39.9 | 3.3 | 24.3 |

"Local" is the on-device Apple model through apfel (about 3B parameters, 4096-token window). "Gemini" is
`gemini-3.1-flash-lite` on the free tier with requests spaced 4.2 s apart (15 requests/minute); the Gemini columns
count only those runs (from 2026-11-12 03:55, after the forced-final-answer step and the pacing were added). The rule
baseline runs in every run, so its sample is larger. The local pipeline was re-run offline on the recorded evidence
(`labctl rerun`) after its compact digest was fixed, so all its runs use the same code; S4 and S5 have one extra
run there because the re-run also covered the runs made for Gemini.

## What it shows

- **Known failure classes (S1-S3): the model adds nothing.** Seven if-then rules over the fixed read-only sequence are
  right every time in under 0.2 s. The hosted model is equally right, slower, and costs tokens.
- **Combining sources (S4): the model helps, and the pipeline is enough.** Every pod is healthy, so no rule fires
  (the rules still name the component and the change, but not a category). The hosted model joins the frontend logs
  ("backend call failed" on port 9899), the latency and the rollout diff: 3/3 for the pipeline, 3/3 for the agent.
- **Cause in another workload (S5): only the agent can reach it, but it is not reliable yet.** The fixed sequence only
  inspects the alerted (healthy) backend, so the pipeline never finds the cause; at best it says "Unknown, probably
  upstream". The agent moved to the frontend's logs in every run, but checked the rollout history only of the backend,
  whose newest revision holds an unrelated earlier change (the removed `--stress-memory` flag); it was misled by that
  in two of three runs.
- **The local model is too weak for the strict judge** (3 fully correct answers out of 32), even with the decisive
  facts in its input. It remains the private, free option.

## Earlier runs, re-judged

The three local-model S1 runs of 2026-11-11 counted as 3/3 correct under the old, category-only judge. Re-judged with
`labctl judge`, they are 0/3: none of them called the rollout-history tool, so none named the change. The single
Gemini S1 run of that day did look at the rollout history and stays correct.

Exploratory Gemini runs on other models (`gemini-3.6-flash`, `gemini-3-flash-preview`) were mostly cut off by the
free tier's 20-requests-per-day limit and are not in the table; where they answered, they agree with it (S4 pipeline
correct, S5 pipeline wrong). One run's trace shows the agent moving from the backend to the frontend's pod status,
rollout history and logs before the quota ended it.

## Other measurements

- Alert detection latency: 100-176 s depending on the alert rule (S1 and S4 about 100 s, S2/S3 about 152 s, S5 about
  176 s); the fixed sequence takes 19-22 read-only calls in under 0.1 s.
- Messages: every model answer was delivered to the local chat stand-in (61/61).
- Automatic path: with `labctl watch` running, an S2 fault injected by hand was picked up from the receiver 2.5 min
  later, the agent ran and the message was delivered without any manual step (`results/watch/demo-20261112.txt`).
- Two S3 runs (2026-11-11 and the first of 2026-11-12) did not reset: `kubectl apply` cannot remove the env entry the
  fault adds. The reset now rolls back first; every later S3 run reset cleanly.

## Caveats

- Three runs per scenario and method are a small sample; the bachelor thesis extends the evaluation.
- The judge matches keywords; every mismatch was reviewed by hand (two S5 Gemini answers named port 9899 but blamed
  the backend, correctly judged wrong on the component).
- "Grounded" means every evidence item cites a call that really ran, not that the quoted fact is true.
- The prompts are generic and contain no scenario causes. They were revised once (rollout history as a step, related
  workloads) after the supervisor's review, before S4 and S5 were run.
