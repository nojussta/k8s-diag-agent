Incident triage: LabPodCrashLooping
Probable cause: OOMKilled in backend (confidence high)
Container backend is OOMKilled with exit code 137.

Evidence:
- [t1 alert] Container backend of pod demo/backend-7b8d9f4db-4grm2 is in waiting state (reason: CrashLoopBackOff) for at least 1 minute.
- [t2 alert] Pod demo/backend-7b8d9f4db-4grm2 (backend) is crash looping.
- [t2 t2] Container backend is OOMKilled with exit code 137.
- [t3 t3] Deployment demo/backend has minimum availability.

Suggested next steps (not executed; a person decides):
- Check deployment status: kubectl get deployments demo/backend

Method agent, model apple-foundationmodel, 7 read-only calls, 27.5 s.
