Incident triage: LabPodNotReady
Probable cause: ConfigurationError in demo/deployment/frontend (confidence high)
The deployment 'frontend' was updated to use a ConfigMap 'frontend-config' (Revision 54), but that ConfigMap does not exist in the 'demo' namespace.

Evidence:
- [t1 K8s API: pods] Pod is in 'Pending' state with message: 'configmap "frontend-config" not found' and reason 'CreateContainerConfigError'.
- [t2 K8s API: replicasets, deployments] Rollout history shows Revision 54 introduced a change to use 'frontend-config' via ConfigMapKeyRef.
- [t3 K8s API: events] Kubernetes events confirm repeated warnings: 'Error: configmap "frontend-config" not found'.
- [t6 Prometheus] Querying for existing ConfigMaps in namespace 'demo' shows only 'kube-root-ca.crt' exists, 'frontend-config' is missing.

Suggested next steps (not executed; a person decides):
- Create the missing ConfigMap 'frontend-config' in the 'demo' namespace with the required keys.: kubectl create configmap frontend-config --from-literal=ui.color=blue -n demo
- Alternatively, rollback the deployment to the previous stable revision.: kubectl rollout undo deployment/frontend -n demo

Method agent, model gemini-3.1-flash-lite, 7 read-only calls, 38.6 s.
