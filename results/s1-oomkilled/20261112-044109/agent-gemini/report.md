Incident triage: LabPodCrashLooping
Probable cause: OOMKilled in demo/deployment/backend (container backend) (confidence high)
A recent deployment revision (revision 12) introduced a new argument '--stress-memory=256' which increased the container's memory consumption beyond the configured limit of 128Mi.

Evidence:
- [t1 K8s API: pods] The backend container is in 'CrashLoopBackOff' state with 'OOMKilled' as the last termination reason and an exit code of 137.
- [t3 K8s API: replicasets, deployments] The latest deployment revision 12 added the flag '--stress-memory=256' to the backend container arguments.
- [t1 K8s API: pods] The container has a memory limit of 128Mi.
- [t2 Prometheus] The working set memory usage is near the 128Mi limit (128,729,088 bytes).

Suggested next steps (not executed; a person decides):
- Increase the memory limit for the backend container in the deployment spec to accommodate the increased memory usage.: kubectl edit deployment backend -n demo
- Alternatively, revert to the previous deployment revision if the memory-stress argument is not required.: kubectl rollout undo deployment/backend -n demo

Method agent, model gemini-3.1-flash-lite, 9 read-only calls, 114.3 s.
