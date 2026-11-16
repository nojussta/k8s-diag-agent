Incident triage: LabTrafficDrop
Probable cause: ConfigurationError in demo/deployment/frontend (confidence high)
A recent rollout of the frontend deployment changed the --backend-url argument to use port 9899, while the backend service is configured to listen on port 9898.

Evidence:
- [t6 K8s API: replicasets, deployments] The latest revision of the frontend deployment (revision 6) updated the --backend-url argument from '...:9898/echo' to '...:9899/echo'.
- [t9 K8s API: pods/log] The backend pod logs show it is listening on port 9898.
- [t7 K8s API: pods/log] The frontend pod logs show repeated connection failures when trying to reach the backend at port 9899.

Suggested next steps (not executed; a person decides):
- Revert the frontend deployment to the previous stable version that uses the correct backend port (9898).: kubectl rollout undo deployment/frontend -n demo

Method agent, model gemini-3.1-flash-lite, 10 read-only calls, 58.4 s.
