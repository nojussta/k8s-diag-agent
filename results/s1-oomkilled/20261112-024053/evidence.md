# Evidence: LabPodCrashLooping demo/backend-7b8d9f4db-qgxx8

- Read-only calls: **20**, collection time: **0.05 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 0 calls, 0 s)

```json
{
  "alertname": "LabPodCrashLooping",
  "namespace": "demo",
  "pod": "backend-7b8d9f4db-qgxx8",
  "container": "backend",
  "deployment": "",
  "summary": "Pod demo/backend-7b8d9f4db-qgxx8 (backend) is crash looping",
  "startsAt": "2026-11-12T00:42:15.907Z"
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
        "finishedAt": "2026-11-12T00:41:39Z",
        "reason": "OOMKilled",
        "startedAt": "2026-11-12T00:41:35Z",
        "state": "terminated"
      },
      "name": "backend",
      "ready": false,
      "restartCount": 4,
      "state": {
        "exitCode": 137,
        "finishedAt": "2026-11-12T00:42:27Z",
        "reason": "OOMKilled",
        "startedAt": "2026-11-12T00:42:23Z",
        "state": "terminated"
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
  "2026-11-12T00:40:53Z Normal Scheduled x1: Successfully assigned demo/backend-7b8d9f4db-qgxx8 to kursinis-worker2",
  "2026-11-12T00:42:23Z Normal Pulled x5: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-12T00:42:23Z Normal Created x5: Container created",
  "2026-11-12T00:42:23Z Normal Started x5: Container started",
  "2026-11-12T00:42:23Z Warning Unhealthy x1: Liveness probe failed: Get \"http://10.244.1.27:9898/healthz\": dial tcp 10.244.1.27:9898: connect: connection refused",
  "2026-11-12T00:42:24Z Warning Unhealthy x8: Readiness probe failed: Get \"http://10.244.1.27:9898/readyz\": dial tcp 10.244.1.27:9898: connect: connection refused",
  "2026-11-12T00:42:27Z Warning BackOff x9: Back-off restarting failed container backend in pod backend-7b8d9f4db-qgxx8_demo(18a10ac7-96c6-4789-a44f-1ea632c0cb26)"
]
```

## A4 Container logs (current and previous) (K8s API: pods/log; 2 calls, 0.01 s)

```json
{
  "current": [],
  "previous": [
    "unable to retrieve container logs for containerd://3c72d44936062cdf392541d8cf7f5f071f19a248dde7999e46cca15fa8a60805"
  ]
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
      "revision": 2
    },
    {
      "change_cause": null,
      "created": "2026-11-11T23:56:25Z",
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
      "pod": "backend-7b8d9f4db-qgxx8",
      "reason": "OOMKilled",
      "service": "kps-kube-state-metrics",
      "uid": "18a10ac7-96c6-4789-a44f-1ea632c0cb26",
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
      "pod": "backend-7b8d9f4db-qgxx8",
      "service": "kps-kube-state-metrics",
      "uid": "18a10ac7-96c6-4789-a44f-1ea632c0cb26",
      "value": 2.5815352614104565
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
      "pod": "backend-7b8d9f4db-qgxx8",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "18a10ac7-96c6-4789-a44f-1ea632c0cb26",
      "unit": "byte",
      "value": 134217728
    }
  ],
  "memory_working_set_max_15m_bytes": [
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod18a10ac7_96c6_4789_a44f_1ea632c0cb26.slice/cri-containerd-c860fc30c90d3a464113d5ebc55644674d36b5e52da17e829cfaa0323805bab3.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "c860fc30c90d3a464113d5ebc55644674d36b5e52da17e829cfaa0323805bab3",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "backend-7b8d9f4db-qgxx8",
      "service": "kps-kubelet",
      "value": 93671424
    },
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod18a10ac7_96c6_4789_a44f_1ea632c0cb26.slice/cri-containerd-a33ba53483b628d973144c69ddf999eff1066babe4b5087f6bafecabca052b8a.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "a33ba53483b628d973144c69ddf999eff1066babe4b5087f6bafecabca052b8a",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "backend-7b8d9f4db-qgxx8",
      "service": "kps-kubelet",
      "value": 127569920
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
    "name=backend-7b8d9f4db-qgxx8 kind=Pod objectAPIversion=v1 objectRV=6662 eventRV=6678 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=1 msg=\"Readiness probe failed: Get \\\"http://10.244.1.27:9898/readyz\\\": dial tcp 10.244.1.27:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-qgxx8 kind=Pod objectAPIversion=v1 objectRV=6662 eventRV=6677 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Started type=Normal count=1 msg=\"Container started\" ",
    "name=backend-7b8d9f4db-qgxx8 kind=Pod objectAPIversion=v1 objectRV=6662 eventRV=6676 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Created type=Normal count=1 msg=\"Container created\" ",
    "name=backend-7b8d9f4db-qgxx8 kind=Pod objectAPIversion=v1 objectRV=6660 eventRV=6665 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/backend-7b8d9f4db-qgxx8 to kursinis-worker2\" ",
    "name=backend-7b8d9f4db kind=ReplicaSet objectAPIversion=apps/v1 objectRV=6657 eventRV=6661 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: backend-7b8d9f4db-qgxx8\" ",
    "name=backend-7b8d9f4db-qgxx8 kind=Pod objectAPIversion=v1 objectRV=6662 eventRV=6674 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-7b8d9f4db-qgxx8 kind=Pod objectAPIversion=v1 objectRV=6662 eventRV=6682 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=2 msg=\"Readiness probe failed: Get \\\"http://10.244.1.27:9898/readyz\\\": dial tcp 10.244.1.27:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-qgxx8 kind=Pod objectAPIversion=v1 objectRV=6662 eventRV=6688 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Created type=Normal count=2 msg=\"Container created\" ",
    "name=backend-7b8d9f4db-qgxx8 kind=Pod objectAPIversion=v1 objectRV=6662 eventRV=6690 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=3 msg=\"Readiness probe failed: Get \\\"http://10.244.1.27:9898/readyz\\\": dial tcp 10.244.1.27:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-qgxx8 kind=Pod objectAPIversion=v1 objectRV=6662 eventRV=6689 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Started type=Normal count=2 msg=\"Container started\" ",
    "name=backend-7b8d9f4db-qgxx8 kind=Pod objectAPIversion=v1 objectRV=6662 eventRV=6686 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=2 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-7b8d9f4db-qgxx8 kind=Pod objectAPIversion=v1 objectRV=6662 eventRV=6694 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=BackOff type=Warning count=1 msg=\"Back-off restarting failed container backend in pod backend-7b8d9f4db-qgxx8_demo(18a10ac7-96c6-4789-a44f-1ea632c0cb26)\" ",
    "name=backend-7b8d9f4db-qgxx8 kind=Pod objectAPIversion=v1 objectRV=6662 eventRV=6696 reportinginstance=kursinis-work
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
      "value": 0.7614141952869514
    }
  ]
}
```

