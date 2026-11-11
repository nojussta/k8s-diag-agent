# Evidence: LabPodCrashLooping demo/backend-8698b7497d-sh9wq

- Read-only calls: **20**, collection time: **0.03 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 0 calls, 0 s)

```json
{
  "alertname": "LabPodCrashLooping",
  "namespace": "demo",
  "pod": "backend-8698b7497d-sh9wq",
  "container": "backend",
  "deployment": "",
  "summary": "Pod demo/backend-8698b7497d-sh9wq (backend) is crash looping",
  "startsAt": "2026-11-11T11:31:57.348Z"
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
        "finishedAt": "2026-11-11T11:31:22Z",
        "reason": "OOMKilled",
        "startedAt": "2026-11-11T11:31:18Z",
        "state": "terminated"
      },
      "name": "backend",
      "ready": false,
      "restartCount": 3,
      "state": {
        "message": "back-off 40s restarting failed container=backend pod=backend-8698b7497d-sh9wq_demo(283dab2b-1942-4441-b58a-14552b08900b)",
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
  "2026-11-11T11:30:33Z Normal Scheduled x1: Successfully assigned demo/backend-8698b7497d-sh9wq to kursinis-worker",
  "2026-11-11T11:31:18Z Normal Pulled x4: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-11T11:31:18Z Normal Created x4: Container created",
  "2026-11-11T11:31:18Z Normal Started x4: Container started",
  "2026-11-11T11:31:19Z Warning Unhealthy x7: Readiness probe failed: Get \"http://10.244.2.26:9898/readyz\": dial tcp 10.244.2.26:9898: connect: connection refused",
  "2026-11-11T11:31:23Z Warning BackOff x7: Back-off restarting failed container backend in pod backend-8698b7497d-sh9wq_demo(283dab2b-1942-4441-b58a-14552b08900b)"
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
    "Progressing=True (ReplicaSetUpdated): ReplicaSet \"backend-8698b7497d\" is progressing."
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
      "created": "2026-11-11T11:22:13Z",
      "name": "backend-8698b7497d",
      "replicas": 1,
      "revision": 40
    },
    {
      "change_cause": "release 2.0: warm in-memory cache on startup",
      "created": "2026-11-11T11:19:33Z",
      "name": "backend-6dfbd8dd76",
      "replicas": 2,
      "revision": 39
    },
    {
      "change_cause": "release 2.0: warm in-memory cache on startup",
      "created": "2026-11-11T11:10:55Z",
      "name": "backend-75dcdb6575",
      "replicas": 0,
      "revision": 34
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
      "instance": "10.244.2.2:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "pod": "backend-8698b7497d-sh9wq",
      "reason": "OOMKilled",
      "service": "kps-kube-state-metrics",
      "uid": "283dab2b-1942-4441-b58a-14552b08900b",
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
      "pod": "backend-8698b7497d-sh9wq",
      "service": "kps-kube-state-metrics",
      "uid": "283dab2b-1942-4441-b58a-14552b08900b",
      "value": 3.2020243544731346
    }
  ]
}
```

