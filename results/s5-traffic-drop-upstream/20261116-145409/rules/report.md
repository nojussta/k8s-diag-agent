Incident triage: LabTrafficDrop
Probable cause: Unknown in demo/deployment/backend (confidence low)
No rule matched the pod status, events or node conditions. The newest revision (not linked to a rule): revision 18 (change-cause: release 2.0.1: UI colour) changed backend args: ["--port=9898","--port-metrics=9797","--level=info"] -> ["--port=9898","--port-metrics=9797","--level=info","--ui-color=#2f6f4f"].

Evidence:
- [A2 pod status] phase=Running; backend restarts=0
- [A3 events] 2026-11-16T12:54:10Z Warning Unhealthy x1: Readiness probe failed: Get "http://10.244.1.47:9898/readyz": dial tcp 10.244.1.47:9898: connect: connection refused
- [A5 rollout history] revision 18 (change-cause: release 2.0.1: UI colour) changed backend args: ["--port=9898","--port-metrics=9797","--level=info"] -> ["--port=9898","--port-metrics=9797","--level=info","--ui-color=#2f6f4f"].

Suggested next steps (not executed; a person decides):
- Investigate manually: no rule covers this case

Unknown: the fault is not one of the known failure classes R1-R7

Method rules, model none, 22 read-only calls, 0.1 s.
