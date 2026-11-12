Incident triage: LabTrafficDrop
Probable cause: Unknown in demo/deployment/backend (confidence low)
No rule matched the pod status, events or node conditions.

Evidence:
- [A2 pod status] phase=Running; backend restarts=0
- [A3 events] 2026-11-11T23:56:29Z Warning Unhealthy x1: Readiness probe failed: Get "http://10.244.1.10:9898/readyz": dial tcp 10.244.1.10:9898: connect: connection refused

Suggested next steps (not executed; a person decides):
- Investigate manually: no rule covers this case

Unknown: the fault is not one of the known failure classes R1-R7

Method rules, model none, 22 read-only calls, 0.0 s.
