# Evidence: LabPodCrashLooping demo/backend-7b8d9f4db-kmgbx

- Read-only calls: **20**, collection time: **0.06 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 0 calls, 0 s)

```json
{
  "alertname": "LabPodCrashLooping",
  "namespace": "demo",
  "pod": "backend-7b8d9f4db-kmgbx",
  "container": "backend",
  "deployment": "",
  "summary": "Pod demo/backend-7b8d9f4db-kmgbx (backend) is crash looping",
  "startsAt": "2026-11-04T10:35:35.999Z"
}
```

## A2 Pod and container status (K8s API: pods; 1 calls, 0 s)

```json
{
  "conditions": {
    "ContainersReady": "False",
    "Initialized": "True",
    "PodReadyToStartContainers": "True",
    "PodScheduled": "True",
    "Ready": "False"
  },
  "containers": [
    {
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "lastState": {
        "exitCode": 137,
        "finishedAt": "2026-11-04T10:34:54Z",
        "reason": "OOMKilled",
        "startedAt": "2026-11-04T10:34:54Z",
        "state": "terminated"
      },
      "name": "backend",
      "ready": false,
      "restartCount": 3,
      "state": {
        "message": "back-off 40s restarting failed container=backend pod=backend-7b8d9f4db-kmgbx_demo(c970a157-db2e-45f3-8d47-1208a7d481a3)",
        "reason": "CrashLoopBackOff",
        "state": "waiting"
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
  "2026-11-04T10:34:19Z Normal Scheduled x1: Successfully assigned demo/backend-7b8d9f4db-kmgbx to kursinis-worker2",
  "2026-11-04T10:34:54Z Normal Pulled x4: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-04T10:34:54Z Normal Created x4: Container created",
  "2026-11-04T10:34:54Z Normal Started x4: Container started",
  "2026-11-04T10:34:54Z Warning Unhealthy x4: Readiness probe failed: Get \"http://10.244.1.12:9898/readyz\": dial tcp 10.244.1.12:9898: connect: connection refused",
  "2026-11-04T10:34:59Z Warning BackOff x9: Back-off restarting failed container backend in pod backend-7b8d9f4db-kmgbx_demo(c970a157-db2e-45f3-8d47-1208a7d481a3)"
]
```

## A4 Container logs (current and previous) (K8s API: pods/log; 2 calls, 0.01 s)

```json
{
  "current": [],
  "previous": []
}
```

## A5 Owner, rollout history and recent change (K8s API: replicasets, deployments; 3 calls, 0.01 s)

```json
{
  "conditions": [
    "Available=True (MinimumReplicasAvailable): Deployment has minimum availability.",
    "Progressing=True (ReplicaSetUpdated): ReplicaSet \"backend-7b8d9f4db\" is progressing."
  ],
  "deployment": "backend",
  "replicas": {
    "availableReplicas": 2,
    "readyReplicas": 2,
    "replicas": 3,
    "unavailableReplicas": 1,
    "updatedReplicas": 1
  },
  "revisions": [
    {
      "change_cause": "release 2.0: warm in-memory cache on startup",
      "created": "2026-11-04T10:34:19Z",
      "name": "backend-7b8d9f4db",
      "replicas": 1,
      "revision": 2
    },
    {
      "change_cause": null,
      "created": "2026-11-04T10:33:15Z",
      "name": "backend-5b97dd7dcb",
      "replicas": 2,
      "revision": 1
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
        "--stress-memory=256"
      ]
    }
  ]
}
```

## A6 Restarts and last termination reason (Prometheus (kube-state-metrics); 2 calls, 0.01 s)

```json
{
  "last_terminated_reason": [
    {
      "__name__": "kube_pod_container_status_last_terminated_reason",
      "container": "backend",
      "endpoint": "http",
      "instance": "10.244.1.4:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "pod": "backend-7b8d9f4db-kmgbx",
      "reason": "OOMKilled",
      "service": "kps-kube-state-metrics",
      "uid": "c970a157-db2e-45f3-8d47-1208a7d481a3",
      "value": 1
    }
  ],
  "restarts_15m": [
    {
      "container": "backend",
      "endpoint": "http",
      "instance": "10.244.1.4:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "pod": "backend-7b8d9f4db-kmgbx",
      "service": "kps-kube-state-metrics",
      "uid": "c970a157-db2e-45f3-8d47-1208a7d481a3",
      "value": 2.2319987200511977
    }
  ]
}
```

