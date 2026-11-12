Incident triage: LabPodNotReady
Probable cause: ImagePullError in demo/deployment/frontend (container frontend) (confidence high)
The latest rollout (revision 46) updated the image tag from '6.15.0' to '6.15.99', which is an invalid or non-existent image reference.

Evidence:
- [t1 K8s API: pods] Pod 'frontend-68747558f7-zksdv' is in 'Pending' phase with container state 'ImagePullBackOff' and error 'ghcr.io/stefanprodan/podinfo:6.15.99: not found'.
- [t2 K8s API: events] Failed to pull image due to error 'rpc error: code = NotFound'.
- [t3 K8s API: replicasets, deployments] The latest revision (46) changed the image from 'ghcr.io/stefanprodan/podinfo:6.15.0' to 'ghcr.io/stefanprodan/podinfo:6.15.99'.

Suggested next steps (not executed; a person decides):
- Roll back the frontend deployment to the previous stable revision (45).: kubectl rollout undo deployment/frontend -n demo --to-revision=45

Method agent, model gemini-3.1-flash-lite, 5 read-only calls, 31.5 s.
