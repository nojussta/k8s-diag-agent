Incident triage: LabTrafficDrop
Probable cause: ConfigurationError in demo/deployment/frontend (container frontend) (confidence high)
The latest revision of the frontend deployment (revision 8, rolled out at 2026-11-16T12:16:40Z) changed the --backend-url argument to point to port 9899 instead of 9898.

Evidence:
- [t5 K8s API: replicasets, deployments] The latest frontend revision 8 changed the --backend-url argument from 'http://backend.demo.svc.cluster.local:9898/echo' to 'http://backend.demo.svc.cluster.local:9899/echo'.
- [t6 K8s API: pods/log] The frontend container logs show repeated errors: 'Post http://backend.demo.svc.cluster.local:9899/echo: context canceled'.
- [t13 K8s API: pods/log] The backend container is confirmed to be listening on port 9898: 'Starting HTTP Server.', 'addr': ':9898'.

Suggested next steps (not executed; a person decides):
- Revert the frontend deployment to the previous revision to restore the correct backend URL port.: kubectl rollout undo deployment/frontend -n demo

Method agent, model gemini-3.1-flash-lite, 14 read-only calls, 110.2 s.
