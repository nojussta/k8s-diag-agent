Incident triage: LabHighLatency
Probable cause: ConfigurationError in demo/deployment/backend (confidence high)
The backend deployment was recently updated (revision 30) with arguments --random-delay, --random-delay-min=600, and --random-delay-max=900, which causes deliberate request latency between 600ms and 900ms, exceeding the 500ms threshold.

Evidence:
- [t7 K8s API: replicasets, deployments] The backend deployment's latest revision (30) added arguments: ['--random-delay', '--random-delay-unit=ms', '--random-delay-min=600', '--random-delay-max=900'].
- [t7 K8s API: replicasets, deployments] The backend last_rollout_at was 2026-11-16T13:50:59Z, which matches the timeframe of the alert start (2026-11-16T13:52:30Z).

Suggested next steps (not executed; a person decides):
- Roll back the backend deployment to the previous stable revision.: kubectl rollout undo deployment/backend -n demo

Not verified automatically (number not found in the cited tool result): evidence[1] (t7): 52

Method agent, model gemini-3.1-flash-lite, 12 read-only calls, 158.1 s.
