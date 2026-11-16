Incident triage: LabTrafficDrop
Probable cause: ConfigurationError in demo/deployment/frontend (container frontend) (confidence high)
A recent deployment updated the `--backend-url` argument in the frontend container to use port `9899`, while the backend service is listening on port `9898`.

Evidence:
- [t7 K8s API: replicasets, deployments] The latest revision of the `frontend` deployment changed the `--backend-url` from `http://backend.demo.svc.cluster.local:9898/echo` to `http://backend.demo.svc.cluster.local:9899/echo`.
- [t8 K8s API: pods/log] The `frontend` pod logs show recurring errors: `Post "http://backend.demo.svc.cluster.local:9899/echo": context canceled`.
- [t4 K8s API: pods/log] The `backend` service is explicitly started and listening on port `9898`.

Suggested next steps (not executed; a person decides):
- Roll back the frontend deployment to the previous revision to restore the correct backend port configuration.: kubectl rollout undo deployment/frontend -n demo

Method agent, model gemini-3.1-flash-lite, 15 read-only calls, 96.8 s.
