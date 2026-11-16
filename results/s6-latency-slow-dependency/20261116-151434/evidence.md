# Evidence: LabHighLatency demo/frontend-84cbcbb56f-f5jhx

- Read-only calls: **22**, collection time: **0.07 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 2 calls, 0 s)

```json
{
  "alertname": "LabHighLatency",
  "namespace": "demo",
  "pod": "frontend-84cbcbb56f-f5jhx",
  "container": "",
  "deployment": "",
  "service": "frontend",
  "summary": "Service demo/frontend is slow",
  "startsAt": "2026-11-16T13:17:15.907Z"
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
        "startedAt": "2026-11-16T13:14:34Z",
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
  "2026-11-16T13:14:34Z Normal Scheduled x1: Successfully assigned demo/frontend-84cbcbb56f-f5jhx to kursinis-worker2",
  "2026-11-16T13:14:34Z Normal Pulled x1: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-16T13:14:34Z Normal Created x1: Container created",
  "2026-11-16T13:14:34Z Normal Started x1: Container started",
  "2026-11-16T13:14:34Z Warning Unhealthy x1: Readiness probe failed: Get \"http://10.244.1.55:9898/readyz\": dial tcp 10.244.1.55:9898: connect: connection refused"
]
```

## A4 Container logs (current and previous) (K8s API: pods/log; 2 calls, 0.01 s)

```json
{
  "current": [
    "{\"level\":\"info\",\"ts\":\"2026-11-16T13:14:34.737Z\",\"caller\":\"podinfo/main.go:170\",\"msg\":\"Starting podinfo\",\"version\":\"6.15.0\",\"revision\":\"dd507173b7b75b2312a36cabe0de5f09c1ce69c8\",\"port\":\"9898\"}",
    "{\"level\":\"info\",\"ts\":\"2026-11-16T13:14:34.738Z\",\"caller\":\"http/server.go:273\",\"msg\":\"Starting HTTP Server.\",\"addr\":\":9898\"}"
  ],
  "previous": "unavailable: the server rejected our request for an unknown reason (get pods frontend-84cbcbb56f-f5jhx)"
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
  "last_rollout_at": "2026-11-16T13:14:35Z",
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
      "revision": 20
    },
    {
      "change_cause": null,
      "created": "2026-11-16T11:48:00Z",
      "name": "frontend-6554997569",
      "replicas": 0,
      "revision": 19
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

## A6 Restarts and last termination reason (Prometheus (kube-state-metrics); 2 calls, 0.01 s)

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
      "pod": "frontend-84cbcbb56f-f5jhx",
      "service": "kps-kube-state-metrics",
      "uid": "4f140384-f4a3-4759-bd8c-3fba437c3c2c",
      "value": 0
    }
  ]
}
```

## A7 Resource usage vs. limits (Prometheus (cAdvisor + kube-state-metrics); 3 calls, 0.02 s)

```json
{
  "cpu_usage_cores_5m": [
    {
      "container": "frontend",
      "cpu": "total",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod4f140384_f4a3_4759_bd8c_3fba437c3c2c.slice/cri-containerd-77c40f8d5e7557007482dcd00620aa45ab4627d1dbdf3b68742f3040d9849997.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "77c40f8d5e7557007482dcd00620aa45ab4627d1dbdf3b68742f3040d9849997",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "frontend-84cbcbb56f-f5jhx",
      "service": "kps-kubelet",
      "value": 0.0008426894252091532
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
      "pod": "frontend-84cbcbb56f-f5jhx",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "4f140384-f4a3-4759-bd8c-3fba437c3c2c",
      "unit": "byte",
      "value": 134217728
    }
  ],
  "memory_working_set_max_15m_bytes": [
    {
      "container": "frontend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod4f140384_f4a3_4759_bd8c_3fba437c3c2c.slice/cri-containerd-77c40f8d5e7557007482dcd00620aa45ab4627d1dbdf3b68742f3040d9849997.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "77c40f8d5e7557007482dcd00620aa45ab4627d1dbdf3b68742f3040d9849997",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "frontend-84cbcbb56f-f5jhx",
      "service": "kps-kubelet",
      "value": 30588928
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
    "name=frontend-84cbcbb56f-f5jhx kind=Pod objectAPIversion=v1 objectRV=101371 eventRV=101382 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=1 msg=\"Readiness probe failed: Get \\\"http://10.244.1.55:9898/readyz\\\": dial tcp 10.244.1.55:9898: connect: connection refused\" ",
    "name=frontend-84cbcbb56f-f5jhx kind=Pod objectAPIversion=v1 objectRV=101371 eventRV=101381 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Started type=Normal count=1 msg=\"Container started\" ",
    "name=frontend-84cbcbb56f-f5jhx kind=Pod objectAPIversion=v1 objectRV=101371 eventRV=101380 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Created type=Normal count=1 msg=\"Container created\" ",
    "name=frontend-84cbcbb56f-f5jhx kind=Pod objectAPIversion=v1 objectRV=101366 eventRV=101373 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/frontend-84cbcbb56f-f5jhx to kursinis-worker2\" ",
    "name=frontend-84cbcbb56f kind=ReplicaSet objectAPIversion=apps/v1 objectRV=101365 eventRV=101369 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: frontend-84cbcbb56f-f5jhx\" ",
    "name=frontend-84cbcbb56f-f5jhx kind=Pod objectAPIversion=v1 objectRV=101371 eventRV=101378 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" "
  ],
  "pod_logs": [
    "{\"level\":\"info\",\"ts\":\"2026-11-16T13:14:34.737Z\",\"caller\":\"podinfo/main.go:170\",\"msg\":\"Starting podinfo\",\"version\":\"6.15.0\",\"revision\":\"dd507173b7b75b2312a36cabe0de5f09c1ce69c8\",\"port\":\"9898\"}\n",
    "{\"level\":\"info\",\"ts\":\"2026-11-16T13:14:34.738Z\",\"caller\":\"http/server.go:273\",\"msg\":\"Starting HTTP Server.\",\"addr\":\":9898\"}\n"
  ]
}
```

## A11 Service golden signals (RED) (Prometheus (application metrics); 3 calls, 0.01 s)

```json
{
  "error_rate_5xx_rps": [],
  "latency_p95_s": [
    {
      "value": 0.9031307172842031
    }
  ],
  "request_rate_rps": [
    {
      "value": 0.7794248126148589
    }
  ]
}
```

