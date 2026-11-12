Incident triage: LabHighLatency
Probable cause: ConfigurationError in demo/deployment/frontend (confidence high)
In revision 6, the container argument --backend-url was changed from http://backend.demo.svc.cluster.local:9898/echo to http://backend.demo.svc.cluster.local:9899/echo, causing backend HTTP requests to fail with context canceled errors and increasing response latency.

Evidence:
- [A5 K8s API: replicasets, deployments] Template diff for deployment frontend shows revision 6 updated --backend-url from http://backend.demo.svc.cluster.local:9898/echo to http://backend.demo.svc.cluster.local:9899/echo.
- [A4 K8s API: pods/log] Container logs show backend call failed error: Post http://backend.demo.svc.cluster.local:9899/echo: context canceled.
- [A11 Prometheus (application metrics)] The 95th percentile latency for frontend is ~2.29 seconds.

Suggested next steps (not executed; a person decides):
- Roll back deployment frontend to revision 5 to revert the backend URL port back to 9898.: kubectl rollout undo deployment/frontend -n demo

Method pipeline, model gemini-3.6-flash, 22 read-only calls, 39.2 s.
