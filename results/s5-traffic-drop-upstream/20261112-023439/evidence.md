# Evidence: LabTrafficDrop demo/backend-5b97dd7dcb-f44wr

- Read-only calls: **22**, collection time: **0.07 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 2 calls, 0.01 s)

```json
{
  "alertname": "LabTrafficDrop",
  "namespace": "demo",
  "pod": "backend-5b97dd7dcb-f44wr",
  "container": "",
  "deployment": "",
  "service": "backend",
  "summary": "Service demo/backend receives almost no requests",
  "startsAt": "2026-11-12T00:37:30.907Z"
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
        "startedAt": "2026-11-11T23:56:29Z",
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

## A3 Events for the pod (K8s API: events; 1 calls, 0 s)

```json
[
  "2026-11-11T23:56:25Z Normal Scheduled x1: Successfully assigned demo/backend-5b97dd7dcb-f44wr to kursinis-worker2",
  "2026-11-11T23:56:25Z Normal Pulling x1: Pulling image \"ghcr.io/stefanprodan/podinfo:6.15.0\"",
  "2026-11-11T23:56:29Z Normal Pulled x1: Successfully pulled image \"ghcr.io/stefanprodan/podinfo:6.15.0\" in 3.736s (3.736s including waiting). Image size: 33001367 bytes.",
  "2026-11-11T23:56:29Z Normal Created x1: Container created",
  "2026-11-11T23:56:29Z Normal Started x1: Container started",
  "2026-11-11T23:56:29Z Warning Unhealthy x1: Readiness probe failed: Get \"http://10.244.1.10:9898/readyz\": dial tcp 10.244.1.10:9898: connect: connection refused"
]
```

## A4 Container logs (current and previous) (K8s API: pods/log; 2 calls, 0.01 s)

```json
{
  "current": [
    "{\"level\":\"info\",\"ts\":\"2026-11-11T23:56:29.589Z\",\"caller\":\"podinfo/main.go:170\",\"msg\":\"Starting podinfo\",\"version\":\"6.15.0\",\"revision\":\"dd507173b7b75b2312a36cabe0de5f09c1ce69c8\",\"port\":\"9898\"}",
    "{\"level\":\"info\",\"ts\":\"2026-11-11T23:56:29.590Z\",\"caller\":\"http/server.go:273\",\"msg\":\"Starting HTTP Server.\",\"addr\":\":9898\"}"
  ],
  "previous": "unavailable: the server rejected our request for an unknown reason (get pods backend-5b97dd7dcb-f44wr)"
}
```

## A5 Owner, rollout history and recent change (K8s API: replicasets, deployments; 3 calls, 0.01 s)

```json
{
  "conditions": [
    "Available=True (MinimumReplicasAvailable): Deployment has minimum availability.",
    "Progressing=True (NewReplicaSetAvailable): ReplicaSet \"backend-5b97dd7dcb\" has successfully progressed."
  ],
  "deployment": "backend",
  "replicas": {
    "availableReplicas": 2,
    "readyReplicas": 2,
    "replicas": 2,
    "unavailableReplicas": 0,
    "updatedReplicas": 2
  },
  "revisions": [
    {
      "change_cause": null,
      "created": "2026-11-11T23:56:25Z",
      "name": "backend-5b97dd7dcb",
      "replicas": 2,
      "revision": 1
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
      "pod": "backend-5b97dd7dcb-f44wr",
      "service": "kps-kube-state-metrics",
      "uid": "bd23f116-a5d8-40df-bca2-3d5e854e0941",
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
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-podbd23f116_a5d8_40df_bca2_3d5e854e0941.slice/cri-containerd-1851a34190e818a407f37475e143456f28f7830238b35a2da95c38ed16c65144.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "1851a34190e818a407f37475e143456f28f7830238b35a2da95c38ed16c65144",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "backend-5b97dd7dcb-f44wr",
      "service": "kps-kubelet",
      "value": 0.0005927290414049325
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
      "pod": "backend-5b97dd7dcb-f44wr",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "bd23f116-a5d8-40df-bca2-3d5e854e0941",
      "unit": "byte",
      "value": 134217728
    }
  ],
  "memory_working_set_max_15m_bytes": [
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-podbd23f116_a5d8_40df_bca2_3d5e854e0941.slice/cri-containerd-1851a34190e818a407f37475e143456f28f7830238b35a2da95c38ed16c65144.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "1851a34190e818a407f37475e143456f28f7830238b35a2da95c38ed16c65144",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "backend-5b97dd7dcb-f44wr",
      "service": "kps-kubelet",
      "value": 38592512
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
  "events": [],
  "pod_logs": []
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
      "value": 0.3509633945179518
    }
  ]
}
```

