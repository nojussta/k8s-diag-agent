Incident triage: LabHighLatency
Probable cause: ConfigurationError in demo/deployment/backend (confidence high)
The latest rollout of the backend deployment added random delay arguments (--random-delay-min=600 --random-delay-max=900) which causes increased latency for all dependent services like frontend.

Evidence:
- [t8 K8s API: replicasets, deployments] The latest backend rollout revision 24 added arguments '--random-delay', '--random-delay-unit=ms', '--random-delay-min=600', '--random-delay-max=900'.
- [t8 K8s API: replicasets, deployments] The last_rollout_at for backend is '2026-11-16T13:25:23Z', which aligns with the start of the latency spike.

Suggested next steps (not executed; a person decides):
- Roll back the backend deployment to the previous revision to remove the random delay configuration.: kubectl rollout undo deployment/backend -n demo

Method agent, model gemini-3.1-flash-lite, 12 read-only calls, 93.6 s.
