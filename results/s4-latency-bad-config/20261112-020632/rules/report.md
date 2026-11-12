Incident triage: LabHighLatency
Probable cause: Unknown in demo/deployment/frontend (confidence low)
No rule matched the pod status, events or node conditions. The newest revision (not linked to a rule): revision 4 (change-cause: release 2.1: frontend configuration update) changed frontend args: ["--port=9898","--port-metrics=9797","--level=info","--backend-url=http://backend.demo.svc.cluster.local:9898/echo"] -> ["--port=9898","--port-metrics=9797","--level=info","--backend-url=http://backend.demo.svc.cluster.local:9899/echo"].

Evidence:
- [A2 pod status] phase=Running; frontend restarts=0
- [A3 events] 2026-11-12T00:06:33Z Warning Unhealthy x1: Readiness probe failed: Get "http://10.244.1.15:9898/readyz": dial tcp 10.244.1.15:9898: connect: connection refused
- [A5 rollout history] revision 4 (change-cause: release 2.1: frontend configuration update) changed frontend args: ["--port=9898","--port-metrics=9797","--level=info","--backend-url=http://backend.demo.svc.cluster.local:9898/echo"] -> ["--port=9898","--port-metrics=9797","--level=info","--backend-url=http://backend.demo.svc.cluster.local:9899/echo"].

Suggested next steps (not executed; a person decides):
- Investigate manually: no rule covers this case

Unknown: the fault is not one of the known failure classes R1-R7

Method rules, model none, 22 read-only calls, 0.1 s.
