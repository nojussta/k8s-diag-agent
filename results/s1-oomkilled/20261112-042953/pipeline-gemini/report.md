Incident triage: LabPodCrashLooping
Probable cause: OOMKilled in demo/deployment/backend (container backend) (confidence high)
A recent deployment revision (revision 8) introduced the '--stress-memory=256' argument to the container, which is causing the application to exceed the configured memory limit of 128Mi.

Evidence:
- [A2 K8s API: pods] Container 'backend' has a last termination reason of 'OOMKilled' and a restart count of 3.
- [A2 K8s API: pods] The memory limit for the 'backend' container is 128Mi.
- [A5 K8s API: deployments] The latest deployment revision 8 added '--stress-memory=256' to the container arguments.
- [A6 Prometheus (kube-state-metrics)] The last terminated reason is confirmed as 'OOMKilled'.

Suggested next steps (not executed; a person decides):
- Revert the deployment to the previous revision (7) or remove the '--stress-memory=256' argument from the container configuration.: kubectl edit deployment backend -n demo

Method pipeline, model gemini-3.1-flash-lite, 20 read-only calls, 2.6 s.
