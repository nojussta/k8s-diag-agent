# Results so far

Everything below comes from the runs recorded in `results/` (all on a 3-node kind cluster, 2026-11-04 and
2026-11-11). Sample sizes are tiny. Read this as "the chain works and here is how it behaved", not as accuracy
figures.

## Baseline: fixed evidence collection (no model)

`make scenario S=s1-oomkilled`, 3 runs: the alert arrives 86.7, 93.1 and 98.9 s after the fault (mean 92.9 s), and
the fixed action set makes 20 read-only calls in well under a second. It always makes the same 20 calls; the agent's
point is to need fewer and to explain.

## MCP only, no model

`results/mcp-probe/s1-oomkilled-20261111.txt`: a scripted client (`cmd/mcp-probe`) got the deciding evidence
(`lastState.terminated.reason = OOMKilled`, and the changed argument `--stress-memory=256`) in 6 MCP calls instead
of 20. The calls were chosen by the program, not by a model, so this only shows that the tools return what is
needed.

## Agent runs

| Scenario | Model | Run | Tool calls (rejected) | Time | Tokens in / out | Category (confidence) | Verdict |
|---|---|---|---|---|---|---|---|
| S1 | `gemini-3-flash-preview` | 20261111-102848 | 5 (0) | 14.5 s | 21,290 / 1,999 | OOMKilled (high) | correct |
| S1 | `gemini-3.5-flash` | 20261111-102351 | 4 (0) | 44.3 s | 7,804 / 437 | none, daily quota exhausted (HTTP 429) | agent failed |
| S1 | apple-foundationmodel (apfel) | 20261111-132213 | 7 (2) | 28.6 s | 24,858 / 997 | OOMKilled (high) | correct |
| S1 | apple-foundationmodel | 20261111-132633 | 6 (2) | 24.0 s | 20,916 / 908 | OOMKilled (medium) | correct |
| S1 | apple-foundationmodel | 20261111-133033 | 7 (2) | 29.5 s | 24,797 / 1,072 | OOMKilled (medium) | correct |
| S2 | apple-foundationmodel | 20261111-133456 | 8 (1) | 25.8 s | 23,176 / 1,092 | Unknown (medium) | wrong |
| S3 | apple-foundationmodel | 20261111-133820 | 6 (2) | 20.2 s | 17,680 / 806 | ConfigurationError (medium) | correct |

The failed Gemini run is kept on purpose as an honest record. The local model costs nothing; the three S1 runs
used 70,571 input and 2,977 output tokens in total.

## Caveats

- **S1 is a development scenario for the local model.** Earlier local attempts (not kept) went wrong: the model saw
  OOMKilled and still answered NodeProblem; once it marked a command as not needing approval and the validator
  rejected the answer; one run started on a stale alert for a pod that no longer existed. The category rubric in the
  local answer prompt was added after that, so 3 of 3 on S1 is not an unbiased estimate. S2 and S3 were run once,
  after the prompt was fixed, but one run each says little: the model missed S2 (answered Unknown although the pod
  could not pull its image) and solved S3.
- **The verdict is lenient** (see [scenarios.md](scenarios.md)): a correct cause word in the explanation counts even if the category label is another one. In the seven runs above every diagnosis that produced an answer also carries the exact ground-truth category, except S2.
- **"Grounded" means cited, not true.** The validator checks that each evidence item cites a call that ran. The local
  model sometimes states facts that are in no tool output.
- **No comparison on equal terms yet.** The Gemini and local runs used different prompts and the Gemini S1 run used a
  different model from the default; a fair comparison needs the same scenarios, at least three runs each and a frozen
  prompt.

## Reproduce

```bash
make up && make check
make agent-local S=s1-oomkilled N=3
make summary
```
