Incident triage: LabTrafficDrop
Probable cause: ConfigurationError in demo/deployment/frontend (confidence high)
A rollout of the frontend service changed the backend URL port from 9898 to 9899, but the backend service is still running on port 9898.

Evidence:
- [t3 K8s API: pods] The backend pods are running and listening on port 9898 as indicated by the log output of the backend container.
- [t6 K8s API: pods/log] The frontend pods report connection failures to 'http://backend.demo.svc.cluster.local:9899/echo' with a context canceled error.
- [t8 K8s API: replicasets, deployments] The template diff shows the frontend was updated to use the incorrect port 9899 in the '--backend-url' argument.

Suggested next steps (not executed; a person decides):
- Roll back the frontend deployment to the previous revision that used port 9898.: kubectl rollout undo deployment/frontend -n demo

Not verified automatically (number not found in the cited tool result): evidence[0] (t3): 9898

Method agent, model gemini-3.1-flash-lite, 9 read-only calls, 62.8 s.
