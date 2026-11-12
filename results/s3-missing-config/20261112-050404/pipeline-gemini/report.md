Incident triage: LabPodNotReady
Probable cause: ConfigurationError in demo/deployment/frontend (confidence high)
A recent configuration update added a dependency on a ConfigMap named 'frontend-config' (via PODINFO_UI_COLOR), but this ConfigMap does not exist in the 'demo' namespace.

Evidence:
- [A2 K8s API: pods] The pod is in a 'Pending' phase with the container 'frontend' waiting due to 'configmap "frontend-config" not found'.
- [A3 K8s API: events] Multiple warning events confirm: 'Error: configmap "frontend-config" not found'.
- [A5 K8s API: deployments] Template diff reveals that the latest deployment revision added an environment variable 'PODINFO_UI_COLOR' which references the missing 'frontend-config' ConfigMap.

Suggested next steps (not executed; a person decides):
- Create the missing ConfigMap 'frontend-config' in the 'demo' namespace with the required key 'ui.color'.: kubectl create configmap frontend-config --from-literal=ui.color=blue -n demo
- Alternatively, revert the deployment to a previous revision if the configuration change was unintended.: kubectl rollout undo deployment/frontend -n demo

Method pipeline, model gemini-3.1-flash-lite, 19 read-only calls, 3.4 s.
