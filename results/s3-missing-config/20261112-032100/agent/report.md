Incident triage: LabPodNotReady
Probable cause: ConfigurationError in frontend (confidence medium)
Missing configmap "frontend-config" caused the pod to be in a non-ready state.

Evidence:
- [t2 get_pod_status] configmap "frontend-config" not found
- [t3 get_rollout_history] Deployment has minimum availability.

Suggested next steps (not executed; a person decides):
- Check if the configmap "frontend-config" exists and is accessible.: kubectl get configmap frontend-config

Method agent, model apple-foundationmodel, 7 read-only calls, 28.6 s.
