Incident triage: LabPodCrashLooping
Probable cause: OOMKilled in demo/deployment/backend (container backend) (confidence high)
A configuration change in revision 10 added the argument '--stress-memory=256' to the container, causing it to consume more memory than the allocated 128Mi limit, resulting in OOMKilled events.

Evidence:
- [A2 K8s API: pods] The pod 'backend-7b8d9f4db-gmppc' container 'backend' has a lastState of 'OOMKilled' with exitCode 137.
- [A5 K8s API: deployments] Revision 10 of the deployment added '--stress-memory=256' to the container arguments.
- [A2 K8s API: pods] The container 'backend' is configured with a memory limit of 128Mi.

Suggested next steps (not executed; a person decides):
- Revert the deployment to the previous revision (revision 9) to remove the stress-memory argument.: kubectl rollout undo deployment/backend -n demo
- Alternatively, increase the memory limit for the backend container if the memory increase is intended.: kubectl edit deployment/backend -n demo

Method pipeline, model gemini-3.1-flash-lite, 20 read-only calls, 2.4 s.
