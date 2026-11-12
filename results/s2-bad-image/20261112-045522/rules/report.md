Incident triage: LabPodNotReady
Probable cause: ImagePullError in demo/deployment/frontend (confidence high)
image ghcr.io/stefanprodan/podinfo:6.15.99 of container frontend cannot be pulled (ImagePullBackOff). It started with the newest revision: revision 50 changed frontend image: "ghcr.io/stefanprodan/podinfo:6.15.0" -> "ghcr.io/stefanprodan/podinfo:6.15.99".

Evidence:
- [A2 pod status] phase=Pending; frontend restarts=0 state=ImagePullBackOff
- [A3 events] 2026-11-12T02:57:39Z Warning Failed x8: Error: ImagePullBackOff
- [A5 rollout history] revision 50 changed frontend image: "ghcr.io/stefanprodan/podinfo:6.15.0" -> "ghcr.io/stefanprodan/podinfo:6.15.99".

Suggested next steps (not executed; a person decides):
- Roll back to the previous revision: kubectl -n demo rollout undo deployment/frontend

Method rules, model none, 19 read-only calls, 0.1 s.
