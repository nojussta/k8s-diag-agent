# Evidence: LabPodCrashLooping demo/backend-7b8d9f4db-gmppc

- Read-only calls: **20**, collection time: **0.03 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 0 calls, 0 s)

```json
{
  "alertname": "LabPodCrashLooping",
  "namespace": "demo",
  "pod": "backend-7b8d9f4db-gmppc",
  "container": "backend",
  "deployment": "",
  "summary": "Pod demo/backend-7b8d9f4db-gmppc (backend) is crash looping",
  "startsAt": "2026-11-12T02:37:00.907Z"
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
        "finishedAt": "2026-11-12T02:36:59Z",
        "reason": "OOMKilled",
        "startedAt": "2026-11-12T02:36:54Z",
        "state": "terminated"
      },
      "name": "backend",
      "ready": false,
      "restartCount": 4,
      "state": {
        "message": "back-off 1m20s restarting failed container=backend pod=backend-7b8d9f4db-gmppc_demo(91d7d247-1821-4f44-b3ce-4715b00caf23)",
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
  "2026-11-12T02:35:09Z Normal Scheduled x1: Successfully assigned demo/backend-7b8d9f4db-gmppc to kursinis-worker2",
  "2026-11-12T02:36:54Z Normal Pulled x5: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-12T02:36:54Z Normal Created x5: Container created",
  "2026-11-12T02:36:54Z Normal Started x5: Container started",
  "2026-11-12T02:37:00Z Warning Unhealthy x11: Readiness probe failed: Get \"http://10.244.1.54:9898/readyz\": dial tcp 10.244.1.54:9898: connect: connection refused",
  "2026-11-12T02:37:00Z Warning Unhealthy x2: Liveness probe failed: Get \"http://10.244.1.54:9898/healthz\": dial tcp 10.244.1.54:9898: connect: connection refused",
  "2026-11-12T02:37:10Z Warning BackOff x8: Back-off restarting failed container backend in pod backend-7b8d9f4db-gmppc_demo(91d7d247-1821-4f44-b3ce-4715b00caf23)"
]
```

## A4 Container logs (current and previous) (K8s API: pods/log; 2 calls, 0 s)

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
      "created": "2026-11-12T00:40:53Z",
      "name": "backend-7b8d9f4db",
      "replicas": 1,
      "revision": 10
    },
    {
      "change_cause": null,
      "created": "2026-11-11T23:56:25Z",
      "name": "backend-5b97dd7dcb",
      "replicas": 2,
      "revision": 9
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

## A6 Restarts and last termination reason (Prometheus (kube-state-metrics); 2 calls, 0 s)

```json
{
  "last_terminated_reason": [
    {
      "__name__": "kube_pod_container_status_last_terminated_reason",
      "container": "backend",
      "endpoint": "http",
      "instance": "10.244.2.4:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "pod": "backend-7b8d9f4db-gmppc",
      "reason": "OOMKilled",
      "service": "kps-kube-state-metrics",
      "uid": "91d7d247-1821-4f44-b3ce-4715b00caf23",
      "value": 1
    }
  ],
  "restarts_15m": [
    {
      "container": "backend",
      "endpoint": "http",
      "instance": "10.244.2.4:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "pod": "backend-7b8d9f4db-gmppc",
      "service": "kps-kube-state-metrics",
      "uid": "91d7d247-1821-4f44-b3ce-4715b00caf23",
      "value": 3.489659765467154
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
      "instance": "10.244.2.4:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "backend-7b8d9f4db-gmppc",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "91d7d247-1821-4f44-b3ce-4715b00caf23",
      "unit": "byte",
      "value": 134217728
    }
  ],
  "memory_working_set_max_15m_bytes": [
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod91d7d247_1821_4f44_b3ce_4715b00caf23.slice/cri-containerd-4dd242f0f6bd6f94bb96e8af5325c7db54aea2348f257d13392c9e3caa66421b.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "4dd242f0f6bd6f94bb96e8af5325c7db54aea2348f257d13392c9e3caa66421b",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "backend-7b8d9f4db-gmppc",
      "service": "kps-kubelet",
      "value": 125923328
    },
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod91d7d247_1821_4f44_b3ce_4715b00caf23.slice/cri-containerd-18d12cb6c48dee36e81ba2cb7bc88cf9e55aa1fbd4a70c9f7929b7dd98e97185.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "18d12cb6c48dee36e81ba2cb7bc88cf9e55aa1fbd4a70c9f7929b7dd98e97185",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "backend-7b8d9f4db-gmppc",
      "service": "kps-kubelet",
      "value": 126074880
    }
  ]
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
    "name=backend-7b8d9f4db-gmppc kind=Pod objectAPIversion=v1 objectRV=19424 eventRV=19430 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/backend-7b8d9f4db-gmppc to kursinis-worker2\" ",
    "name=backend-7b8d9f4db kind=ReplicaSet objectAPIversion=apps/v1 objectRV=19422 eventRV=19426 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: backend-7b8d9f4db-gmppc\" ",
    "name=backend-7b8d9f4db-gmppc kind=Pod objectAPIversion=v1 objectRV=19428 eventRV=19440 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=1 msg=\"Readiness probe failed: Get \\\"http://10.244.1.54:9898/readyz\\\": dial tcp 10.244.1.54:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-gmppc kind=Pod objectAPIversion=v1 objectRV=19428 eventRV=19439 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Started type=Normal count=1 msg=\"Container started\" ",
    "name=backend-7b8d9f4db-gmppc kind=Pod objectAPIversion=v1 objectRV=19428 eventRV=19438 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Created type=Normal count=1 msg=\"Container created\" ",
    "name=backend-7b8d9f4db-gmppc kind=Pod objectAPIversion=v1 objectRV=19428 eventRV=19436 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-7b8d9f4db-gmppc kind=Pod objectAPIversion=v1 objectRV=19428 eventRV=19444 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=2 msg=\"Readiness probe failed: Get \\\"http://10.244.1.54:9898/readyz\\\": dial tcp 10.244.1.54:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-gmppc kind=Pod objectAPIversion=v1 objectRV=19428 eventRV=19455 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=3 msg=\"Readiness probe failed: Get \\\"http://10.244.1.54:9898/readyz\\\": dial tcp 10.244.1.54:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-gmppc kind=Pod objectAPIversion=v1 objectRV=19428 eventRV=19454 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Started type=Normal count=2 msg=\"Container started\" ",
    "name=backend-7b8d9f4db-gmppc kind=Pod objectAPIversion=v1 objectRV=19428 eventRV=19453 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Created type=Normal count=2 msg=\"Container created\" ",
    "name=backend-7b8d9f4db-gmppc kind=Pod objectAPIversion=v1 objectRV=19428 eventRV=19451 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=2 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-7b8d9f4db-gmppc kind=Pod objectAPIversion=v1 objectRV=19428 eventRV=19459 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=4 msg=\"Readiness probe failed: Get \\\"http://10.244.1.54:9898/readyz\\\": dial tcp 10.244.1.54:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-gmppc kind=Pod objectAPIversion=v1 objectRV=19428 eventRV=19464
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
      "value": 0.9929859403015449
    }
  ]
}
```

