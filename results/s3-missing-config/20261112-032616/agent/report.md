Incident triage: LabPodNotReady
Probable cause: Unknown in demo/frontend (confidence medium)
Pod is not ready, but there is no clear indication of the cause from the available information.

Evidence:
- [t2 get_pod_status] Pod demo/frontend-6df678c5fd-hgvmg is not ready
- [t1 alert] Pod demo/frontend-6df678c5fd-hgvmg has been in a non-ready phase for longer than 2 minutes
- [t3 get_rollout_history] Deployment demo/frontend has minimum replicas available

Suggested next steps (not executed; a person decides):
- Check Deployment Status: kubectl describe deployment demo/frontend

Unknown: Pod readiness issue not fully resolved

Method agent, model apple-foundationmodel, 6 read-only calls, 42.2 s.
