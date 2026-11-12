Incident triage: LabPodNotReady
Probable cause: ConfigurationError in frontend (confidence medium)
Missing configmap caused the pod to be in a non-ready state.

Evidence:
- [t2 get_pod_status] configmap "frontend-config" not found
- [t3 get_rollout_history] Deployment has minimum availability.
- [t2 get_pod_status] Container ready but configmap missing.

Suggested next steps (not executed; a person decides):
- Check configmap availability: kubectl get configmap frontend-config

Unknown: Pod is not ready because it cannot start containers.

Method agent, model apple-foundationmodel, 6 read-only calls, 20.6 s.
