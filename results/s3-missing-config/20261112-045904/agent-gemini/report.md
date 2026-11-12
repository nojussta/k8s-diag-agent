Incident triage: LabPodNotReady
Probable cause: ConfigurationError in demo/deployment/frontend (confidence high)
A recent deployment update introduced a dependency on 'frontend-config' which does not exist in the namespace.

Evidence:
- [t1 get_pod_status] Pod status is 'Pending' with reason 'CreateContainerConfigError' and message 'configmap "frontend-config" not found'.
- [t2 get_events] Events confirm 'Warning' event: 'Error: configmap "frontend-config" not found'.
- [t3 get_rollout_history] Rollout revision 52 added an environment variable using 'configMapKeyRef' for 'frontend-config'.
- [t8 query_metrics] Prometheus metric 'kube_configmap_info' returns no data for 'frontend-config'.

Suggested next steps (not executed; a person decides):
- Create the missing ConfigMap 'frontend-config' in the 'demo' namespace with the required keys.: kubectl create configmap frontend-config --from-literal=ui.color=blue -n demo
- Alternatively, roll back the deployment to the previous stable revision.: kubectl rollout undo deployment/frontend -n demo

Method agent, model gemini-3.1-flash-lite, 15 read-only calls, 89.6 s.
