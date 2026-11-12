Incident triage: LabHighLatency
Probable cause: NodeProblem in node kursinis-worker2 (confidence medium)
Node is overloaded with CPU and memory usage

Evidence:
- [t1 query_logs] Service demo/frontend is slow
- [t3 get_pod_status] Pod frontend-69bd4c6cb4-nn7wk is Running
- [t5 get_resource_usage] Pod frontend-69bd4c6cb4-nn7wk has 100% CPU usage
- [t5 get_resource_usage] Pod frontend-69bd4c6cb4-nn7wk has 100% memory usage
- [t6 get_node_status] Pod frontend-69bd4c6cb4-nn7wk has 100% CPU usage on node kursinis-worker2
- [t6 get_node_status] Pod frontend-69bd4c6cb4-nn7wk has 100% memory usage on node kursinis-worker2
- [t6 get_node_status] Pod frontend-69bd4c6cb4-nn7wk has 100% CPU usage on node kursinis-worker2

Suggested next steps (not executed; a person decides):
- Check CPU and memory limits: kubectl get pods --namespace demo --selector app=demo.service.demo.frontend --show-labels --show-resource-usage

Method agent, model apple-foundationmodel, 7 read-only calls, 23.3 s.
