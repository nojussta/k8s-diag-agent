# The local model (apfel)

[apfel](https://apfel.franzai.com/) (MIT, by Arthur-Ficial) exposes the on-device Apple foundation model
(about 3 billion parameters) as an OpenAI-compatible server on macOS 26 and later. The agent uses it by default
when it is installed:

| | apfel (local) | Gemini (hosted) |
|---|---|---|
| cost | none | free tier ~20 requests/day/model, then billing |
| context window | **4096 tokens** | hundreds of thousands |
| tool budget | 8 calls, 300 s | 15 calls, 180 s |
| data leaves the machine | no | yes |
| key needed | no | yes |

Nothing about the alert or the cluster leaves the computer, there is no quota, and it works offline. The price is
a small model with a very small context window.

**It runs only on a Mac:** apfel needs macOS 26 or later, Apple Silicon (M1 or newer) and Apple Intelligence
enabled. On macOS 26 every Mac has the 4096-token window; on macOS 27, Macs with M3 or newer and 12 GB+ get Apple's
larger on-device model with 8192 tokens (`apfel --model-info`). On Linux, Windows or a server, use
`--provider openai` with any OpenAI-compatible server (Ollama, llama.cpp server, vLLM): the same client, with
ordinary tool turns and no window trimming unless `LLM_CONTEXT_TOKENS` is set.

## Setup

```bash
brew install apfel
apfel --model-info        # should say "available: yes"
make agent-local S=s1-oomkilled
```

`build/agent` starts its own `apfel --serve --host 127.0.0.1` on a free port and stops it afterwards. To use a server
you already run, set `APFEL_URL`. `LOCAL_LLM_DUMP=/tmp/dump.jsonl` writes every request and response.

## The 4096-token window

System prompt, tool declarations, alert, tool results and answer all share 4096 tokens. The 10 tool declarations
alone are about 1.2k tokens. The client in `internal/agent/openai.go` therefore:

1. **Fits every request into a budget.** Tool outputs are shrunk largest-first (to 240, then 80 characters if
   needed) using a characters-per-token estimate. If apfel still answers `context_length_exceeded` the estimate is
   tightened and the request retried (up to 4 times). The alert's noisy labels (uid, generatorURL, ...) and the
   per-parameter descriptions of the tools are dropped.
2. **Sends every step as a fresh conversation** (system prompt plus one user message with the alert and all tool
   results so far), not as assistant/tool turns. apfel 1.16 answers HTTP 500 (`GenerationError -1`) when a request
   that declares tools contains an earlier tool result and the model wants to call another tool. A fresh
   conversation does not trigger it. A side effect is that the prompt stays fully under our control.
3. **Separates investigating from answering.** When tools and a response schema are in the same request apfel
   ignores the tools. So the investigation steps send tools and no schema, and when the model stops (it replies
   `DONE`), one more request sends the schema and no tools.
4. **Constrains the evidence ids.** In the answer request the schema is tightened so that
   `evidence[].tool_call` can only be an id that appears in the tool outputs. (apfel cannot enforce an enum on a
   boolean, so "requires_human_approval must be true" is a prompt rule backed by the orchestrator's validation.)

## Why the orchestrator has a Finalizer

A 3B model often asks for the same call again or keeps calling tools. After 2 consecutive rejected calls (duplicates,
or tools outside the allow-list), or when the call budget is spent, the orchestrator asks the model for the
final answer directly (`Finalizer.Final`) instead of failing. That answer goes through exactly the same validation
as any other.

## Prompts

`internal/agent/prompt.go` has the Gemini system prompt and the shorter local variants. The local answer prompt
contains a short rubric that says which evidence belongs to which category (for example "a container whose
lastState reason is OOMKilled is OOMKilled; do not choose NodeProblem just because a node is named"). It was added
because early local runs misclassified an OOMKilled pod as NodeProblem. It covers all nine categories, not one
scenario, but it was tuned while looking at S1, so S1 numbers for the local model are development results.

## Known limits

- In the evaluation of 2026-11-12 (five scenarios, see [results.md](results.md)) it was too weak for the strict
  judge (category, component and triggering change): it rarely names the change, often answers `Unknown`, and in
  pipeline mode it sometimes picks a category that the data does not support. It remains useful as the private,
  free option; the accuracy comparison uses a hosted model.

- Accuracy is clearly below hosted models (see [results.md](results.md)). It invents facts now and then, for
  example a node memory limit that is in no tool output. "Evidence is grounded" in this project means that it cites
  a tool call that ran, not that the quoted fact is true.
- It wastes calls: the recorded runs use 6-8 tool calls (Gemini used 5 on S1) with 1-3 rejected duplicates.
- A run takes 20-30 s end to end.
- Suggested commands are sometimes not real shell commands (for example a tool name). They are still marked as
  needing human approval.
