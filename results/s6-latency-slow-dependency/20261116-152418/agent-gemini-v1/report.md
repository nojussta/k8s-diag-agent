Incident triage: LabHighLatency
Probable cause: DependencyFailure in demo/deployment/backend (confidence high)
The backend service's '/echo' endpoint is experiencing high latency (p95 > 900ms), causing cascading latency increases in the frontend service.

Evidence:
- [t10 Prometheus] The frontend service's POST /echo requests have an average duration of ~550ms, whereas other endpoints are very fast.
- [t14 Prometheus] The p95 latency for the backend service's /echo endpoint is measured at 970ms.

Suggested next steps (not executed; a person decides):
- Investigate the backend service code or resource constraints, as it is failing to process /echo requests within a reasonable timeframe.: kubectl logs -l app=podinfo -n demo

Not verified automatically (number not found in the cited tool result): explanation: 900

Method agent, model gemini-3.1-flash-lite, 14 read-only calls, 94.2 s.
