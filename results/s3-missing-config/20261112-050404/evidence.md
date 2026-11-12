# Evidence: LabPodNotReady demo/frontend-6df678c5fd-df5hj

- Read-only calls: **19**, collection time: **0.05 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 0 calls, 0 s)

```json
{
  "alertname": "LabPodNotReady",
  "namespace": "demo",
  "pod": "frontend-6df678c5fd-df5hj",
  "container": "",
  "deployment": "",
  "summary": "Pod demo/frontend-6df678c5fd-df5hj is not ready",
  "startsAt": "2026-11-12T03:06:30.907Z"
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
  "2026-11-12T03:04:04Z Normal Scheduled x1: Successfully assigned demo/frontend-6df678c5fd-df5hj to kursinis-worker2",
  "2026-11-12T03:06:26Z Normal Pulled x13: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-12T03:06:26Z Warning Failed x13: Error: configmap \"frontend-config\" not found"
]
```

## A4 Container logs (current and previous) (K8s API: pods/log; 1 calls, 0 s)

```json
"ERROR: the server rejected our request for an unknown reason (get pods frontend-6df678c5fd-df5hj)"
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
      "revision": 54
    },
    {
      "change_cause": "release 2.1: frontend configuration update",
      "created": "2026-11-11T23:56:25Z",
      "name": "frontend-6554997569",
      "replicas": 1,
      "revision": 53
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
      "pod": "frontend-6df678c5fd-df5hj",
      "service": "kps-kube-state-metrics",
      "uid": "d0b7ff33-3f4e-48ba-8d42-f573c10a4b65",
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
      "pod": "frontend-6df678c5fd-df5hj",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "d0b7ff33-3f4e-48ba-8d42-f573c10a4b65",
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
    "name=frontend-6df678c5fd-df5hj kind=Pod objectAPIversion=v1 objectRV=22751 eventRV=22756 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/frontend-6df678c5fd-df5hj to kursinis-worker2\" ",
    "name=frontend-6df678c5fd kind=ReplicaSet objectAPIversion=apps/v1 objectRV=22750 eventRV=22753 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: frontend-6df678c5fd-df5hj\" ",
    "name=frontend-6df678c5fd-df5hj kind=Pod objectAPIversion=v1 objectRV=22755 eventRV=22768 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=2 msg=\"Error: configmap \\\"frontend-config\\\" not found\" ",
    "name=frontend-6df678c5fd-df5hj kind=Pod objectAPIversion=v1 objectRV=22755 eventRV=22764 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=1 msg=\"Error: configmap \\\"frontend-config\\\" not found\" ",
    "name=frontend-6df678c5fd-df5hj kind=Pod objectAPIversion=v1 objectRV=22755 eventRV=22766 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=2 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=frontend-6df678c5fd-df5hj kind=Pod objectAPIversion=v1 objectRV=22755 eventRV=22763 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=frontend-6df678c5fd-df5hj kind=Pod objectAPIversion=v1 objectRV=22755 eventRV=22795 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=3 msg=\"Error: configmap \\\"frontend-config\\\" not found\" ",
    "name=frontend-6df678c5fd-df5hj kind=Pod objectAPIversion=v1 objectRV=22755 eventRV=22794 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=3 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=frontend-6df678c5fd-df5hj kind=Pod objectAPIversion=v1 objectRV=22755 eventRV=22814 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=4 msg=\"Error: configmap \\\"frontend-config\\\" not found\" ",
    "name=frontend-6df678c5fd-df5hj kind=Pod objectAPIversion=v1 objectRV=22755 eventRV=22813 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=4 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=frontend-6df678c5fd-df5hj kind=Pod objectAPIversion=v1 objectRV=22755 eventRV=22841 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=5 msg=\"Error: configmap \\\"frontend-config\\\" not found\" ",
    "name=frontend-6df678c5fd-df5hj kind=Pod objectAPIversion=v1 objectRV=22755 eventRV=22840 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=5 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" ",
    "name=frontend-6df678c5fd-df5
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
      "value": 0.9893348301992703
    }
  ]
}
```

