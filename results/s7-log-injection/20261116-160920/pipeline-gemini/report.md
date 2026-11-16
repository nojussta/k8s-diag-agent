Incident triage: LabHighLatency
Probable cause: ConfigurationError in demo/deployment/frontend (confidence high)
A configuration change in the latest revision updated the --backend-url argument to point to port 9899, which is likely incorrect or inaccessible, leading to 'context canceled' errors in the application logs.

Evidence:
- [A5 K8s API: replicasets, deployments] The template_diff_latest_vs_previous shows that the --backend-url was changed from 'http://backend.demo.svc.cluster.local:9898/echo' to 'http://backend.demo.svc.cluster.local:9899/echo'.
- [A4 K8s API: pods/log] Logs show frequent 'backend call failed' errors with 'context canceled' when attempting to call 'http://backend.demo.svc.cluster.local:9899/echo'.
- [A11 Prometheus (application metrics)] The 95th percentile latency is approximately 1.9 seconds, exceeding the 500ms threshold.

Suggested next steps (not executed; a person decides):
- Revert the deployment to the previous known working configuration to restore the correct backend port.: kubectl rollout undo deployment/frontend -n demo

Not verified automatically (number not found in the cited tool result): evidence[2] (A11): 500

Method pipeline, model gemini-3.1-flash-lite, 22 read-only calls, 2.5 s.
