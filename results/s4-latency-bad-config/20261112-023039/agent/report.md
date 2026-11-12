Incident triage: LabHighLatency
Probable cause: Unknown in frontend-69bd4c6cb4-bflbc (confidence medium)
The 95th percentile latency of service demo/frontend has been above 500 ms for 1 minute. The pod frontend-69bd4c6cb4-bflbc is Running, but the latency is consistently high. There are no obvious errors in the logs or metrics, and the node conditions are normal. The issue might be related to the workload itself, such as high CPU or memory usage, or it could be an issue with the deployment or the environment. Further investigation is needed to determine the root cause.

Evidence:
- [t1 alert] Service demo/frontend is slow
- [t2 list_pods] Pod frontend-69bd4c6cb4-bflbc is Running
- [t3 query_logs] Pod frontend-69bd4c6cb4-bflbc has 95th percentile latency > 500 ms
- [t6 query_metrics] Pod frontend-69bd4c6cb4-bflbc has 95th percentile latency > 500 ms
- [t5 get_resource_usage] Pod frontend-69bd4c6cb4-bflbc has 95th percentile latency > 500 ms
- [t7 get_node_status] Pod frontend-69bd4c6cb4-bflbc has 95th percentile latency > 500 ms
- [t7 get_node_status] Pod frontend-69bd4c6cb4-bflbc has 95th percentile latency > 500 ms

Suggested next steps (not executed; a person decides):
- Check resource usage: kubectl get resource usage frontend-69bd4c6cb4-bflbc

Unknown: Possible high CPU or memory usage; Deployment or environment issue

Method agent, model apple-foundationmodel, 7 read-only calls, 26.1 s.