## A7 Resource usage vs. limits (Prometheus (cAdvisor + kube-state-metrics); 3 calls, 0 s)

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
      "node": "kursinis-worker",
      "pod": "backend-8698b7497d-sh9wq",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "283dab2b-1942-4441-b58a-14552b08900b",
      "unit": "byte",
      "value": 134217728
    }
  ],
  "memory_working_set_max_15m_bytes": [
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod283dab2b_1942_4441_b58a_14552b08900b.slice/cri-containerd-eb4abbae460843b1f8e0b932c88ee9c37af6050e8203564b7308abc00351aee5.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "eb4abbae460843b1f8e0b932c88ee9c37af6050e8203564b7308abc00351aee5",
      "namespace": "demo",
      "node": "kursinis-worker",
      "pod": "backend-8698b7497d-sh9wq",
      "service": "kps-kubelet",
      "value": 129130496
    },
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod283dab2b_1942_4441_b58a_14552b08900b.slice/cri-containerd-123b6263df9db60ef2b05f2a0e3dcb7c21710ecb55312885b7360849c3fcfb80.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "123b6263df9db60ef2b05f2a0e3dcb7c21710ecb55312885b7360849c3fcfb80",
      "namespace": "demo",
      "node": "kursinis-worker",
      "pod": "backend-8698b7497d-sh9wq",
      "service": "kps-kubelet",
      "value": 127852544
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
  "node": "kursinis-worker"
}
```

## A10 Log and event history (Loki) (Loki; 2 calls, 0.01 s)

```json
{
  "events": [
    "name=backend-8698b7497d-sh9wq kind=Pod objectAPIversion=v1 objectRV=24434 eventRV=24445 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Unhealthy type=Warning count=1 msg=\"Readiness probe failed: Get \\\"http://10.244.2.26:9898/readyz\\\": dial tcp 10.244.2.26:9898: connect: connection refused\" ",
    "name=backend-8698b7497d-sh9wq kind=Pod objectAPIversion=v1 objectRV=24434 eventRV=24444 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Started type=Normal count=1 msg=\"Container started\" ",
    "name=backend-8698b7497d-sh9wq kind=Pod objectAPIversion=v1 objectRV=24434 eventRV=24443 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Created type=Normal count=1 msg=\"Container created\" ",
    "name=backend-8698b7497d-sh9wq kind=Pod objectAPIversion=v1 objectRV=24430 eventRV=24436 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/backend-8698b7497d-sh9wq to kursinis-worker\" ",
    "name=backend-8698b7497d kind=ReplicaSet objectAPIversion=apps/v1 objectRV=24429 eventRV=24433 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: backend-8698b7497d-sh9wq\" ",
    "name=backend-8698b7497d-sh9wq kind=Pod objectAPIversion=v1 objectRV=24434 eventRV=24441 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-8698b7497d-sh9wq kind=Pod objectAPIversion=v1 objectRV=24434 eventRV=24449 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Unhealthy type=Warning count=2 msg=\"Readiness probe failed: Get \\\"http://10.244.2.26:9898/readyz\\\": dial tcp 10.244.2.26:9898: connect: connection refused\" ",
    "name=backend-8698b7497d-sh9wq kind=Pod objectAPIversion=v1 objectRV=24434 eventRV=24459 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Unhealthy type=Warning count=3 msg=\"Readiness probe failed: Get \\\"http://10.244.2.26:9898/readyz\\\": dial tcp 10.244.2.26:9898: connect: connection refused\" ",
    "name=backend-8698b7497d-sh9wq kind=Pod objectAPIversion=v1 objectRV=24434 eventRV=24458 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Started type=Normal count=2 msg=\"Container started\" ",
    "name=backend-8698b7497d-sh9wq kind=Pod objectAPIversion=v1 objectRV=24434 eventRV=24457 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Created type=Normal count=2 msg=\"Container created\" ",
    "name=backend-8698b7497d-sh9wq kind=Pod objectAPIversion=v1 objectRV=24434 eventRV=24455 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Pulled type=Normal count=2 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-8698b7497d-sh9wq kind=Pod objectAPIversion=v1 objectRV=24434 eventRV=24461 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=BackOff type=Warning count=1 msg=\"Back-off restarting failed container backend in pod backend-8698b7497d-sh9wq_demo(283dab2b-1942-4441-b58a-14552b08900b)\" ",
    "name=backend-8698b7497d-sh9wq kind=Pod objectAPIversion=v1 objectRV=24434 eventRV=24466 reportingi
```

## A11 Service golden signals (RED) (Prometheus (application metrics); 3 calls, 0.01 s)

```json
{
  "error_rate_5xx_rps": [],
  "latency_p95_s": [
    {
      "value": 0.004750000000000001
    }
  ],
  "request_rate_rps": [
    {
      "value": 0.992954583730983
    }
  ]
}
```

