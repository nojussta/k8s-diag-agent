Incident triage: LabPodNotReady
Probable cause: ImagePullError in demo/deployment/frontend (container frontend) (confidence high)
The latest deployment revision (50) updated the container image to 'ghcr.io/stefanprodan/podinfo:6.15.99', which fails to pull. Previous revision (49) uses 'ghcr.io/stefanprodan/podinfo:6.15.0' and is running successfully.

Evidence:
- [t1 K8s API: pods] Pod 'frontend-68747558f7-fmjrs' is in 'Pending' phase due to 'ImagePullBackOff' with message: 'failed to pull and unpack image "ghcr.io/stefanprodan/podinfo:6.15.99": failed to resolve reference "ghcr.io/stefanprodan/podinfo:6.15.99": ghcr.io/stefanprodan/podinfo:6.15.99: not found'
- [t2 K8s API: replicasets, deployments] Revision 50 changed the image from 'ghcr.io/stefanprodan/podinfo:6.15.0' to 'ghcr.io/stefanprodan/podinfo:6.15.99'.
- [t3 K8s API: events] Events show consistent failure to resolve the image reference 'ghcr.io/stefanprodan/podinfo:6.15.99'.
- [t8 K8s API: pods] An older replica 'frontend-6554997569-grccp' running version '6.15.0' is successfully running.

Suggested next steps (not executed; a person decides):
- Roll back the 'frontend' deployment to the previous stable revision (49) to restore full capacity.: kubectl rollout undo deployment/frontend -n demo
- Verify if 'ghcr.io/stefanprodan/podinfo:6.15.99' was a typo or if the image needs to be pushed to the registry.

Method agent, model gemini-3.1-flash-lite, 8 read-only calls, 44.6 s.
