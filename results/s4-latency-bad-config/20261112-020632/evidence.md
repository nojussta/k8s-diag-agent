# Evidence: LabHighLatency demo/frontend-69bd4c6cb4-xjfk6

- Read-only calls: **22**, collection time: **0.07 s**

## A1 Parse alert and resolve target (Alert payload + K8s API; 2 calls, 0.01 s)

```json
{
  "alertname": "LabHighLatency",
  "namespace": "demo",
  "pod": "frontend-69bd4c6cb4-xjfk6",
  "container": "",
  "deployment": "",
  "service": "frontend",
  "summary": "Service demo/frontend is slow",
  "startsAt": "2026-11-12T00:08:00.907Z"
}
```

## A2 Pod and container status (K8s API: pods; 1 calls, 0 s)

```json
{
  "conditions": {
    "ContainersReady": "True",
    "Initialized": "True",
    "PodReadyToStartContainers": "True",
    "PodScheduled": "True",
    "Ready": "True"
  },
  "containers": [
    {
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "lastState": null,
      "name": "frontend",
      "ready": true,
      "restartCount": 0,
      "state": {
        "startedAt": "2026-11-12T00:06:33Z",
        "state": "running"
      }
    }
  ],
  "node": "kursinis-worker2",
  "phase": "Running",
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
  "2026-11-12T00:06:32Z Normal Scheduled x1: Successfully assigned demo/frontend-69bd4c6cb4-xjfk6 to kursinis-worker2",
  "2026-11-12T00:06:33Z Normal Pulled x1: Container image \"ghcr.io/stefanprodan/podinfo:6.15.0\" already present on machine and can be accessed by the pod",
  "2026-11-12T00:06:33Z Normal Created x1: Container created",
  "2026-11-12T00:06:33Z Normal Started x1: Container started",
  "2026-11-12T00:06:33Z Warning Unhealthy x1: Readiness probe failed: Get \"http://10.244.1.15:9898/readyz\": dial tcp 10.244.1.15:9898: connect: connection refused"
]
```

## A4 Container logs (current and previous) (K8s API: pods/log; 2 calls, 0.01 s)

