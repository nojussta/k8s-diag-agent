Incident triage: LabHighLatency
Probable cause: ConfigurationError in demo/deployment/backend (confidence high)
A recent deployment (revision 28) added command-line arguments '--random-delay', '--random-delay-min=600', and '--random-delay-max=900' to the backend container, directly increasing request response times beyond the 500ms threshold.

Evidence:
- [t14 Prometheus] The 95th percentile latency for the backend service is 975ms, which explains the high latency in the frontend.
- [t7 K8s API: replicasets, deployments] The latest revision (28) of the backend deployment introduced '--random-delay' arguments, which were not present in previous versions.

Suggested next steps (not executed; a person decides):
- Revert the backend deployment to the previous revision (27) to remove the artificial delays.: kubectl rollout undo deployment/backend -n demo

Not verified automatically (number not found in the cited tool result): evidence[0] (t14): 95

Method agent, model gemini-3.1-flash-lite, 15 read-only calls, 106.1 s.
