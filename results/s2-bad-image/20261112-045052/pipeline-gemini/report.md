Incident triage: LabPodNotReady
Probable cause: ImagePullError in demo/deployment/frontend (confidence high)
The deployment was updated to use image version '6.15.99', which does not exist in the registry.

Evidence:
- [A2 K8s API: pods] The pod phase is Pending and the container is in 'ImagePullBackOff' state with message 'ghcr.io/stefanprodan/podinfo:6.15.99: not found'.
- [A3 K8s API: events] Multiple events confirm a failure to pull the image: 'Failed to pull image ...: not found'.
- [A5 K8s API: replicasets, deployments] The template diff shows the image was recently updated from '6.15.0' to '6.15.99'.

Suggested next steps (not executed; a person decides):
- Revert the deployment image to the previous known-good version (6.15.0).: kubectl set image deployment/frontend frontend=ghcr.io/stefanprodan/podinfo:6.15.0 -n demo

Method pipeline, model gemini-3.1-flash-lite, 19 read-only calls, 2.7 s.
