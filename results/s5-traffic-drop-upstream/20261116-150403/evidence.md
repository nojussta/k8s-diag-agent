# Evidence: LabTrafficDrop demo/backend-765cf7c4db-4wmvc

- Read-only calls: **22**, collection time: **0.08 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 2 calls, 0 s)

```json
{
  "alertname": "LabTrafficDrop",
  "namespace": "demo",
  "pod": "backend-765cf7c4db-4wmvc",
  "container": "",
  "deployment": "",
  "service": "backend",
  "summary": "Service demo/backend receives almost no requests",
  "startsAt": "2026-11-16T13:07:45.907Z"
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
        "startedAt": "2026-11-16T13:04:03Z",
        "state": "running"
      }
    }
  ],
  "node": "kursinis-worker2",
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
  "2026-11-16T13:04:03Z Normal Scheduled x1: Successfully assigned demo/backend-765cf7c4db-4wmvc to kursinis-worker2",
  "2026-11-16T13:04:03Z Normal Pulled x1: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-16T13:04:03Z Normal Created x1: Container created",
  "2026-11-16T13:04:03Z Normal Started x1: Container started",
  "2026-11-16T13:04:04Z Warning Unhealthy x1: Readiness probe failed: Get \"http://10.244.1.51:9898/readyz\": dial tcp 10.244.1.51:9898: connect: connection refused"
]
```

## A4 Container logs (current and previous) (K8s API: pods/log; 2 calls, 0.01 s)

```json
{
  "current": [
    "{\"level\":\"info\",\"ts\":\"2026-11-16T13:04:04.008Z\",\"caller\":\"podinfo/main.go:170\",\"msg\":\"Starting podinfo\",\"version\":\"6.15.0\",\"revision\":\"dd507173b7b75b2312a36cabe0de5f09c1ce69c8\",\"port\":\"9898\"}",
    "{\"level\":\"info\",\"ts\":\"2026-11-16T13:04:04.009Z\",\"caller\":\"http/server.go:273\",\"msg\":\"Starting HTTP Server.\",\"addr\":\":9898\"}"
  ],
  "previous": "unavailable: the server rejected our request for an unknown reason (get pods backend-765cf7c4db-4wmvc)"
}
```

## A5 Owner, rollout history and recent change (K8s API: replicasets, deployments; 3 calls, 0.01 s)

```json
{
  "conditions": [
    "Available=True (MinimumReplicasAvailable): Deployment has minimum availability.",
    "Progressing=True (NewReplicaSetAvailable): ReplicaSet \"backend-765cf7c4db\" has successfully progressed."
  ],
  "deployment": "backend",
  "last_rollout_at": "2026-11-16T13:04:05Z",
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
      "revision": 20
    },
    {
      "change_cause": null,
      "created": "2026-11-16T11:48:00Z",
      "name": "backend-5b97dd7dcb",
      "replicas": 0,
      "revision": 19
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

## A6 Restarts and last termination reason (Prometheus (kube-state-metrics); 2 calls, 0.01 s)

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
      "pod": "backend-765cf7c4db-4wmvc",
      "service": "kps-kube-state-metrics",
      "uid": "46acc66c-1d4f-4606-9609-5243dc0587f8",
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
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod46acc66c_1d4f_4606_9609_5243dc0587f8.slice/cri-containerd-3f80a050a63f54cb1f4240d9bdaa9a9844fc7d6f0e9bb0cea2692f38d71f1c99.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "3f80a050a63f54cb1f4240d9bdaa9a9844fc7d6f0e9bb0cea2692f38d71f1c99",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "backend-765cf7c4db-4wmvc",
      "service": "kps-kubelet",
      "value": 0.0004182513079184638
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
      "node": "kursinis-worker2",
      "pod": "backend-765cf7c4db-4wmvc",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "46acc66c-1d4f-4606-9609-5243dc0587f8",
      "unit": "byte",
      "value": 134217728
    }
  ],
  "memory_working_set_max_15m_bytes": [
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod46acc66c_1d4f_4606_9609_5243dc0587f8.slice/cri-containerd-3f80a050a63f54cb1f4240d9bdaa9a9844fc7d6f0e9bb0cea2692f38d71f1c99.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "3f80a050a63f54cb1f4240d9bdaa9a9844fc7d6f0e9bb0cea2692f38d71f1c99",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "backend-765cf7c4db-4wmvc",
      "service": "kps-kubelet",
      "value": 30486528
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
  "node": "kursinis-worker2"
}
```

## A10 Log and event history (Loki) (Loki; 2 calls, 0.02 s)

```json
{
  "events": [
    "name=backend-765cf7c4db-4wmvc kind=Pod objectAPIversion=v1 objectRV=100088 eventRV=100098 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Started type=Normal count=1 msg=\"Container started\" ",
    "name=backend-765cf7c4db-4wmvc kind=Pod objectAPIversion=v1 objectRV=100088 eventRV=100097 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Created type=Normal count=1 msg=\"Container created\" ",
    "name=backend-765cf7c4db-4wmvc kind=Pod objectAPIversion=v1 objectRV=100084 eventRV=100090 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/backend-765cf7c4db-4wmvc to kursinis-worker2\" ",
    "name=backend-765cf7c4db kind=ReplicaSet objectAPIversion=apps/v1 objectRV=100082 eventRV=100086 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: backend-765cf7c4db-4wmvc\" ",
    "name=backend-765cf7c4db-4wmvc kind=Pod objectAPIversion=v1 objectRV=100088 eventRV=100095 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-765cf7c4db-4wmvc kind=Pod objectAPIversion=v1 objectRV=100088 eventRV=100099 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=1 msg=\"Readiness probe failed: Get \\\"http://10.244.1.51:9898/readyz\\\": dial tcp 10.244.1.51:9898: connect: connection refused\" "
  ],
  "pod_logs": [
    "{\"level\":\"info\",\"ts\":\"2026-11-16T13:04:04.008Z\",\"caller\":\"podinfo/main.go:170\",\"msg\":\"Starting podinfo\",\"version\":\"6.15.0\",\"revision\":\"dd507173b7b75b2312a36cabe0de5f09c1ce69c8\",\"port\":\"9898\"}\n",
    "{\"level\":\"info\",\"ts\":\"2026-11-16T13:04:04.009Z\",\"caller\":\"http/server.go:273\",\"msg\":\"Starting HTTP Server.\",\"addr\":\":9898\"}\n"
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
      "value": 0.40344677040758825
    }
  ]
}
```

