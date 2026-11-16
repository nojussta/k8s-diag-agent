# Evidence: LabTrafficDrop demo/backend-765cf7c4db-9t7wx

- Read-only calls: **22**, collection time: **0.06 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 2 calls, 0.01 s)

```json
{
  "alertname": "LabTrafficDrop",
  "namespace": "demo",
  "pod": "backend-765cf7c4db-9t7wx",
  "container": "",
  "deployment": "",
  "service": "backend",
  "summary": "Service demo/backend receives almost no requests",
  "startsAt": "2026-11-16T12:11:15.907Z"
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
      "name": "backend",
      "ready": true,
      "restartCount": 0,
      "state": {
        "startedAt": "2026-11-16T12:07:38Z",
        "state": "running"
      }
    }
  ],
  "node": "kursinis-worker",
  "phase": "Running",
  "resources": {
    "backend": {
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

## A3 Events for the pod (K8s API: events; 1 calls, 0.01 s)

```json
[
  "2026-11-16T12:07:38Z Normal Scheduled x1: Successfully assigned demo/backend-765cf7c4db-9t7wx to kursinis-worker",
  "2026-11-16T12:07:38Z Normal Pulled x1: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-16T12:07:38Z Normal Created x1: Container created",
  "2026-11-16T12:07:38Z Normal Started x1: Container started",
  "2026-11-16T12:07:38Z Warning Unhealthy x1: Readiness probe failed: Get \"http://10.244.2.14:9898/readyz\": dial tcp 10.244.2.14:9898: connect: connection refused"
]
```

## A4 Container logs (current and previous) (K8s API: pods/log; 2 calls, 0.01 s)

```json
{
  "current": [
    "{\"level\":\"info\",\"ts\":\"2026-11-16T12:07:39.001Z\",\"caller\":\"podinfo/main.go:170\",\"msg\":\"Starting podinfo\",\"version\":\"6.15.0\",\"revision\":\"dd507173b7b75b2312a36cabe0de5f09c1ce69c8\",\"port\":\"9898\"}",
    "{\"level\":\"info\",\"ts\":\"2026-11-16T12:07:39.002Z\",\"caller\":\"http/server.go:273\",\"msg\":\"Starting HTTP Server.\",\"addr\":\":9898\"}"
  ],
  "previous": "unavailable: the server rejected our request for an unknown reason (get pods backend-765cf7c4db-9t7wx)"
}
```

## A5 Owner, rollout history and recent change (K8s API: replicasets, deployments; 3 calls, 0 s)

```json
{
  "conditions": [
    "Available=True (MinimumReplicasAvailable): Deployment has minimum availability.",
    "Progressing=True (NewReplicaSetAvailable): ReplicaSet \"backend-765cf7c4db\" has successfully progressed."
  ],
  "deployment": "backend",
  "last_rollout_at": "2026-11-16T12:07:39Z",
  "replicas": {
    "availableReplicas": 2,
    "readyReplicas": 2,
    "replicas": 2,
    "unavailableReplicas": 0,
    "updatedReplicas": 2
  },
  "revisions": [
    {
      "change_cause": "release 2.0.1: UI colour",
      "created": "2026-11-16T11:58:37Z",
      "name": "backend-765cf7c4db",
      "replicas": 2,
      "revision": 8
    },
    {
      "change_cause": null,
      "created": "2026-11-16T11:48:00Z",
      "name": "backend-5b97dd7dcb",
      "replicas": 0,
      "revision": 7
    },
    {
      "change_cause": "release 2.2: request smoothing",
      "created": "2026-11-16T11:50:44Z",
      "name": "backend-6f4858f6b5",
      "replicas": 0,
      "revision": 2
    }
  ],
  "template_diff_latest_vs_previous": [
    {
      "container": "backend",
      "field": "args",
      "previous": [
        "--port=9898",
        "--port-metrics=9797",
        "--level=info"
      ],
      "current": [
        "--port=9898",
        "--port-metrics=9797",
        "--level=info",
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
      "container": "backend",
      "endpoint": "http",
      "instance": "10.244.2.4:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "pod": "backend-765cf7c4db-9t7wx",
      "service": "kps-kube-state-metrics",
      "uid": "4e70c4a9-18b3-4f6e-903d-256a9ef57399",
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
      "container": "backend",
      "cpu": "total",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod4e70c4a9_18b3_4f6e_903d_256a9ef57399.slice/cri-containerd-0933124938bb9dd75a1e1c0f9cf4abc0c47d89de619057beb5a278730073f653.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.4:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "0933124938bb9dd75a1e1c0f9cf4abc0c47d89de619057beb5a278730073f653",
      "namespace": "demo",
      "node": "kursinis-worker",
      "pod": "backend-765cf7c4db-9t7wx",
      "service": "kps-kubelet",
      "value": 0.0004968773162888139
    }
  ],
  "memory_limit_bytes": [
    {
      "__name__": "kube_pod_container_resource_limits",
      "container": "backend",
      "endpoint": "http",
      "instance": "10.244.2.4:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "node": "kursinis-worker",
      "pod": "backend-765cf7c4db-9t7wx",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "4e70c4a9-18b3-4f6e-903d-256a9ef57399",
      "unit": "byte",
      "value": 134217728
    }
  ],
  "memory_working_set_max_15m_bytes": [
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod4e70c4a9_18b3_4f6e_903d_256a9ef57399.slice/cri-containerd-0933124938bb9dd75a1e1c0f9cf4abc0c47d89de619057beb5a278730073f653.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.4:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "0933124938bb9dd75a1e1c0f9cf4abc0c47d89de619057beb5a278730073f653",
      "namespace": "demo",
      "node": "kursinis-worker",
      "pod": "backend-765cf7c4db-9t7wx",
      "service": "kps-kubelet",
      "value": 28635136
    }
  ]
}
```

## A8 Service endpoints of the workload (K8s API: services, endpointslices; 2 calls, 0 s)

```json
[
  {
    "endpoints": 2,
    "ready": 2,
    "service": "backend"
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
  "node": "kursinis-worker"
}
```

## A10 Log and event history (Loki) (Loki; 2 calls, 0.01 s)

```json
{
  "events": [
    "name=backend-765cf7c4db-9t7wx kind=Pod objectAPIversion=v1 objectRV=93059 eventRV=93070 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Unhealthy type=Warning count=1 msg=\"Readiness probe failed: Get \\\"http://10.244.2.14:9898/readyz\\\": dial tcp 10.244.2.14:9898: connect: connection refused\" ",
    "name=backend-765cf7c4db-9t7wx kind=Pod objectAPIversion=v1 objectRV=93059 eventRV=93069 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Started type=Normal count=1 msg=\"Container started\" ",
    "name=backend-765cf7c4db-9t7wx kind=Pod objectAPIversion=v1 objectRV=93059 eventRV=93068 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Created type=Normal count=1 msg=\"Container created\" ",
    "name=backend-765cf7c4db-9t7wx kind=Pod objectAPIversion=v1 objectRV=93056 eventRV=93062 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/backend-765cf7c4db-9t7wx to kursinis-worker\" ",
    "name=backend-765cf7c4db kind=ReplicaSet objectAPIversion=apps/v1 objectRV=93053 eventRV=93058 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: backend-765cf7c4db-9t7wx\" ",
    "name=backend-765cf7c4db-9t7wx kind=Pod objectAPIversion=v1 objectRV=93059 eventRV=93066 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" "
  ],
  "pod_logs": [
    "{\"level\":\"info\",\"ts\":\"2026-11-16T12:07:39.001Z\",\"caller\":\"podinfo/main.go:170\",\"msg\":\"Starting podinfo\",\"version\":\"6.15.0\",\"revision\":\"dd507173b7b75b2312a36cabe0de5f09c1ce69c8\",\"port\":\"9898\"}\n",
    "{\"level\":\"info\",\"ts\":\"2026-11-16T12:07:39.002Z\",\"caller\":\"http/server.go:273\",\"msg\":\"Starting HTTP Server.\",\"addr\":\":9898\"}\n"
  ]
}
```

## A11 Service golden signals (RED) (Prometheus (application metrics); 3 calls, 0.01 s)

```json
{
  "error_rate_5xx_rps": [],
  "latency_p95_s": [
    {
      "value": 0.00475
    }
  ],
  "request_rate_rps": [
    {
      "value": 0.3810746666666667
    }
  ]
}
```

