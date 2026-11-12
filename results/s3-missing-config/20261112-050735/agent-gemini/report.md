Incident triage: LabPodNotReady
Probable cause: ConfigurationError in demo/deployment/frontend (container frontend) (confidence high)
The latest revision (56) introduced a dependency on a ConfigMap named 'frontend-config' which does not exist in the 'demo' namespace, causing a CreateContainerConfigError.

Evidence:
- [t1 K8s API: pods] Pod 'frontend-6df678c5fd-f6cpj' has a 'CreateContainerConfigError' state with the message 'configmap "frontend-config" not found'.
- [t2 K8s API: events] Events confirm the repeated warning: 'Error: configmap "frontend-config" not found'.
- [t3 K8s API: replicasets, deployments] The latest revision (56) changed the container environment variables to include a 'configMapKeyRef' pointing to 'frontend-config'.
- [t6 Prometheus] A query for ConfigMaps in the 'demo' namespace did not return 'frontend-config'.

Suggested next steps (not executed; a person decides):
- Verify if the 'frontend-config' ConfigMap was intended to be created in the 'demo' namespace.
- Create the missing 'frontend-config' ConfigMap in the 'demo' namespace or revert the deployment to the previous revision if the configuration change was incorrect.: kubectl apply -f <path-to-configmap.yaml> -n demo OR kubectl rollout undo deployment/frontend -n demo

Method agent, model gemini-3.1-flash-lite, 8 read-only calls, 50.2 s.
