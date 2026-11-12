# Evidence: LabPodCrashLooping demo/backend-7b8d9f4db-4grm2

- Read-only calls: **20**, collection time: **0.06 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 0 calls, 0 s)

```json
{
  "alertname": "LabPodCrashLooping",
  "namespace": "demo",
  "pod": "backend-7b8d9f4db-4grm2",
  "container": "backend",
  "deployment": "",
  "summary": "Pod demo/backend-7b8d9f4db-4grm2 (backend) is crash looping",
  "startsAt": "2026-11-12T00:47:45.907Z"
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
        "finishedAt": "2026-11-12T00:47:34Z",
        "reason": "OOMKilled",
        "startedAt": "2026-11-12T00:47:29Z",
        "state": "terminated"
      },
      "name": "backend",
      "ready": false,
      "restartCount": 4,
      "state": {
        "message": "back-off 1m20s restarting failed container=backend pod=backend-7b8d9f4db-4grm2_demo(a15d81e6-47eb-4ee3-9955-7c952715bf41)",
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

## A3 Events for the pod (K8s API: events; 1 calls, 0.01 s)

```json
[
  "2026-11-12T00:46:01Z Normal Scheduled x1: Successfully assigned demo/backend-7b8d9f4db-4grm2 to kursinis-worker2",
  "2026-11-12T00:47:29Z Normal Pulled x5: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-12T00:47:29Z Normal Created x5: Container created",
  "2026-11-12T00:47:29Z Normal Started x5: Container started",
  "2026-11-12T00:47:32Z Warning Unhealthy x1: Liveness probe failed: Get \"http://10.244.1.28:9898/healthz\": dial tcp 10.244.1.28:9898: connect: connection refused",
  "2026-11-12T00:47:35Z Warning Unhealthy x11: Readiness probe failed: Get \"http://10.244.1.28:9898/readyz\": dial tcp 10.244.1.28:9898: connect: connection refused",
  "2026-11-12T00:47:42Z Warning BackOff x8: Back-off restarting failed container backend in pod backend-7b8d9f4db-4grm2_demo(a15d81e6-47eb-4ee3-9955-7c952715bf41)"
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
      "created": "2026-11-12T00:40:53Z",
      "name": "backend-7b8d9f4db",
      "replicas": 1,
      "revision": 4
    },
    {
      "change_cause": null,
      "created": "2026-11-11T23:56:25Z",
      "name": "backend-5b97dd7dcb",
      "replicas": 2,
      "revision": 3
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
      "pod": "backend-7b8d9f4db-4grm2",
      "reason": "OOMKilled",
      "service": "kps-kube-state-metrics",
      "uid": "a15d81e6-47eb-4ee3-9955-7c952715bf41",
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
      "pod": "backend-7b8d9f4db-4grm2",
      "service": "kps-kube-state-metrics",
      "uid": "a15d81e6-47eb-4ee3-9955-7c952715bf41",
      "value": 3.604239712971391
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
      "pod": "backend-7b8d9f4db-4grm2",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "a15d81e6-47eb-4ee3-9955-7c952715bf41",
      "unit": "byte",
      "value": 134217728
    }
  ],
  "memory_working_set_max_15m_bytes": [
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-poda15d81e6_47eb_4ee3_9955_7c952715bf41.slice/cri-containerd-332cee0dd33e63f944317c046b03b8e7d968fc6cbd92cea1b664c14603037e7b.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "332cee0dd33e63f944317c046b03b8e7d968fc6cbd92cea1b664c14603037e7b",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "backend-7b8d9f4db-4grm2",
      "service": "kps-kubelet",
      "value": 133705728
    },
    {
      "container": "backend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-poda15d81e6_47eb_4ee3_9955_7c952715bf41.slice/cri-containerd-0bdfc0627124d97f9f972094942ba128e3b7def0a733827e243aca89bc42eb4f.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "0bdfc0627124d97f9f972094942ba128e3b7def0a733827e243aca89bc42eb4f",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "backend-7b8d9f4db-4grm2",
      "service": "kps-kubelet",
      "value": 133185536
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
    "name=backend-7b8d9f4db-4grm2 kind=Pod objectAPIversion=v1 objectRV=7243 eventRV=7248 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/backend-7b8d9f4db-4grm2 to kursinis-worker2\" ",
    "name=backend-7b8d9f4db kind=ReplicaSet objectAPIversion=apps/v1 objectRV=7241 eventRV=7245 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: backend-7b8d9f4db-4grm2\" ",
    "name=backend-7b8d9f4db-4grm2 kind=Pod objectAPIversion=v1 objectRV=7247 eventRV=7259 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=1 msg=\"Readiness probe failed: Get \\\"http://10.244.1.28:9898/readyz\\\": dial tcp 10.244.1.28:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-4grm2 kind=Pod objectAPIversion=v1 objectRV=7247 eventRV=7258 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Started type=Normal count=1 msg=\"Container started\" ",
    "name=backend-7b8d9f4db-4grm2 kind=Pod objectAPIversion=v1 objectRV=7247 eventRV=7257 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Created type=Normal count=1 msg=\"Container created\" ",
    "name=backend-7b8d9f4db-4grm2 kind=Pod objectAPIversion=v1 objectRV=7247 eventRV=7255 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-7b8d9f4db-4grm2 kind=Pod objectAPIversion=v1 objectRV=7247 eventRV=7269 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=2 msg=\"Readiness probe failed: Get \\\"http://10.244.1.28:9898/readyz\\\": dial tcp 10.244.1.28:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-4grm2 kind=Pod objectAPIversion=v1 objectRV=7247 eventRV=7268 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Started type=Normal count=2 msg=\"Container started\" ",
    "name=backend-7b8d9f4db-4grm2 kind=Pod objectAPIversion=v1 objectRV=7247 eventRV=7267 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Created type=Normal count=2 msg=\"Container created\" ",
    "name=backend-7b8d9f4db-4grm2 kind=Pod objectAPIversion=v1 objectRV=7247 eventRV=7265 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=2 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=backend-7b8d9f4db-4grm2 kind=Pod objectAPIversion=v1 objectRV=7247 eventRV=7271 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=3 msg=\"Readiness probe failed: Get \\\"http://10.244.1.28:9898/readyz\\\": dial tcp 10.244.1.28:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-4grm2 kind=Pod objectAPIversion=v1 objectRV=7247 eventRV=7274 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=4 msg=\"Readiness probe failed: Get \\\"http://10.244.1.28:9898/readyz\\\": dial tcp 10.244.1.28:9898: connect: connection refused\" ",
    "name=backend-7b8d9f4db-4grm2 kind=Pod objectAPIversion=v1 objectRV=7247 eventRV=7275 reportinginstance=kursini
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
      "value": 0.9930103301145294
    }
  ]
}
```

