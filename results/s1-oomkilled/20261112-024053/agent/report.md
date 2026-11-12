Incident triage: LabPodCrashLooping
Probable cause: OOMKilled in namespace/deployment/name (confidence high)
Container backend is OOMKilled, causing it to crash loop.

Evidence:
- [t1 alert] Container backend of pod demo/backend-7b8d9f4db-qgxx8 is in waiting state (reason: CrashLoopBackOff) for at least 1 minute.
- [t1 alert] Pod demo/backend-7b8d9f4db-qgxx8 (backend) is crash looping.
- [t2 t2] Container backend is OOMKilled.
- [t2 t2] Container backend is terminated.
- [t2 t2] Container backend is restarting.
- [t4 t4] Container backend is ready.

Suggested next steps (not executed; a person decides):
- Check the deployment status: kubectl get deployment demo/backend

Method agent, model apple-foundationmodel, 4 read-only calls, 38.7 s.
