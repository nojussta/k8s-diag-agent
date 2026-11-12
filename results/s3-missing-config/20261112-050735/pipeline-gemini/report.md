Incident triage: LabPodNotReady
Probable cause: ConfigurationError in demo/deployment/frontend (confidence high)
A recent deployment revision (v56) introduced a dependency on a ConfigMap named 'frontend-config' in the 'env' field, which does not exist in the 'demo' namespace.

Evidence:
- [A2 K8s API] Pod is in 'Pending' phase with container 'frontend' failing to start due to 'configmap "frontend-config" not found' (CreateContainerConfigError).
- [A3 K8s API] Pod events confirm repeated warnings: 'Warning Failed x12: Error: configmap "frontend-config" not found'.
- [A5 K8s API] Latest deployment revision (56) added a new environment variable 'PODINFO_UI_COLOR' which references 'configMapKeyRef' to 'frontend-config', which was not present in previous revisions.

Suggested next steps (not executed; a person decides):
- Verify if the 'frontend-config' ConfigMap exists in the 'demo' namespace.: kubectl get configmap frontend-config -n demo
- Create the missing ConfigMap if it was accidentally deleted, or update the deployment to reference an existing one.: kubectl apply -f <path_to_configmap_definition>.yaml -n demo
- If the change was incorrect, roll back the deployment to the previous stable revision.: kubectl rollout undo deployment/frontend -n demo --to-revision=55

Method pipeline, model gemini-3.1-flash-lite, 19 read-only calls, 2.7 s.
