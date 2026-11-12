Incident triage: LabPodCrashLooping
Probable cause: OOMKilled in demo/deployment/backend (container backend) (confidence high)
container backend was killed for exceeding its memory limit (OOMKilled, exit code 137). It started with the newest revision: revision 4 (change-cause: release 2.0: warm in-memory cache on startup) changed backend args: ["--port=9898","--port-metrics=9797","--level=info"] -> ["--port=9898","--port-metrics=9797","--level=info","--stress-memory=256"].

Evidence:
- [A2 pod status] phase=Running; backend restarts=4 state=CrashLoopBackOff lastState=OOMKilled, exit code 137
- [A3 events] 2026-11-12T00:47:42Z Warning BackOff x8: Back-off restarting failed container backend in pod backend-7b8d9f4db-4grm2_demo(a15d81e6-47eb-4ee3-9955-7c952715bf41)
- [A5 rollout history] revision 4 (change-cause: release 2.0: warm in-memory cache on startup) changed backend args: ["--port=9898","--port-metrics=9797","--level=info"] -> ["--port=9898","--port-metrics=9797","--level=info","--stress-memory=256"].

Suggested next steps (not executed; a person decides):
- Roll back to the previous revision: kubectl -n demo rollout undo deployment/backend

Method rules, model none, 20 read-only calls, 0.1 s.
