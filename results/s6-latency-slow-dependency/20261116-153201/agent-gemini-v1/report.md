Incident triage: LabHighLatency
Probable cause: ConfigurationError in demo/deployment/backend (confidence high)
The latest revision of the backend deployment introduced arguments '--random-delay', '--random-delay-min=600', and '--random-delay-max=900', causing significant artificial response latency.

Evidence:
- [t13 K8s API: replicasets, deployments] The latest revision (26) of the 'backend' deployment added '--random-delay', '--random-delay-min=600', and '--random-delay-max=900' to the container arguments.
- [t12 Prometheus] Metrics show backend API endpoints have high latency, consistent with the injected random delay (e.g., path 'readyz' shows ~0.75s latency).
- [t10 Prometheus] The 'frontend' service shows high latency specifically for the 'echo' endpoint (0.75s), which routes to the backend.

Suggested next steps (not executed; a person decides):
- Roll back the 'backend' deployment to the previous revision to remove the random delay configuration.: kubectl rollout undo deployment/backend -n demo

Method agent, model gemini-3.1-flash-lite, 15 read-only calls, 160.1 s.
