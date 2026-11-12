Incident triage: LabPodCrashLooping
Probable cause: ApplicationCrash in backend (confidence medium)
Pod is in a crash loop, likely due to a crash in the backend container.

Evidence:
- [A1 A1] Container backend of pod demo/backend-7b8d9f4db-tmhb6 is in waiting state (reason: CrashLoopBackOff) for at least 1 minute.
- [A1 A1] Pod demo/backend-7b8d9f4db-tmhb6 (backend) is crash looping.
- [A3 A3] Events for the pod (K8s API)
- [A11 A11] Service golden signals (RED) (Prometheus (application metrics)

Suggested next steps (not executed; a person decides):
- Check logs for crash details.: kubectl logs demo/backend-7b8d9f4db-tmhb6

Method pipeline, model apple-foundationmodel, 20 read-only calls, 26.8 s.
