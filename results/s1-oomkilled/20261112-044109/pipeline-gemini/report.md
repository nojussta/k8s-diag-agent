Incident triage: LabPodCrashLooping
Probable cause: OOMKilled in demo/deployment/backend (confidence high)
A recent deployment update (revision 12) added the argument '--stress-memory=256' to the 'backend' container, which causes it to exceed the allocated memory limit of 128Mi.

Evidence:
- [A2 K8s API: pods] The container 'backend' terminated with exitCode 137 and reason 'OOMKilled'.
- [A5 K8s API: deployments] The latest revision (12) introduced the flag '--stress-memory=256' to the container args.
- [A2 K8s API: pods] The container 'backend' has a memory limit of 128Mi.

Suggested next steps (not executed; a person decides):
- Revert the deployment 'backend' to the previous revision to remove the memory-stressing flag.: kubectl rollout undo deployment/backend -n demo
- Alternatively, increase the memory limit for the 'backend' container in the deployment spec if the flag is required.: kubectl edit deployment/backend -n demo

Method pipeline, model gemini-3.1-flash-lite, 20 read-only calls, 2.7 s.
