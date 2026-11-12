Incident triage: LabHighLatency
Probable cause: Unknown in service.demo.frontend (confidence medium)
Slow service is not due to OOM, crash, image pull error, configuration error, probe failure, scheduling, or node problem. Further investigation is needed to determine the cause.

Evidence:
- [t1 query_logs] Service demo/frontend is slow
- [t3 get_pod_status] Pod frontend-69bd4c6cb4-xjfk6 is Running
- [t6 query_metrics] Pod frontend-69bd4c6cb4-xjfk6 has 95th percentile latency > 500 ms
- [t8 get_node_status] Node kursinis-worker2 is Running

Suggested next steps (not executed; a person decides):
- Check node resources: kubectl get nodes

Unknown: Possible resource constraints on node kursinis-worker2; Possible issues with the frontend container or its dependencies

Method agent, model apple-foundationmodel, 7 read-only calls, 23.3 s.
