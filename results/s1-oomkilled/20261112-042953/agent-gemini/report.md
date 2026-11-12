Incident triage: LabPodCrashLooping
Probable cause: OOMKilled in demo/deployment/backend (container backend) (confidence high)
The latest revision (rev 8) added the '--stress-memory=256' argument to the container, which is causing it to exceed its memory limit of 128Mi.

Evidence:
- [t1 K8s API: pods] Pod 'backend-7b8d9f4db-4grxg' has a container 'backend' in 'OOMKilled' state with exit code 137.
- [t3 K8s API: replicasets, deployments] The latest revision (rev 8) added the argument '--stress-memory=256' to the 'backend' container compared to the previous revision.
- [t2 Prometheus] The 'backend' container has a memory limit of 128Mi (134217728 bytes).

Suggested next steps (not executed; a person decides):
- Roll back the deployment to the previous revision (7) to restore a stable version.: kubectl rollout undo deployment/backend -n demo
- Alternatively, increase the memory limit for the backend deployment to accommodate the stress test argument.: kubectl edit deployment/backend -n demo

Method agent, model gemini-3.1-flash-lite, 15 read-only calls, 89.5 s.
