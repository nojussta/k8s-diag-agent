Incident triage: LabHighLatency
Probable cause: Scheduling in demo/frontend (confidence high)
Service demo/frontend-69bd4c6cb4-xjfk6 is Pending with a FailedScheduling event, indicating that the pod is not being scheduled properly.

Evidence:
- [A1 alert] Service demo/frontend is slow
- [A10 A10] Service demo/frontend-69bd4c6cb4-xjfk6 is Pending with a FailedScheduling event
- [A11 A11] Service demo/frontend-69bd4c6cb4-xjfk6 has a 95th percentile latency of 1.96 seconds

Suggested next steps (not executed; a person decides):
- Review the deployment configuration and ensure that the workload is properly scaled and that there are no misconfigurations.

Method pipeline, model apple-foundationmodel, 22 read-only calls, 27.6 s.
