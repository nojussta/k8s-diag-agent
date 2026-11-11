# Evidence: LabPodCrashLooping demo/backend-8698b7497d-w4k29

- Read-only calls: **20**, collection time: **0.09 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 0 calls, 0 s)

```json
{
  "alertname": "LabPodCrashLooping",
  "namespace": "demo",
  "pod": "backend-8698b7497d-w4k29",
  "container": "backend",
  "deployment": "",
  "summary": "Pod demo/backend-8698b7497d-w4k29 (backend) is crash looping",
  "startsAt": "2026-11-11T11:23:42.348Z"
}
```

## A2 Pod and container status (K8s API: pods; 1 calls, 0.01 s)

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
        "finishedAt": "2026-11-11T11:23:02Z",
        "reason": "OOMKilled",
        "startedAt": "2026-11-11T11:22:58Z",
        "state": "terminated"
      },
      "name": "backend",
      "ready": false,
      "restartCount": 4,
      "state": {
        "exitCode": 137,
        "finishedAt": "2026-11-11T11:23:48Z",
        "reason": "OOMKilled",
        "startedAt": "2026-11-11T11:23:44Z",
        "state": "terminated"
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
  "2026-11-11T11:22:13Z Normal Scheduled x1: Successfully assigned demo/backend-8698b7497d-w4k29 to kursinis-worker",
  "2026-11-11T11:22:34Z Warning Unhealthy x1: Liveness probe failed: Get \"http://10.244.2.24:9898/healthz\": dial tcp 10.244.2.24:9898: connect: connection refused",
  "2026-11-11T11:23:44Z Normal Pulled x5: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-11T11:23:44Z Normal Created x5: Container created",
  "2026-11-11T11:23:44Z Normal Started x5: Container started",
  "2026-11-11T11:23:45Z Warning Unhealthy x10: Readiness probe failed: Get \"http://10.244.2.24:9898/readyz\": dial tcp 10.244.2.24:9898: connect: connection refused",
  "2026-11-11T11:23:49Z Warning BackOff x7: Back-off restarting failed container backend in pod backend-8698b7497d-w4k29_demo(c4e19532-c7e8-4a07-a8a2-cead45fb8b2c)"
]
```

## A4 Container logs (current and previous) (K8s API: pods/log; 2 calls, 0.01 s)

```json
{
  "current": [],
  "previous": [
    "unable to retrieve container logs for containerd://4753e769f95b4cacfb8359759636339becfdc1612b1f7150cbc5bf903b6c76ef"
  ]
}
```

## A5 Owner, rollout history and recent change (K8s API: replicasets, deployments; 3 calls, 0.01 s)

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
      "revision": 36
    },
    {
      "change_cause": "release 2.0: warm in-memory cache on startup",
      "created": "2026-11-11T11:19:33Z",
      "name": "backend-6dfbd8dd76",
      "replicas": 2,
      "revision": 35
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
      "pod": "backend-8698b7497d-w4k29",
      "reason": "OOMKilled",
      "service": "kps-kube-state-metrics",
      "uid": "c4e19532-c7e8-4a07-a8a2-cead45fb8b2c",
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
      "pod": "backend-8698b7497d-w4k29",
      "service": "kps-kube-state-metrics",
      "uid": "c4e19532-c7e8-4a07-a8a2-cead45fb8b2c",
      "value": 4.202168792498111
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
      "node": "kursinis-worker",
      "pod": "backend-8698b7497d-w4k29",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "c4e19532-c7e8-4a07-a8a2-cead45fb8b2c",
      "unit": "byte",
      "value": 134217728
    }
  ],
  "memory_working_set_max_15m_bytes": [
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-podc4e19532_c7e8_4a07_a8a2_cead45fb8b2c.slice/cri-containerd-9b50255537fa50821ff0b658240f10a91a6a21fa497dddc750da8236451b8631.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "9b50255537fa50821ff0b658240f10a91a6a21fa497dddc750da8236451b8631",
      "namespace": "demo",
      "node": "kursinis-worker",
      "pod": "backend-8698b7497d-w4k29",
      "service": "kps-kubelet",
      "value": 132759552
    },
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-podc4e19532_c7e8_4a07_a8a2_cead45fb8b2c.slice/cri-containerd-4753e769f95b4cacfb8359759636339becfdc1612b1f7150cbc5bf903b6c76ef.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "4753e769f95b4cacfb8359759636339becfdc1612b1f7150cbc5bf903b6c76ef",
      "namespace": "demo",
      "node": "kursinis-worker",
      "pod": "backend-8698b7497d-w4k29",
      "service": "kps-kubelet",
      "value": 131612672
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

## A10 Log and event history (Loki) (Loki; 2 calls, 0.02 s)

```json
{
  "events": [
    "name=backend-8698b7497d-w4k29 kind=Pod objectAPIversion=v1 objectRV=23459 eventRV=23465 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/backend-8698b7497d-w4k29 to kursinis-worker\" ",
    "name=backend-8698b7497d kind=ReplicaSet objectAPIversion=apps/v1 objectRV=23457 eventRV=23461 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: backend-8698b7497d-w4k29\" ",
    "name=backend-8698b7497d-w4k29 kind=Pod objectAPIversion=v1 objectRV=23462 eventRV=23472 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Unhealthy type=Warning count=1 msg=\"Readiness probe failed: Get \\\"http://10.244.2.24:9898/readyz\\\": dial tcp 10.244.2.24:9898: connect: connection refused\" ",
    "name=backend-8698b7497d-w4k29 kind=Pod objectAPIversion=v1 objectRV=23462 eventRV=23471 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Started type=Normal count=1 msg=\"Container started\" ",
    "name=backend-8698b7497d-w4k29 kind=Pod objectAPIversion=v1 objectRV=23462 eventRV=23470 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Created type=Normal count=1 msg=\"Container created\" ",
    "name=backend-8698b7497d-w4k29 kind=Pod objectAPIversion=v1 objectRV=23462 eventRV=23468 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-8698b7497d-w4k29 kind=Pod objectAPIversion=v1 objectRV=23462 eventRV=23476 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Unhealthy type=Warning count=2 msg=\"Readiness probe failed: Get \\\"http://10.244.2.24:9898/readyz\\\": dial tcp 10.244.2.24:9898: connect: connection refused\" ",
    "name=backend-8698b7497d-w4k29 kind=Pod objectAPIversion=v1 objectRV=23462 eventRV=23484 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Created type=Normal count=2 msg=\"Container created\" ",
    "name=backend-8698b7497d-w4k29 kind=Pod objectAPIversion=v1 objectRV=23462 eventRV=23486 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Unhealthy type=Warning count=3 msg=\"Readiness probe failed: Get \\\"http://10.244.2.24:9898/readyz\\\": dial tcp 10.244.2.24:9898: connect: connection refused\" ",
    "name=backend-8698b7497d-w4k29 kind=Pod objectAPIversion=v1 objectRV=23462 eventRV=23485 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Started type=Normal count=2 msg=\"Container started\" ",
    "name=backend-8698b7497d-w4k29 kind=Pod objectAPIversion=v1 objectRV=23462 eventRV=23482 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Pulled type=Normal count=2 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-8698b7497d-w4k29 kind=Pod objectAPIversion=v1 objectRV=23462 eventRV=23492 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Unhealthy type=Warning count=4 msg=\"Readiness probe failed: Get \\\"http://10.244.2.24:9898/readyz\\\": dial tcp 10.244.2.24:9898: connect: connection refused\" ",
    "name=backend-8698b7497d-w4k29 kind=Pod objectAPIversion=v1 objectRV=23462 eventRV=23498 repor
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
      "value": 0.9904651363475191
    }
  ]
}
```

