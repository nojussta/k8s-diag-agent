# Evidence: LabHighLatency demo/frontend-84cbcbb56f-5pg7l

- Read-only calls: **22**, collection time: **0.05 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 2 calls, 0 s)

```json
{
  "alertname": "LabHighLatency",
  "namespace": "demo",
  "pod": "frontend-84cbcbb56f-5pg7l",
  "container": "",
  "deployment": "",
  "service": "frontend",
  "summary": "Service demo/frontend is slow",
  "startsAt": "2026-11-16T13:43:30.907Z"
}
```

## A2 Pod and container status (K8s API: pods; 1 calls, 0 s)

```json
{
  "conditions": {
    "ContainersReady": "True",
    "Initialized": "True",
    "PodReadyToStartContainers": "True",
    "PodScheduled": "True",
    "Ready": "True"
  },
  "containers": [
    {
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "lastState": null,
      "name": "frontend",
      "ready": true,
      "restartCount": 0,
      "state": {
        "startedAt": "2026-11-16T13:41:02Z",
        "state": "running"
      }
    }
  ],
  "node": "kursinis-worker2",
  "phase": "Running",
  "resources": {
    "frontend": {
      "limits": {
        "memory": "128Mi"
      },
      "requests": {
        "cpu": "10m",
        "memory": "32Mi"
      }
    }
  }
}
```

## A3 Events for the pod (K8s API: events; 1 calls, 0 s)

```json
[
  "2026-11-16T13:41:02Z Normal Scheduled x1: Successfully assigned demo/frontend-84cbcbb56f-5pg7l to kursinis-worker2",
  "2026-11-16T13:41:02Z Normal Pulled x1: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-16T13:41:02Z Normal Created x1: Container created",
  "2026-11-16T13:41:02Z Normal Started x1: Container started",
  "2026-11-16T13:41:02Z Warning Unhealthy x1: Readiness probe failed: Get \"http://10.244.1.67:9898/readyz\": dial tcp 10.244.1.67:9898: connect: connection refused"
]
```

## A4 Container logs (current and previous) (K8s API: pods/log; 2 calls, 0.01 s)

```json
{
  "current": [
    "{\"level\":\"info\",\"ts\":\"2026-11-16T13:41:02.862Z\",\"caller\":\"podinfo/main.go:170\",\"msg\":\"Starting podinfo\",\"version\":\"6.15.0\",\"revision\":\"dd507173b7b75b2312a36cabe0de5f09c1ce69c8\",\"port\":\"9898\"}",
    "{\"level\":\"info\",\"ts\":\"2026-11-16T13:41:02.862Z\",\"caller\":\"http/server.go:273\",\"msg\":\"Starting HTTP Server.\",\"addr\":\":9898\"}"
  ],
  "previous": "unavailable: the server rejected our request for an unknown reason (get pods frontend-84cbcbb56f-5pg7l)"
}
```

## A5 Owner, rollout history and recent change (K8s API: replicasets, deployments; 3 calls, 0 s)

```json
{
  "conditions": [
    "Available=True (MinimumReplicasAvailable): Deployment has minimum availability.",
    "Progressing=True (NewReplicaSetAvailable): ReplicaSet \"frontend-84cbcbb56f\" has successfully progressed."
  ],
  "deployment": "frontend",
  "last_rollout_at": "2026-11-16T13:41:03Z",
  "replicas": {
    "availableReplicas": 1,
    "readyReplicas": 1,
    "replicas": 1,
    "unavailableReplicas": 0,
    "updatedReplicas": 1
  },
  "revisions": [
    {
      "change_cause": "release 2.1.1: UI colour",
      "created": "2026-11-16T11:49:43Z",
      "name": "frontend-84cbcbb56f",
      "replicas": 1,
      "revision": 26
    },
    {
      "change_cause": null,
      "created": "2026-11-16T11:48:00Z",
      "name": "frontend-6554997569",
      "replicas": 0,
      "revision": 25
    },
    {
      "change_cause": "release 2.1: frontend configuration update",
      "created": "2026-11-16T11:59:57Z",
      "name": "frontend-69bd4c6cb4",
      "replicas": 0,
      "revision": 18
    }
  ],
  "template_diff_latest_vs_previous": [
    {
      "container": "frontend",
      "field": "args",
      "previous": [
        "--port=9898",
        "--port-metrics=9797",
        "--level=info",
        "--backend-url=http://backend.demo.svc.cluster.local:9898/echo"
      ],
      "current": [
        "--port=9898",
        "--port-metrics=9797",
        "--level=info",
        "--backend-url=http://backend.demo.svc.cluster.local:9898/echo",
        "--ui-color=#2f6f4f"
      ]
    }
  ]
}
```

## A6 Restarts and last termination reason (Prometheus (kube-state-metrics); 2 calls, 0 s)

