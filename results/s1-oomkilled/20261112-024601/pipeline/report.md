Incident triage: LabPodCrashLooping
Probable cause: OOMKilled in backend (confidence high)
Container backend is in a terminated state with reason OOMKilled.

Evidence:
- [A1 Alert] Container backend of pod demo/backend-7b8d9f4db-4grm2 is in waiting state (reason: CrashLoopBackOff) for at least 1 minute.
- [A1 Alert] Pod demo/backend-7b8d9f4db-4grm2 (backend) is crash looping.
- [A2 Pod and container status] Container backend is terminated with reason OOMKilled.
- [A2 Pod and container status] Container backend is restarted 4 times.
- [A2 Pod and container status] Container backend is in a terminated state.
- [A6 Restarts and last termination reason] Container backend is in a terminated state with reason OOMKilled.
- [A7 Resource usage vs. limits] Container backend is in a terminated state with reason OOMKilled.
- [A11 Service golden signals (RED)] Container backend is in a terminated state with reason OOMKilled.
- [A8 Service endpoints of the workload] Container backend is in a terminated state with reason OOMKilled.
- [A10 Log and event history (Loki)] Container backend is in a terminated state with reason OOMKilled.
- [A9 Node conditions] Container backend is in a terminated state with reason OOMKilled.

Suggested next steps (not executed; a person decides):
- Restart the container: kubectl restart backend

Method pipeline, model apple-foundationmodel, 20 read-only calls, 11.3 s.
