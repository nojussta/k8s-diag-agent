Incident triage: LabPodNotReady
Probable cause: ImagePullError in demo/deployment/frontend (confidence high)
The deployment was updated to use a non-existent container image tag '6.15.99', whereas the previous tag was '6.15.0'.

Evidence:
- [A2 K8s API] Pod 'frontend-68747558f7-zksdv' is in 'ImagePullBackOff' state because image 'ghcr.io/stefanprodan/podinfo:6.15.99' is not found.
- [A5 K8s API] The container image was changed from 'ghcr.io/stefanprodan/podinfo:6.15.0' to 'ghcr.io/stefanprodan/podinfo:6.15.99' in the latest deployment revision.

Suggested next steps (not executed; a person decides):
- Roll back the deployment to the previous stable revision (45).: kubectl rollout undo deployment/frontend -n demo --to-revision=45
- Verify if the image tag '6.15.99' was intended and exists in the registry, or correct the tag in the deployment configuration.: kubectl edit deployment/frontend -n demo

Method pipeline, model gemini-3.1-flash-lite, 19 read-only calls, 2.0 s.
