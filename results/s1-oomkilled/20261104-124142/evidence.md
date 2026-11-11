# Evidence: LabPodCrashLooping demo/backend-7b8d9f4db-hjbpx

- Read-only calls: **20**, collection time: **0.08 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 0 calls, 0 s)

```json
{
  "alertname": "LabPodCrashLooping",
  "namespace": "demo",
  "pod": "backend-7b8d9f4db-hjbpx",
  "container": "backend",
  "deployment": "",
  "summary": "Pod demo/backend-7b8d9f4db-hjbpx (backend) is crash looping",
  "startsAt": "2026-11-04T10:43:20.999Z"
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
        "finishedAt": "2026-11-04T10:43:14Z",
        "reason": "OOMKilled",
        "startedAt": "2026-11-04T10:43:14Z",
        "state": "terminated"
      },
      "name": "backend",
      "ready": false,
      "restartCount": 4,
      "state": {
        "message": "back-off 1m20s restarting failed container=backend pod=backend-7b8d9f4db-hjbpx_demo(48a50715-2cb8-4927-8044-64c1e72a27eb)",
        "reason": "CrashLoopBackOff",
        "state": "waiting"
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

## A3 Events for the pod (K8s API: events; 1 calls, 0 s)

```json
[
  "2026-11-04T10:41:42Z Normal Scheduled x1: Successfully assigned demo/backend-7b8d9f4db-hjbpx to kursinis-worker",
  "2026-11-04T10:43:14Z Normal Pulled x5: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-04T10:43:14Z Normal Created x5: Container created",
  "2026-11-04T10:43:14Z Normal Started x5: Container started",
  "2026-11-04T10:43:14Z Warning Unhealthy x8: Readiness probe failed: Get \"http://10.244.2.8:9898/readyz\": dial tcp 10.244.2.8:9898: connect: connection refused",
  "2026-11-04T10:43:15Z Warning BackOff x9: Back-off restarting failed container backend in pod backend-7b8d9f4db-hjbpx_demo(48a50715-2cb8-4927-8044-64c1e72a27eb)"
]
```

## A4 Container logs (current and previous) (K8s API: pods/log; 2 calls, 0.01 s)

```json
{
  "current": [],
  "previous": []
}
```

## A5 Owner, rollout history and recent change (K8s API: replicasets, deployments; 3 calls, 0 s)

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
      "revision": 6
    },
    {
      "change_cause": null,
      "created": "2026-11-04T10:33:15Z",
      "name": "backend-5b97dd7dcb",
      "replicas": 2,
      "revision": 5
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
      "pod": "backend-7b8d9f4db-hjbpx",
      "reason": "OOMKilled",
      "service": "kps-kube-state-metrics",
      "uid": "48a50715-2cb8-4927-8044-64c1e72a27eb",
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
      "pod": "backend-7b8d9f4db-hjbpx",
      "service": "kps-kube-state-metrics",
      "uid": "48a50715-2cb8-4927-8044-64c1e72a27eb",
      "value": 3.449293356888103
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
      "node": "kursinis-worker",
      "pod": "backend-7b8d9f4db-hjbpx",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "48a50715-2cb8-4927-8044-64c1e72a27eb",
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
  "node": "kursinis-worker"
}
```

## A10 Log and event history (Loki) (Loki; 2 calls, 0.04 s)

```json
{
  "events": [
    "name=backend-7b8d9f4db-hjbpx kind=Pod objectAPIversion=v1 objectRV=2655 eventRV=2670 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Unhealthy type=Warning count=2 msg=\"Readiness probe failed: Get \\\"http://10.244.2.8:9898/readyz\\\": dial tcp 10.244.2.8:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-hjbpx kind=Pod objectAPIversion=v1 objectRV=2655 eventRV=2666 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Unhealthy type=Warning count=1 msg=\"Readiness probe failed: Get \\\"http://10.244.2.8:9898/readyz\\\": dial tcp 10.244.2.8:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-hjbpx kind=Pod objectAPIversion=v1 objectRV=2655 eventRV=2665 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Started type=Normal count=1 msg=\"Container started\" ",
    "name=backend-7b8d9f4db-hjbpx kind=Pod objectAPIversion=v1 objectRV=2655 eventRV=2664 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Created type=Normal count=1 msg=\"Container created\" ",
    "name=backend-7b8d9f4db-hjbpx kind=Pod objectAPIversion=v1 objectRV=2650 eventRV=2657 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/backend-7b8d9f4db-hjbpx to kursinis-worker\" ",
    "name=backend-7b8d9f4db kind=ReplicaSet objectAPIversion=apps/v1 objectRV=2649 eventRV=2654 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: backend-7b8d9f4db-hjbpx\" ",
    "name=backend-7b8d9f4db-hjbpx kind=Pod objectAPIversion=v1 objectRV=2655 eventRV=2662 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-7b8d9f4db-hjbpx kind=Pod objectAPIversion=v1 objectRV=2655 eventRV=2675 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Unhealthy type=Warning count=3 msg=\"Readiness probe failed: Get \\\"http://10.244.2.8:9898/readyz\\\": dial tcp 10.244.2.8:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-hjbpx kind=Pod objectAPIversion=v1 objectRV=2655 eventRV=2674 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Started type=Normal count=2 msg=\"Container started\" ",
    "name=backend-7b8d9f4db-hjbpx kind=Pod objectAPIversion=v1 objectRV=2655 eventRV=2673 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Created type=Normal count=2 msg=\"Container created\" ",
    "name=backend-7b8d9f4db-hjbpx kind=Pod objectAPIversion=v1 objectRV=2655 eventRV=2671 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Pulled type=Normal count=2 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-7b8d9f4db-hjbpx kind=Pod objectAPIversion=v1 objectRV=2655 eventRV=2679 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=BackOff type=Warning count=1 msg=\"Back-off restarting failed container backend in pod backend-7b8d9f4db-hjbpx_demo(48a50715-2cb8-4927-8044-64c1e72a27eb)\" ",
    "name=backend-7b8d9f4db-hjbpx kind=Pod objectAPIversion=v1 objectRV=2655 eventRV=2682 reportinginstance=kursinis-worker reportingcontroller=kube
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
      "value": 0.9894563253276256
    }
  ]
}
```