```json
{
  "current": [
    "{\"level\":\"info\",\"ts\":\"2026-11-12T00:06:33.353Z\",\"caller\":\"podinfo/main.go:170\",\"msg\":\"Starting podinfo\",\"version\":\"6.15.0\",\"revision\":\"dd507173b7b75b2312a36cabe0de5f09c1ce69c8\",\"port\":\"9898\"}",
    "{\"level\":\"info\",\"ts\":\"2026-11-12T00:06:33.353Z\",\"caller\":\"http/server.go:273\",\"msg\":\"Starting HTTP Server.\",\"addr\":\":9898\"}",
    "{\"level\":\"error\",\"ts\":\"2026-11-12T00:06:36.756Z\",\"caller\":\"http/echo.go:65\",\"msg\":\"backend call failed\",\"error\":\"Post \\\"http://backend.demo.svc.cluster.local:9899/echo\\\": context canceled\",\"url\":\"http://backend.demo.svc.cluster.local:9899/echo\",\"stacktrace\":\"github.com/stefanprodan/podinfo/pkg/api/http.(*Server).echoHandler.func1\\n\\t/podinfo/pkg/api/http/echo.go:65\"}",
    "{\"level\":\"error\",\"ts\":\"2026-11-12T00:06:39.763Z\",\"caller\":\"http/echo.go:65\",\"msg\":\"backend call failed\",\"error\":\"Post \\\"http://backend.demo.svc.cluster.local:9899/echo\\\": context canceled\",\"url\":\"http://backend.demo.svc.cluster.local:9899/echo\",\"stacktrace\":\"github.com/stefanprodan/podinfo/pkg/api/http.(*Server).echoHandler.func1\\n\\t/podinfo/pkg/api/http/echo.go:65\"}",
    "{\"level\":\"error\",\"ts\":\"2026-11-12T00:06:42.771Z\",\"caller\":\"http/echo.go:65\",\"msg\":\"backend call failed\",\"error\":\"Post \\\"http://backend.demo.svc.cluster.local:9899/echo\\\": context canceled\",\"url\":\"http://backend.demo.svc.cluster.local:9899/echo\",\"stacktrace\":\"github.com/stefanprodan/podinfo/pkg/api/http.(*Server).echoHandler.func1\\n\\t/podinfo/pkg/api/http/echo.go:65\"}",
    "{\"level\":\"error\",\"ts\":\"2026-11-12T00:06:45.776Z\",\"caller\":\"http/echo.go:65\",\"msg\":\"backend call failed\",\"error\":\"Post \\\"http://backend.demo.svc.cluster.local:9899/echo\\\": context canceled\",\"url\":\"http://backend.demo.svc.cluster.local:9899/echo\",\"stacktrace\":\"github.com/stefanprodan/podinfo/pkg/api/http.(*Server).echoHandler.func1\\n\\t/podinfo/pkg/api/http/echo.go:65\"}",
    "{\"level\":\"error\",\"ts\":\"2026-11-12T00:06:48.784Z\",\"caller\":\"http/echo.go:65\",\"msg\":\"backend call failed\",\"error\":\"Post \\\"http://backend.demo.svc.cluster.local:9899/echo\\\": context canceled\",\"url\":\"http://backend.demo.svc.cluster.local:9899/echo\",\"stacktrace\":\"github.com/stefanprodan/podinfo/pkg/api/http.(*Server).echoHandler.func1\\n\\t/podinfo/pkg/api/http/echo.go:65\"}",
    "{\"level\":\"error\",\"ts\":\"2026-11-12T00:06:51.792Z\",\"caller\":\"http/echo.go:65\",\"msg\":\"backend call failed\",\"error\":\"Post \\\"http://backend.demo.svc.cluster.local:9899/echo\\\": context canceled\",\"url\":\"http://backend.demo.svc.cluster.local:9899/echo\",\"stacktrace\":\"github.com/stefanprodan/podinfo/pkg/api/http.(*Server).echoHandler.func1\\n\\t/podinfo/pkg/api/http/echo.go:65\"}",
    "{\"level\":\"error\",\"ts\":\"2026-11-12T00:06:54.800Z\",\"caller\":\"http/echo.go:65\",\"msg\":\"backend call failed\",\"error\":\"Post \\\"http://backend.demo.svc.cluster.local:9899/echo\\\": context canceled\",\"url\":\"http://backend.demo.svc.cluster.local:9899/echo\",\"stacktrace\":\"github.com/stefanprodan/podinfo/pkg/api/http.(*Server).echoHandler.func1\\n\\t/podinfo/pkg/api/http/echo.go:65\"}",
    "{\"level\":\"error\",\"ts\":\"2026-11-12T00:06:57.806Z\",\"caller\":\"http/echo.go:65\",\"msg\":\"backend call failed\",\"error\":\"Post \\\"http://backend.demo.svc.cluster.local:9899/echo\\\": context canceled\",\"url\":\"http://backend.demo.svc.cluster.local:9899/echo\",\"stacktrace\":\"github.com/stefanprodan/podinfo/pkg/api/http.(*Server).echoHandler.func1\\n\\t/podinfo/pkg/api/http/echo.go:65\"}",
    "{\"level\":\"error\",\"ts\":\"2026-11-12T00:07:00.814Z\",\"caller\":\"http/echo.go:65\",\"msg\":\"backend call failed\",\"error\":\"Post \\\"http://backend.demo.svc.cluster.local:9899/echo\\\": context canceled\",\"url\":\"http://backend.demo.svc.cluster.local:9899/echo\",\"stacktrace\":\"github.com/stefanproda
```

## A5 Owner, rollout history and recent change (K8s API: replicasets, deployments; 3 calls, 0 s)

```json
{
  "conditions": [
    "Available=True (MinimumReplicasAvailable): Deployment has minimum availability.",
    "Progressing=True (NewReplicaSetAvailable): ReplicaSet \"frontend-69bd4c6cb4\" has successfully progressed."
  ],
  "deployment": "frontend",
  "replicas": {
    "availableReplicas": 1,
    "readyReplicas": 1,
    "replicas": 1,
    "unavailableReplicas": 0,
    "updatedReplicas": 1
  },
  "revisions": [
    {
      "change_cause": "release 2.1: frontend configuration update",
      "created": "2026-11-11T23:57:28Z",
      "name": "frontend-69bd4c6cb4",
      "replicas": 1,
      "revision": 4
    },
    {
      "change_cause": "release 2.1: move backend calls to the new API port",
      "created": "2026-11-11T23:56:25Z",
      "name": "frontend-6554997569",
      "replicas": 0,
      "revision": 3
    }
  ],
  "template_diff_latest_vs_previous": [
    {
      "container": "frontend",
      "field": "args",
      "previous": [
        "--port=9898",
        "--port-metrics=9797",
        "--level=info",
        "--backend-url=http://backend.demo.svc.cluster.local:9898/echo"
      ],
      "current": [
        "--port=9898",
        "--port-metrics=9797",
        "--level=info",
        "--backend-url=http://backend.demo.svc.cluster.local:9899/echo"
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
      "pod": "frontend-69bd4c6cb4-xjfk6",
      "service": "kps-kube-state-metrics",
      "uid": "2bde2b53-7d88-4511-9af9-e9e59b6bd9f3",
      "value": 0
    }
  ]
}
```

