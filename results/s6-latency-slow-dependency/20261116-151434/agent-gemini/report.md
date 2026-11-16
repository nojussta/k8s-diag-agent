Incident triage: LabHighLatency
Probable cause: ConfigurationError in demo/deployment/backend (confidence high)
A recent deployment (revision 22) introduced artificial delay arguments (--random-delay-min=600, --random-delay-max=900), which directly impacts request duration.

Evidence:
- [t6 K8s API: replicasets, deployments] The backend deployment was updated with new arguments: --random-delay, --random-delay-unit=ms, --random-delay-min=600, --random-delay-max=900.
- [t10 Prometheus] The backend service 95th percentile latency is ~975ms.
- [t9 Prometheus] The frontend service 95th percentile latency is ~975ms, reflecting the latency introduced by the backend.

Suggested next steps (not executed; a person decides):
- Revert the backend deployment to the previous configuration (revision 21).: kubectl rollout undo deployment/backend -n demo

Not verified automatically (number not found in the cited tool result): evidence[1] (t10): 95; evidence[2] (t9): 95

Method agent, model gemini-3.1-flash-lite, 15 read-only calls, 114.7 s.
