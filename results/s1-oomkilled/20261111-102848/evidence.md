# Evidence: LabPodCrashLooping demo/backend-7b8d9f4db-7v89j

- Read-only calls: **20**, collection time: **0.08 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 0 calls, 0 s)

```json
{
  "alertname": "LabPodCrashLooping",
  "namespace": "demo",
  "pod": "backend-7b8d9f4db-7v89j",
  "container": "backend",
  "deployment": "",
  "summary": "Pod demo/backend-7b8d9f4db-7v89j (backend) is crash looping",
  "startsAt": "2026-11-11T08:30:12.348Z"
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
        "finishedAt": "2026-11-11T08:29:34Z",
        "reason": "OOMKilled",
        "startedAt": "2026-11-11T08:29:34Z",
        "state": "terminated"
      },
      "name": "backend",
      "ready": false,
      "restartCount": 4,
      "state": {
        "startedAt": "2026-11-11T08:30:23Z",
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
  "2026-11-11T08:28:48Z Normal Scheduled x1: Successfully assigned demo/backend-7b8d9f4db-7v89j to kursinis-worker2",
  "2026-11-11T08:29:39Z Warning BackOff x7: Back-off restarting failed container backend in pod backend-7b8d9f4db-7v89j_demo(344238f3-a954-4212-8bf1-f72914fe4abc)",
  "2026-11-11T08:30:23Z Normal Pulled x5: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-11T08:30:23Z Normal Created x5: Container created",
  "2026-11-11T08:30:23Z Normal Started x5: Container started",
  "2026-11-11T08:30:24Z Warning Unhealthy x10: Readiness probe failed: Get \"http://10.244.1.10:9898/readyz\": dial tcp 10.244.1.10:9898: connect: connection refused"
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
      "created": "2026-11-11T07:55:14Z",
      "name": "backend-7b8d9f4db",
      "replicas": 1,
      "revision": 8
    },
    {
      "change_cause": null,
      "created": "2026-11-11T07:52:36Z",
      "name": "backend-5b97dd7dcb",
      "replicas": 2,
      "revision": 7
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
      "instance": "10.244.2.2:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "pod": "backend-7b8d9f4db-7v89j",
      "reason": "OOMKilled",
      "service": "kps-kube-state-metrics",
      "uid": "344238f3-a954-4212-8bf1-f72914fe4abc",
      "value": 1
    }
  ],
  "restarts_15m": [
    {
      "container": "backend",
      "endpoint": "http",
      "instance": "10.244.2.2:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "pod": "backend-7b8d9f4db-7v89j",
      "service": "kps-kube-state-metrics",
      "uid": "344238f3-a954-4212-8bf1-f72914fe4abc",
      "value": 2.3807248193285155
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
      "instance": "10.244.2.2:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "backend-7b8d9f4db-7v89j",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "344238f3-a954-4212-8bf1-f72914fe4abc",
      "unit": "byte",
      "value": 134217728
    }
  ],
  "memory_working_set_max_15m_bytes": [
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod344238f3_a954_4212_8bf1_f72914fe4abc.slice/cri-containerd-d399feb811c55ca60c26972c1d7818ad581e7cba35297b0c6bdb2d9d557532a6.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.3:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "d399feb811c55ca60c26972c1d7818ad581e7cba35297b0c6bdb2d9d557532a6",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "backend-7b8d9f4db-7v89j",
      "service": "kps-kubelet",
      "value": 18825216
    },
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod344238f3_a954_4212_8bf1_f72914fe4abc.slice/cri-containerd-a344c09ae3a2e5c9b48b51f560fb1b466fe2ab0bde8ac8fa1ce5b953da0bafa9.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.3:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "a344c09ae3a2e5c9b48b51f560fb1b466fe2ab0bde8ac8fa1ce5b953da0bafa9",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "backend-7b8d9f4db-7v89j",
      "service": "kps-kubelet",
      "value": 129290240
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

## A10 Log and event history (Loki) (Loki; 2 calls, 0.02 s)

```json
{
  "events": [
    "name=backend-7b8d9f4db-7v89j kind=Pod objectAPIversion=v1 objectRV=5379 eventRV=5384 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/backend-7b8d9f4db-7v89j to kursinis-worker2\" ",
    "name=backend-7b8d9f4db kind=ReplicaSet objectAPIversion=apps/v1 objectRV=5376 eventRV=5380 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: backend-7b8d9f4db-7v89j\" ",
    "name=backend-7b8d9f4db-7v89j kind=Pod objectAPIversion=v1 objectRV=5382 eventRV=5397 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=1 msg=\"Readiness probe failed: Get \\\"http://10.244.1.10:9898/readyz\\\": dial tcp 10.244.1.10:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-7v89j kind=Pod objectAPIversion=v1 objectRV=5382 eventRV=5396 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Started type=Normal count=1 msg=\"Container started\" ",
    "name=backend-7b8d9f4db-7v89j kind=Pod objectAPIversion=v1 objectRV=5382 eventRV=5395 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Created type=Normal count=1 msg=\"Container created\" ",
    "name=backend-7b8d9f4db-7v89j kind=Pod objectAPIversion=v1 objectRV=5382 eventRV=5393 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-7b8d9f4db-7v89j kind=Pod objectAPIversion=v1 objectRV=5382 eventRV=5402 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=2 msg=\"Readiness probe failed: Get \\\"http://10.244.1.10:9898/readyz\\\": dial tcp 10.244.1.10:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-7v89j kind=Pod objectAPIversion=v1 objectRV=5382 eventRV=5412 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=3 msg=\"Readiness probe failed: Get \\\"http://10.244.1.10:9898/readyz\\\": dial tcp 10.244.1.10:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-7v89j kind=Pod objectAPIversion=v1 objectRV=5382 eventRV=5411 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Started type=Normal count=2 msg=\"Container started\" ",
    "name=backend-7b8d9f4db-7v89j kind=Pod objectAPIversion=v1 objectRV=5382 eventRV=5410 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Created type=Normal count=2 msg=\"Container created\" ",
    "name=backend-7b8d9f4db-7v89j kind=Pod objectAPIversion=v1 objectRV=5382 eventRV=5408 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=2 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-7b8d9f4db-7v89j kind=Pod objectAPIversion=v1 objectRV=5382 eventRV=5416 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=4 msg=\"Readiness probe failed: Get \\\"http://10.244.1.10:9898/readyz\\\": dial tcp 10.244.1.10:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-7v89j kind=Pod objectAPIversion=v1 objectRV=5382 eventRV=5422 reportinginstance=kursini
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
      "value": 0.9929824561403507
    }
  ]
}
```

