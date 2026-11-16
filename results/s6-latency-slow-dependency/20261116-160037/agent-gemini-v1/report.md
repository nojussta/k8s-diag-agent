Incident triage: LabHighLatency
Probable cause: ConfigurationError in demo/deployment/backend (confidence high)
A recent deployment (revision 32) added new arguments '--random-delay', '--random-delay-unit=ms', '--random-delay-min=600', and '--random-delay-max=900' to the backend container, which directly impacts response latency.

Evidence:
- [t8 K8s API: replicasets, deployments] The backend deployment's latest revision introduced arguments for a random delay between 600ms and 900ms.
- [t13 Prometheus] The 95th percentile latency of the backend service is approximately 975ms.

Suggested next steps (not executed; a person decides):
- Roll back the backend deployment to the previous revision (31) to remove the artificial delay.: kubectl rollout undo deployment/backend -n demo

Not verified automatically (number not found in the cited tool result): evidence[1] (t13): 95; evidence[1] (t13): 975

Method agent, model gemini-3.1-flash-lite, 14 read-only calls, 87.6 s.