```json
{
  "last_terminated_reason": [],
  "restarts_15m": [
    {
      "container": "frontend",
      "endpoint": "http",
      "instance": "10.244.2.4:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "pod": "frontend-84cbcbb56f-5pg7l",
      "service": "kps-kube-state-metrics",
      "uid": "f2041a8d-728f-459a-a2c6-aa7b93a3ec71",
      "value": 0
    }
  ]
}
```

## A7 Resource usage vs. limits (Prometheus (cAdvisor + kube-state-metrics); 3 calls, 0.01 s)

```json
{
  "cpu_usage_cores_5m": [
    {
      "container": "frontend",
      "cpu": "total",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-podf2041a8d_728f_459a_a2c6_aa7b93a3ec71.slice/cri-containerd-977b96e6295fdf9fddd7a13ccbb3c9dad6f67712e864a1813cde81f6541f843d.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "977b96e6295fdf9fddd7a13ccbb3c9dad6f67712e864a1813cde81f6541f843d",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "frontend-84cbcbb56f-5pg7l",
      "service": "kps-kubelet",
      "value": 0.0007861959524419513
    }
  ],
  "memory_limit_bytes": [
    {
      "__name__": "kube_pod_container_resource_limits",
      "container": "frontend",
      "endpoint": "http",
      "instance": "10.244.2.4:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "frontend-84cbcbb56f-5pg7l",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "f2041a8d-728f-459a-a2c6-aa7b93a3ec71",
      "unit": "byte",
      "value": 134217728
    }
  ],
  "memory_working_set_max_15m_bytes": [
    {
      "container": "frontend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-podf2041a8d_728f_459a_a2c6_aa7b93a3ec71.slice/cri-containerd-977b96e6295fdf9fddd7a13ccbb3c9dad6f67712e864a1813cde81f6541f843d.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "977b96e6295fdf9fddd7a13ccbb3c9dad6f67712e864a1813cde81f6541f843d",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "frontend-84cbcbb56f-5pg7l",
      "service": "kps-kubelet",
      "value": 27508736
    }
  ]
}
```

## A8 Service endpoints of the workload (K8s API: services, endpointslices; 2 calls, 0 s)

```json
[
  {
    "endpoints": 1,
    "ready": 1,
    "service": "frontend"
  }
]
```

## A9 Node conditions (K8s API: nodes; 1 calls, 0 s)

```json
{
  "conditions": {
    "DiskPressure": "False",
    "MemoryPressure": "False",
    "PIDPressure": "False",
    "Ready": "True"
  },
  "node": "kursinis-worker2"
}
```

## A10 Log and event history (Loki) (Loki; 2 calls, 0.02 s)

```json
{
  "events": [
    "name=frontend-84cbcbb56f-5pg7l kind=Pod objectAPIversion=v1 objectRV=104722 eventRV=104735 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=1 msg=\"Readiness probe failed: Get \\\"http://10.244.1.67:9898/readyz\\\": dial tcp 10.244.1.67:9898: connect: connection refused\" ",
    "name=frontend-84cbcbb56f-5pg7l kind=Pod objectAPIversion=v1 objectRV=104722 eventRV=104734 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Started type=Normal count=1 msg=\"Container started\" ",
    "name=frontend-84cbcbb56f-5pg7l kind=Pod objectAPIversion=v1 objectRV=104722 eventRV=104733 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Created type=Normal count=1 msg=\"Container created\" ",
    "name=frontend-84cbcbb56f-5pg7l kind=Pod objectAPIversion=v1 objectRV=104719 eventRV=104724 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/frontend-84cbcbb56f-5pg7l to kursinis-worker2\" ",
    "name=frontend-84cbcbb56f kind=ReplicaSet objectAPIversion=apps/v1 objectRV=104718 eventRV=104723 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: frontend-84cbcbb56f-5pg7l\" ",
    "name=frontend-84cbcbb56f-5pg7l kind=Pod objectAPIversion=v1 objectRV=104722 eventRV=104731 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" "
  ],
  "pod_logs": [
    "{\"level\":\"info\",\"ts\":\"2026-11-16T13:41:02.862Z\",\"caller\":\"podinfo/main.go:170\",\"msg\":\"Starting podinfo\",\"version\":\"6.15.0\",\"revision\":\"dd507173b7b75b2312a36cabe0de5f09c1ce69c8\",\"port\":\"9898\"}\n",
    "{\"level\":\"info\",\"ts\":\"2026-11-16T13:41:02.862Z\",\"caller\":\"http/server.go:273\",\"msg\":\"Starting HTTP Server.\",\"addr\":\":9898\"}\n"
  ]
}
```

## A11 Service golden signals (RED) (Prometheus (application metrics); 3 calls, 0.01 s)

```json
{
  "error_rate_5xx_rps": [],
  "latency_p95_s": [
    {
      "value": 0.9218042920472371
    }
  ],
  "request_rate_rps": [
    {
      "value": 0.8028623862130521
    }
  ]
}
```

