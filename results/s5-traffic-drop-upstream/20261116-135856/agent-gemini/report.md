Incident triage: LabTrafficDrop
Probable cause: ConfigurationError in demo/deployment/frontend (container frontend) (confidence high)
A recent rollout of the frontend (revision 4) changed the '--backend-url' argument to point to port 9899, but the backend service is still running on port 9898.

Evidence:
- [t11 K8s API: replicasets, deployments] The backend deployment is running with args '--port=9898', '--port-metrics=9797', '--level=info', '--ui-color=#2f6f4f'.
- [t12 K8s API: replicasets, deployments] The frontend deployment revision 4 changed the '--backend-url' argument from '...:9898/echo' to '...:9899/echo'.
- [t14 Prometheus] The backend pods show 0 requests per second while the frontend pod is receiving 0.33 requests per second.
- [t8 K8s API: pods/log] The loadgen container logs show consistent 'wget: download timed out' errors, confirming connectivity issues to the backend.

Suggested next steps (not executed; a person decides):
- Revert the frontend deployment to the previous revision (3) or update the '--backend-url' argument to point to port 9898.: kubectl set env deployment/frontend --revision=3 -n demo

Method agent, model gemini-3.1-flash-lite, 14 read-only calls, 86.4 s.
