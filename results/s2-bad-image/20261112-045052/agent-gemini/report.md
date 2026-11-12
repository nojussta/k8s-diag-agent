Incident triage: LabPodNotReady
Probable cause: ImagePullError in demo/deployment/frontend (confidence high)
A recent deployment update (revision 48) changed the image tag to '6.15.99', which is invalid or unavailable.

Evidence:
- [t1 K8s API: pods] Pod 'frontend-68747558f7-8bzbd' is in state 'ImagePullBackOff' with error message 'ghcr.io/stefanprodan/podinfo:6.15.99: not found'.
- [t2 K8s API: replicasets, deployments] Deployment 'frontend' rollout history shows revision 48 changed the image from 'ghcr.io/stefanprodan/podinfo:6.15.0' to 'ghcr.io/stefanprodan/podinfo:6.15.99'.
- [t3 K8s API: events] Failed to pull image 'ghcr.io/stefanprodan/podinfo:6.15.99' with 'rpc error: code = NotFound'.

Suggested next steps (not executed; a person decides):
- Roll back the frontend deployment to the previous stable revision.: kubectl rollout undo deployment/frontend -n demo
- Verify if image tag 6.15.99 exists in the container registry and correct the deployment configuration if necessary.

Method agent, model gemini-3.1-flash-lite, 14 read-only calls, 102.3 s.
