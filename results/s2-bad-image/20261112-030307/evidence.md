# Evidence: LabPodNotReady demo/frontend-68747558f7-qr4m7

- Read-only calls: **19**, collection time: **0.05 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 0 calls, 0 s)

```json
{
  "alertname": "LabPodNotReady",
  "namespace": "demo",
  "pod": "frontend-68747558f7-qr4m7",
  "container": "",
  "deployment": "",
  "summary": "Pod demo/frontend-68747558f7-qr4m7 is not ready",
  "startsAt": "2026-11-12T01:05:30.907Z"
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
      "image": "ghcr.io/stefanprodan/podinfo:6.15.99",
      "lastState": null,
      "name": "frontend",
      "ready": false,
      "restartCount": 0,
      "state": {
        "message": "Back-off pulling image \"ghcr.io/stefanprodan/podinfo:6.15.99\": ErrImagePull: rpc error: code = NotFound desc = failed to pull and unpack image \"ghcr.io/stefanprodan/podinfo:6.15.99\": failed to resolve reference \"ghcr.io/stefanprodan/podinfo:6.15.99\": ghcr.io/stefanprodan/podinfo:6.15.99: not found",
        "reason": "ImagePullBackOff",
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
  "2026-11-12T01:03:07Z Normal Scheduled x1: Successfully assigned demo/frontend-68747558f7-qr4m7 to kursinis-worker2",
  "2026-11-12T01:04:32Z Normal Pulling x4: Pulling image \"ghcr.io/stefanprodan/podinfo:6.15.99\"",
  "2026-11-12T01:04:33Z Warning Failed x4: Failed to pull image \"ghcr.io/stefanprodan/podinfo:6.15.99\": rpc error: code = NotFound desc = failed to pull and unpack image \"ghcr.io/stefanprodan/podinfo:6.15.99\": failed to resolve reference \"ghcr.io/stefanprodan/podinfo:6.15.99\": ghcr.io/stefanprodan/podinfo:6.15.99: not found",
  "2026-11-12T01:04:33Z Warning Failed x4: Error: ErrImagePull",
  "2026-11-12T01:05:38Z Normal BackOff x9: Back-off pulling image \"ghcr.io/stefanprodan/podinfo:6.15.99\"",
  "2026-11-12T01:05:38Z Warning Failed x9: Error: ImagePullBackOff"
]
```

## A4 Container logs (current and previous) (K8s API: pods/log; 1 calls, 0 s)

```json
"ERROR: the server rejected our request for an unknown reason (get pods frontend-68747558f7-qr4m7)"
```

## A5 Owner, rollout history and recent change (K8s API: replicasets, deployments; 3 calls, 0.01 s)

```json
{
  "conditions": [
    "Available=True (MinimumReplicasAvailable): Deployment has minimum availability.",
    "Progressing=True (ReplicaSetUpdated): ReplicaSet \"frontend-68747558f7\" is progressing."
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
      "change_cause": null,
      "created": "2026-11-12T00:58:23Z",
      "name": "frontend-68747558f7",
      "replicas": 1,
      "revision": 18
    },
    {
      "change_cause": "release 2.1: frontend configuration update",
      "created": "2026-11-11T23:56:25Z",
      "name": "frontend-6554997569",
      "replicas": 1,
      "revision": 17
    },
    {
      "change_cause": "release 2.1: frontend configuration update",
      "created": "2026-11-11T23:57:28Z",
      "name": "frontend-69bd4c6cb4",
      "replicas": 0,
      "revision": 14
    }
  ],
  "template_diff_latest_vs_previous": [
    {
      "container": "frontend",
      "field": "image",
      "previous": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "current": "ghcr.io/stefanprodan/podinfo:6.15.99"
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
      "pod": "frontend-68747558f7-qr4m7",
      "service": "kps-kube-state-metrics",
      "uid": "9d089e59-1d92-474a-9970-db1ee8f7ed06",
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
      "pod": "frontend-68747558f7-qr4m7",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "9d089e59-1d92-474a-9970-db1ee8f7ed06",
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
    "name=frontend-68747558f7-qr4m7 kind=Pod objectAPIversion=v1 objectRV=9183 eventRV=9189 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/frontend-68747558f7-qr4m7 to kursinis-worker2\" ",
    "name=frontend-68747558f7 kind=ReplicaSet objectAPIversion=apps/v1 objectRV=9181 eventRV=9185 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: frontend-68747558f7-qr4m7\" ",
    "name=frontend-68747558f7-qr4m7 kind=Pod objectAPIversion=v1 objectRV=9187 eventRV=9192 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulling type=Normal count=1 msg=\"Pulling image \\\"ghcr.io/stefanprodan/podinfo:6.15.99\\\"\" ",
    "name=frontend-68747558f7-qr4m7 kind=Pod objectAPIversion=v1 objectRV=9187 eventRV=9198 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=1 msg=\"Error: ImagePullBackOff\" ",
    "name=frontend-68747558f7-qr4m7 kind=Pod objectAPIversion=v1 objectRV=9187 eventRV=9195 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=1 msg=\"Error: ErrImagePull\" ",
    "name=frontend-68747558f7-qr4m7 kind=Pod objectAPIversion=v1 objectRV=9187 eventRV=9196 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=BackOff type=Normal count=1 msg=\"Back-off pulling image \\\"ghcr.io/stefanprodan/podinfo:6.15.99\\\"\" ",
    "name=frontend-68747558f7-qr4m7 kind=Pod objectAPIversion=v1 objectRV=9187 eventRV=9194 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=1 msg=\"Failed to pull image \\\"ghcr.io/stefanprodan/podinfo:6.15.99\\\": rpc error: code = NotFound desc = failed to pull and unpack image \\\"ghcr.io/stefanprodan/podinfo:6.15.99\\\": failed to resolve reference \\\"ghcr.io/stefanprodan/podinfo:6.15.99\\\": ghcr.io/stefanprodan/podinfo:6.15.99: not found\" ",
    "name=frontend-68747558f7-qr4m7 kind=Pod objectAPIversion=v1 objectRV=9187 eventRV=9221 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulling type=Normal count=2 msg=\"Pulling image \\\"ghcr.io/stefanprodan/podinfo:6.15.99\\\"\" ",
    "name=frontend-68747558f7-qr4m7 kind=Pod objectAPIversion=v1 objectRV=9187 eventRV=9224 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=2 msg=\"Error: ErrImagePull\" ",
    "name=frontend-68747558f7-qr4m7 kind=Pod objectAPIversion=v1 objectRV=9187 eventRV=9223 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=2 msg=\"Failed to pull image \\\"ghcr.io/stefanprodan/podinfo:6.15.99\\\": rpc error: code = NotFound desc = failed to pull and unpack image \\\"ghcr.io/stefanprodan/podinfo:6.15.99\\\": failed to resolve reference \\\"ghcr.io/stefanprodan/podinfo:6.15.99\\\": ghcr.io/stefanprodan/podinfo:6.15.99: not found\" ",
    "name=frontend-68747558f7-qr4m7 kind=Pod objectAPIversion=v1 objectRV=9187 eventRV=9254 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Failed type=Warning count=2 msg=\"Error: ImagePullBackOff\" ",
    "name=frontend-68747558f7-qr4m7 kind=Pod objectAPIversion=v1 objectRV=9187 eventRV=9252 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=BackOff type=Normal count=2 msg=\"Back-off pulling image \\\
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
      "value": 0.9929963929318307
    }
  ]
}
```

