# Evidence: LabPodNotReady demo/frontend-6df678c5fd-f6cpj

- Read-only calls: **19**, collection time: **0.05 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 0 calls, 0 s)

```json
{
  "alertname": "LabPodNotReady",
  "namespace": "demo",
  "pod": "frontend-6df678c5fd-f6cpj",
  "container": "",
  "deployment": "",
  "summary": "Pod demo/frontend-6df678c5fd-f6cpj is not ready",
  "startsAt": "2026-11-12T03:10:00.907Z"
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
      "lastState": null,
      "name": "frontend",
      "ready": false,
      "restartCount": 0,
      "state": {
        "message": "configmap \"frontend-config\" not found",
        "reason": "CreateContainerConfigError",
        "state": "waiting"
      }
    }
  ],
  "node": "kursinis-worker2",
  "phase": "Pending",
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
  "2026-11-12T03:07:35Z Normal Scheduled x1: Successfully assigned demo/frontend-6df678c5fd-f6cpj to kursinis-worker2",
  "2026-11-12T03:09:54Z Normal Pulled x12: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-12T03:09:54Z Warning Failed x12: Error: configmap \"frontend-config\" not found"
]
```

## A4 Container logs (current and previous) (K8s API: pods/log; 1 calls, 0 s)

```json
"ERROR: the server rejected our request for an unknown reason (get pods frontend-6df678c5fd-f6cpj)"
```

## A5 Owner, rollout history and recent change (K8s API: replicasets, deployments; 3 calls, 0.01 s)

```json
{
  "conditions": [
    "Available=True (MinimumReplicasAvailable): Deployment has minimum availability.",
    "Progressing=True (ReplicaSetUpdated): ReplicaSet \"frontend-6df678c5fd\" is progressing."
  ],
  "deployment": "frontend",
  "replicas": {
    "availableReplicas": 1,
    "readyReplicas": 1,
    "replicas": 2,
    "unavailableReplicas": 1,
    "updatedReplicas": 1
  },
  "revisions": [
    {
      "change_cause": "release 2.1: frontend configuration update",
      "created": "2026-11-12T01:13:24Z",
      "name": "frontend-6df678c5fd",
      "replicas": 1,
      "revision": 56
    },
    {
      "change_cause": "release 2.1: frontend configuration update",
      "created": "2026-11-11T23:56:25Z",
      "name": "frontend-6554997569",
      "replicas": 1,
      "revision": 55
    },
    {
      "change_cause": null,
      "created": "2026-11-12T00:58:23Z",
      "name": "frontend-68747558f7",
      "replicas": 0,
      "revision": 50
    }
  ],
  "template_diff_latest_vs_previous": [
    {
      "container": "frontend",
      "field": "env",
      "previous": [
        {
          "name": "PODINFO_UI_MESSAGE",
          "value": "frontend"
        }
      ],
      "current": [
        {
          "name": "PODINFO_UI_MESSAGE",
          "value": "frontend"
        },
        {
          "name": "PODINFO_UI_COLOR",
          "valueFrom": {
            "configMapKeyRef": {
              "name": "frontend-config",
              "key": "ui.color"
            }
          }
        }
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
      "pod": "frontend-6df678c5fd-f6cpj",
      "service": "kps-kube-state-metrics",
      "uid": "55204655-f620-4a01-830d-1b3079a9b15a",
      "value": 0
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
      "container": "frontend",
      "endpoint": "http",
      "instance": "10.244.2.4:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "frontend-6df678c5fd-f6cpj",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "55204655-f620-4a01-830d-1b3079a9b15a",
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
    "endpoints": 2,
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

## A10 Log and event history (Loki) (Loki; 2 calls, 0.01 s)

```json
{
  "events": [
    "name=frontend-6df678c5fd-f6cpj kind=Pod objectAPIversion=v1 objectRV=23163 eventRV=23170 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=1 msg=\"Error: configmap \\\"frontend-config\\\" not found\" ",
    "name=frontend-6df678c5fd-f6cpj kind=Pod objectAPIversion=v1 objectRV=23160 eventRV=23166 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/frontend-6df678c5fd-f6cpj to kursinis-worker2\" ",
    "name=frontend-6df678c5fd kind=ReplicaSet objectAPIversion=apps/v1 objectRV=23159 eventRV=23164 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: frontend-6df678c5fd-f6cpj\" ",
    "name=frontend-6df678c5fd-f6cpj kind=Pod objectAPIversion=v1 objectRV=23163 eventRV=23169 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=frontend-6df678c5fd-f6cpj kind=Pod objectAPIversion=v1 objectRV=23163 eventRV=23176 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=2 msg=\"Error: configmap \\\"frontend-config\\\" not found\" ",
    "name=frontend-6df678c5fd-f6cpj kind=Pod objectAPIversion=v1 objectRV=23163 eventRV=23174 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=2 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=frontend-6df678c5fd-f6cpj kind=Pod objectAPIversion=v1 objectRV=23163 eventRV=23202 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=3 msg=\"Error: configmap \\\"frontend-config\\\" not found\" ",
    "name=frontend-6df678c5fd-f6cpj kind=Pod objectAPIversion=v1 objectRV=23163 eventRV=23201 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=3 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=frontend-6df678c5fd-f6cpj kind=Pod objectAPIversion=v1 objectRV=23163 eventRV=23228 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=4 msg=\"Error: configmap \\\"frontend-config\\\" not found\" ",
    "name=frontend-6df678c5fd-f6cpj kind=Pod objectAPIversion=v1 objectRV=23163 eventRV=23227 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=4 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=frontend-6df678c5fd-f6cpj kind=Pod objectAPIversion=v1 objectRV=23163 eventRV=23253 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=5 msg=\"Error: configmap \\\"frontend-config\\\" not found\" ",
    "name=frontend-6df678c5fd-f6cpj kind=Pod objectAPIversion=v1 objectRV=23163 eventRV=23252 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=5 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=frontend-6df678c5fd-f6c
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
      "value": 0.9929685197400738
    }
  ]
}
```

