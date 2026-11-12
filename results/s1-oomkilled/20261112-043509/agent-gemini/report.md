Incident triage: LabPodCrashLooping
Probable cause: OOMKilled in demo/deployment/backend (container backend) (confidence high)
The latest deployment revision increased memory usage with the '--stress-memory=256' flag, which exceeds the defined memory limit of 128Mi.

Evidence:
- [t1 K8s API: pods] Container backend status is OOMKilled with exitCode 137.
- [t2 Prometheus] The container's memory limit is 128Mi (134217728 bytes) and the observed memory working set is hovering around 126Mi.
- [t3 K8s API: replicasets, deployments] The latest revision added '--stress-memory=256' to the arguments of the backend container.

Suggested next steps (not executed; a person decides):
- Increase the memory limit for the backend container in the deployment spec to accommodate the increased memory usage, or remove the stress argument.: kubectl edit deployment backend -n demo

Method agent, model gemini-3.1-flash-lite, 10 read-only calls, 93.1 s.