## A7 Resource usage vs. limits (Prometheus (cAdvisor + kube-state-metrics); 3 calls, 0.01 s)

```json
{
  "cpu_usage_cores_5m": [
    {
      "container": "frontend",
      "cpu": "total",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod2bde2b53_7d88_4511_9af9_e9e59b6bd9f3.slice/cri-containerd-316263ad76f7d1e0b161d162c6fdbba1e574439214d84d859ebd0b2ae1d37016.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "316263ad76f7d1e0b161d162c6fdbba1e574439214d84d859ebd0b2ae1d37016",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "frontend-69bd4c6cb4-xjfk6",
      "service": "kps-kubelet",
      "value": 0.0006349142346549622
    }
  ],
  "memory_limit_bytes": [
    {
      "__name__": "kube_pod_container_resource_limits",
      "container": "frontend",
      "endpoint": "http",
      "instance": "10.244.2.4:8080",
      "job": "kube-state-metrics",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "frontend-69bd4c6cb4-xjfk6",
      "resource": "memory",
      "service": "kps-kube-state-metrics",
      "uid": "2bde2b53-7d88-4511-9af9-e9e59b6bd9f3",
      "unit": "byte",
      "value": 134217728
    }
  ],
  "memory_working_set_max_15m_bytes": [
    {
      "container": "frontend",
      "endpoint": "https-metrics",
      "id": "/kubelet.slice/kubelet-kubepods.slice/kubelet-kubepods-burstable.slice/kubelet-kubepods-burstable-pod2bde2b53_7d88_4511_9af9_e9e59b6bd9f3.slice/cri-containerd-316263ad76f7d1e0b161d162c6fdbba1e574439214d84d859ebd0b2ae1d37016.scope",
      "image": "ghcr.io/stefanprodan/podinfo:6.15.0",
      "instance": "172.18.0.2:10250",
      "job": "kubelet",
      "metrics_path": "/metrics/cadvisor",
      "name": "316263ad76f7d1e0b161d162c6fdbba1e574439214d84d859ebd0b2ae1d37016",
      "namespace": "demo",
      "node": "kursinis-worker2",
      "pod": "frontend-69bd4c6cb4-xjfk6",
      "service": "kps-kubelet",
      "value": 28700672
    }
  ]
}
```

## A8 Service endpoints of the workload (K8s API: services, endpointslices; 2 calls, 0 s)