## A7 Resource usage vs. limits (Prometheus (cAdvisor + kube-state-metrics); 3 calls, 0.01 s)

```json
{
  "cpu_usage_cores_5m": [],
  "memory_limit_bytes": [
    {
      "__name__": "kube_pod_container_resource_limits",
      "container": "backend",
      "endpoint": "http",
      "instance": "10.244.1.4:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "backend-7b8d9f4db-kmgbx",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "c970a157-db2e-45f3-8d47-1208a7d481a3",
      "unit": "byte",
      "value": 134217728
    }
  ],
  "memory_working_set_max_15m_bytes": []
}
```

## A8 Service endpoints of the workload (K8s API: services, endpointslices; 2 calls, 0 s)

```json
[
  {
    "endpoints": 3,
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

## A10 Log and event history (Loki) (Loki; 2 calls, 0.01 s)

```json
{
  "events": [
    "name=backend-7b8d9f4db-kmgbx kind=Pod objectAPIversion=v1 objectRV=1782 eventRV=1794 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=1 msg=\"Readiness probe failed: Get \\\"http://10.244.1.12:9898/readyz\\\": dial tcp 10.244.1.12:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-kmgbx kind=Pod objectAPIversion=v1 objectRV=1782 eventRV=1793 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Started type=Normal count=1 msg=\"Container started\" ",
    "name=backend-7b8d9f4db-kmgbx kind=Pod objectAPIversion=v1 objectRV=1782 eventRV=1792 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Created type=Normal count=1 msg=\"Container created\" ",
    "name=backend-7b8d9f4db-kmgbx kind=Pod objectAPIversion=v1 objectRV=1778 eventRV=1784 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/backend-7b8d9f4db-kmgbx to kursinis-worker2\" ",
    "name=backend-7b8d9f4db kind=ReplicaSet objectAPIversion=apps/v1 objectRV=1775 eventRV=1780 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: backend-7b8d9f4db-kmgbx\" ",
    "name=backend-7b8d9f4db-kmgbx kind=Pod objectAPIversion=v1 objectRV=1782 eventRV=1790 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-7b8d9f4db-kmgbx kind=Pod objectAPIversion=v1 objectRV=1782 eventRV=1802 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=2 msg=\"Readiness probe failed: Get \\\"http://10.244.1.12:9898/readyz\\\": dial tcp 10.244.1.12:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-kmgbx kind=Pod objectAPIversion=v1 objectRV=1782 eventRV=1801 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Started type=Normal count=2 msg=\"Container started\" ",
    "name=backend-7b8d9f4db-kmgbx kind=Pod objectAPIversion=v1 objectRV=1782 eventRV=1800 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Created type=Normal count=2 msg=\"Container created\" ",
    "name=backend-7b8d9f4db-kmgbx kind=Pod objectAPIversion=v1 objectRV=1782 eventRV=1798 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=2 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-7b8d9f4db-kmgbx kind=Pod objectAPIversion=v1 objectRV=1782 eventRV=1806 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=BackOff type=Warning count=1 msg=\"Back-off restarting failed container backend in pod backend-7b8d9f4db-kmgbx_demo(c970a157-db2e-45f3-8d47-1208a7d481a3)\" ",
    "name=backend-7b8d9f4db-kmgbx kind=Pod objectAPIversion=v1 objectRV=1782 eventRV=1809 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=BackOff type=Warning count=2 msg=\"Back-off restarting failed container backend in pod backend-7b8d9f4db-kmgbx_demo(c970a157-db2e-45f3-8d47-1208a7d481a3)\" ",
    "name=backend-7b8d9f4db-kmgbx kind=Pod objectAPIversion=v1 objectRV=1782 eventRV=1824 reportinginstance=kursinis-worker2 re
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
      "value": 0.40446032320472813
    }
  ]
}
```

