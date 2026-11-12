Incident triage: LabPodCrashLooping
Probable cause: ApplicationCrash in backend (confidence high)
Container backend is OOMKilled with exit code 137, and logs show a crash.

Evidence:
- [A1 Alert] Container backend of pod demo/backend-7b8d9f4db-qgxx8 is in waiting state (reason: CrashLoopBackOff) for at least 1 minute.
- [A1 Alert] Pod demo/backend-7b8d9f4db-qgxx8 (backend) is crash looping.
- [A2 Pod and container status] Container backend is OOMKilled with exit code 137.
- [A4 Container logs] Container logs show a crash.
- [A6 Restarts and last termination reason] Container is in a terminated state with exit code 137.
- [A7 Resource usage vs. limits] Container is in a terminated state with exit code 137.
- [A8 Service endpoints of the workload] Container is in a terminated state with exit code 137.
- [A10 Log and event history] Container is in a terminated state with exit code 137.
- [A9 Node conditions] Container is in a terminated state with exit code 137.

Suggested next steps (not executed; a person decides):
- Restart the container: kubectl restart backend-7b8d9f4db-qgxx8

Method pipeline, model apple-foundationmodel, 20 read-only calls, 11.1 s.
