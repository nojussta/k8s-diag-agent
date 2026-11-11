# Evidence: LabPodCrashLooping demo/backend-7b8d9f4db-vwx5r

- Read-only calls: **20**, collection time: **0.09 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 0 calls, 0 s)

```json
{
  "alertname": "LabPodCrashLooping",
  "namespace": "demo",
  "pod": "backend-7b8d9f4db-vwx5r",
  "container": "backend",
  "deployment": "",
  "summary": "Pod demo/backend-7b8d9f4db-vwx5r (backend) is crash looping",
  "startsAt": "2026-11-11T08:25:12.348Z"
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
        "finishedAt": "2026-11-11T08:24:40Z",
        "reason": "OOMKilled",
        "startedAt": "2026-11-11T08:24:37Z",
        "state": "terminated"
      },
      "name": "backend",
      "ready": false,
      "restartCount": 4,
      "state": {
        "startedAt": "2026-11-11T08:25:23Z",
        "state": "running"
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
  "2026-11-11T08:23:51Z Normal Scheduled x1: Successfully assigned demo/backend-7b8d9f4db-vwx5r to kursinis-worker",
  "2026-11-11T08:24:11Z Warning Unhealthy x1: Liveness probe failed: Get \"http://10.244.2.12:9898/healthz\": dial tcp 10.244.2.12:9898: connect: connection refused",
  "2026-11-11T08:24:41Z Warning BackOff x6: Back-off restarting failed container backend in pod backend-7b8d9f4db-vwx5r_demo(bab68ccf-233e-441a-a4ff-0a35e33eb807)",
  "2026-11-11T08:25:23Z Normal Pulled x5: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-11T08:25:23Z Normal Created x5: Container created",
  "2026-11-11T08:25:23Z Normal Started x5: Container started",
  "2026-11-11T08:25:24Z Warning Unhealthy x11: Readiness probe failed: Get \"http://10.244.2.12:9898/readyz\": dial tcp 10.244.2.12:9898: connect: connection refused"
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
      "revision": 6
    },
    {
      "change_cause": null,
      "created": "2026-11-11T07:52:36Z",
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
      "pod": "backend-7b8d9f4db-vwx5r",
      "reason": "OOMKilled",
      "service": "kps-kube-state-metrics",
      "uid": "bab68ccf-233e-441a-a4ff-0a35e33eb807",
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
      "pod": "backend-7b8d9f4db-vwx5r",
      "service": "kps-kube-state-metrics",
      "uid": "bab68ccf-233e-441a-a4ff-0a35e33eb807",
      "value": 2.366264449807336
    }
  ]
}
```

## A7 Resource usage vs. limits (Prometheus (cAdvisor + kube-state-metrics); 3 calls, 0.02 s)

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
      "pod": "backend-7b8d9f4db-vwx5r",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "bab68ccf-233e-441a-a4ff-0a35e33eb807",
      "unit": "byte",
      "value": 134217728
    }
  ],
  "memory_working_set_max_15m_bytes": [
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-podbab68ccf_233e_441a_a4ff_0a35e33eb807.slice/cri-containerd-bf0173eeb0f14dfba5fb1f52a2dc51ce750a00a4209fa227653edccb9e01b84c.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "bf0173eeb0f14dfba5fb1f52a2dc51ce750a00a4209fa227653edccb9e01b84c",
      "namespace": "demo",
      "node": "kursinis-worker",
      "pod": "backend-7b8d9f4db-vwx5r",
      "service": "kps-kubelet",
      "value": 125886464
    },
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-podbab68ccf_233e_441a_a4ff_0a35e33eb807.slice/cri-containerd-88d87f41b8e27c0d033369b20dae351bcaf135e325258d2e6595f1c4d7b87793.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "88d87f41b8e27c0d033369b20dae351bcaf135e325258d2e6595f1c4d7b87793",
      "namespace": "demo",
      "node": "kursinis-worker",
      "pod": "backend-7b8d9f4db-vwx5r",
      "service": "kps-kubelet",
      "value": 125231104
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
    "name=backend-7b8d9f4db-vwx5r kind=Pod objectAPIversion=v1 objectRV=4819 eventRV=4830 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Unhealthy type=Warning count=1 msg=\"Readiness probe failed: Get \\\"http://10.244.2.12:9898/readyz\\\": dial tcp 10.244.2.12:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-vwx5r kind=Pod objectAPIversion=v1 objectRV=4819 eventRV=4829 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Started type=Normal count=1 msg=\"Container started\" ",
    "name=backend-7b8d9f4db-vwx5r kind=Pod objectAPIversion=v1 objectRV=4819 eventRV=4828 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Created type=Normal count=1 msg=\"Container created\" ",
    "name=backend-7b8d9f4db-vwx5r kind=Pod objectAPIversion=v1 objectRV=4815 eventRV=4820 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/backend-7b8d9f4db-vwx5r to kursinis-worker\" ",
    "name=backend-7b8d9f4db kind=ReplicaSet objectAPIversion=apps/v1 objectRV=4813 eventRV=4817 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: backend-7b8d9f4db-vwx5r\" ",
    "name=backend-7b8d9f4db-vwx5r kind=Pod objectAPIversion=v1 objectRV=4819 eventRV=4826 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-7b8d9f4db-vwx5r kind=Pod objectAPIversion=v1 objectRV=4819 eventRV=4837 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Unhealthy type=Warning count=2 msg=\"Readiness probe failed: Get \\\"http://10.244.2.12:9898/readyz\\\": dial tcp 10.244.2.12:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-vwx5r kind=Pod objectAPIversion=v1 objectRV=4819 eventRV=4847 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Unhealthy type=Warning count=3 msg=\"Readiness probe failed: Get \\\"http://10.244.2.12:9898/readyz\\\": dial tcp 10.244.2.12:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-vwx5r kind=Pod objectAPIversion=v1 objectRV=4819 eventRV=4846 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Started type=Normal count=2 msg=\"Container started\" ",
    "name=backend-7b8d9f4db-vwx5r kind=Pod objectAPIversion=v1 objectRV=4819 eventRV=4845 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Created type=Normal count=2 msg=\"Container created\" ",
    "name=backend-7b8d9f4db-vwx5r kind=Pod objectAPIversion=v1 objectRV=4819 eventRV=4843 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Pulled type=Normal count=2 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-7b8d9f4db-vwx5r kind=Pod objectAPIversion=v1 objectRV=4819 eventRV=4851 reportinginstance=kursinis-worker reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker reason=Unhealthy type=Warning count=4 msg=\"Readiness probe failed: Get \\\"http://10.244.2.12:9898/readyz\\\": dial tcp 10.244.2.12:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-vwx5r kind=Pod objectAPIversion=v1 objectRV=4819 eventRV=4854 reportinginstance=kursinis-worker reportingcon
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
      "value": 0.9929754878913131
    }
  ]
}
```

