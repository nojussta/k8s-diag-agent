Incident triage: LabPodNotReady
Probable cause: ImagePullError in demo/deployment/frontend (confidence high)
A recent configuration change updated the image tag from '6.15.0' to '6.15.99', which does not exist on the registry.

Evidence:
- [A2 pods] The pod is in ImagePullBackOff state with the error: 'failed to resolve reference "ghcr.io/stefanprodan/podinfo:6.15.99": ghcr.io/stefanprodan/podinfo:6.15.99: not found'.
- [A5 deployments] The template diff shows the image was updated from 'ghcr.io/stefanprodan/podinfo:6.15.0' to 'ghcr.io/stefanprodan/podinfo:6.15.99'.
- [A10 Loki] Event logs confirm repeated failure to pull the image 'ghcr.io/stefanprodan/podinfo:6.15.99' due to 'not found'.

Suggested next steps (not executed; a person decides):
- Roll back the deployment to the previous known working image tag.: kubectl rollout undo deployment/frontend -n demo
- Verify if image '6.15.99' was intended to be pushed or if this was a configuration typo.

Method pipeline, model gemini-3.1-flash-lite, 19 read-only calls, 2.5 s.