```json
[
  {
    "endpoints": 1,
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

## A10 Log and event history (Loki) (Loki; 2 calls, 0.02 s)

```json
{
  "events": [
    "name=frontend-69bd4c6cb4-xjfk6 kind=Pod objectAPIversion=v1 objectRV=2790 eventRV=2796 reportingcontroller=default-scheduler sourcecomponent=default-scheduler reason=Scheduled type=Normal count=1 msg=\"Successfully assigned demo/frontend-69bd4c6cb4-xjfk6 to kursinis-worker2\" ",
    "name=frontend-69bd4c6cb4 kind=ReplicaSet objectAPIversion=apps/v1 objectRV=2788 eventRV=2792 reportingcontroller=replicaset-controller sourcecomponent=replicaset-controller reason=SuccessfulCreate type=Normal count=1 msg=\"Created pod: frontend-69bd4c6cb4-xjfk6\" ",
    "name=frontend-69bd4c6cb4-xjfk6 kind=Pod objectAPIversion=v1 objectRV=2794 eventRV=2806 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Unhealthy type=Warning count=1 msg=\"Readiness probe failed: Get \\\"http://10.244.1.15:9898/readyz\\\": dial tcp 10.244.1.15:9898: connect: connection refused\" ",
    "name=frontend-69bd4c6cb4-xjfk6 kind=Pod objectAPIversion=v1 objectRV=2794 eventRV=2805 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Started type=Normal count=1 msg=\"Container started\" ",
    "name=frontend-69bd4c6cb4-xjfk6 kind=Pod objectAPIversion=v1 objectRV=2794 eventRV=2804 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Created type=Normal count=1 msg=\"Container created\" ",
    "name=frontend-69bd4c6cb4-xjfk6 kind=Pod objectAPIversion=v1 objectRV=2794 eventRV=2802 reportinginstance=kursinis-worker2 reportingcontroller=kubelet sourcecomponent=kubelet sourcehost=kursinis-worker2 reason=Pulled type=Normal count=1 msg=\"Container image \\\"ghcr.io/stefanprodan/podinfo:6.15.0\\\" already present on machine and can be accessed by the pod\" "
  ],
  "pod_logs": [
    "{\"level\":\"info\",\"ts\":\"2026-11-12T00:06:33.353Z\",\"caller\":\"podinfo/main.go:170\",\"msg\":\"Starting podinfo\",\"version\":\"6.15.0\",\"revision\":\"dd507173b7b75b2312a36cabe0de5f09c1ce69c8\",\"port\":\"9898\"}\n",
    "{\"level\":\"info\",\"ts\":\"2026-11-12T00:06:33.353Z\",\"caller\":\"http/server.go:273\",\"msg\":\"Starting HTTP Server.\",\"addr\":\":9898\"}\n",
    "{\"level\":\"error\",\"ts\":\"2026-11-12T00:06:36.756Z\",\"caller\":\"http/echo.go:65\",\"msg\":\"backend call failed\",\"error\":\"Post \\\"http://backend.demo.svc.cluster.local:9899/echo\\\": context canceled\",\"url\":\"http://backend.demo.svc.cluster.local:9899/echo\",\"stacktrace\":\"github.com/stefanprodan/podinfo/pkg/api/http.(*Server).echoHandler.func1\\n\\t/podinfo/pkg/api/http/echo.go:65\"}\n",
    "{\"level\":\"error\",\"ts\":\"2026-11-12T00:06:39.763Z\",\"caller\":\"http/echo.go:65\",\"msg\":\"backend call failed\",\"error\":\"Post \\\"http://backend.demo.svc.cluster.local:9899/echo\\\": context canceled\",\"url\":\"http://backend.demo.svc.cluster.local:9899/echo\",\"stacktrace\":\"github.com/stefanprodan/podinfo/pkg/api/http.(*Server).echoHandler.func1\\n\\t/podinfo/pkg/api/http/echo.go:65\"}\n",
    "{\"level\":\"error\",\"ts\":\"2026-11-12T00:06:42.771Z\",\"caller\":\"http/echo.go:65\",\"msg\":\"backend call failed\",\"error\":\"Post \\\"http://backend.demo.svc.cluster.local:9899/echo\\\": context canceled\",\"url\":\"http://backend.demo.svc.cluster.local:9899/echo\",\"stacktrace\":\"github.com/stefanprodan/podinfo/pkg/api/http.(*Server).echoHandler.func1\\n\\t/podinfo/pkg/api/http/echo.go:65\"}\n",
    "{\"level\":\"error\",\"ts\":\"2026-11-12T00:06:45.776Z\",\"caller\":\"http/echo.go:65\",\"msg\":\"backend call failed\",\"error\":\"Post \\\"http://backend.demo.svc.cluster.local:9899/echo\\\": context canceled\",\"url\":\"http://backend.demo.svc.cluster.local:9899/echo\",\"stacktrace\":\"github.com/stefanprodan/podinfo/pkg/api/http.(*Server).echoHandler.func1\\n\\t/podinfo/pkg/api/http/echo.go:65\"}\n",
    "{\"level\":\"error\",\"ts\":\"2026-11-12T00:06:48.784Z\",\"caller\":\"http/ec
```

## A11 Service golden signals (RED) (Prometheus (application metrics); 3 calls, 0.01 s)

```json
{
  "error_rate_5xx_rps": [],
  "latency_p95_s": [
    {
      "value": 1.9625248983800003
    }
  ],
  "request_rate_rps": [
    {
      "value": 0.7687733129077479
    }
  ]
}
```

