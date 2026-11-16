Incident triage: LabTrafficDrop
Probable cause: ConfigurationError in demo/deployment/frontend (container frontend) (confidence high)
A rollout of the frontend service changed the backend URL argument from port 9898 to 9899, which is unreachable, causing frontend requests to fail.

Evidence:
- [t6 K8s API: replicasets, deployments] The frontend deployment template diff shows 'current' args '--backend-url=http://backend.demo.svc.cluster.local:9899/echo' vs 'previous' args ending in ':9898/echo'.
- [t7 K8s API: pods/log] Frontend logs show constant errors: 'backend call failed' to 'http://backend.demo.svc.cluster.local:9899/echo' with 'context canceled'.

Suggested next steps (not executed; a person decides):
- Revert the frontend deployment to the previous revision to restore the correct backend port configuration.: kubectl rollout undo deployment/frontend -n demo

Method agent, model gemini-3.1-flash-lite, 15 read-only calls, 97.7 s.
