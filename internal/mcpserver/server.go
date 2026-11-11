// Package mcpserver exposes the read-only tool layer over the Model Context
// Protocol. Security level 2 of the design lives here: the server registers
// exactly ten tools, all read-only, and there is no generic command tool.
package mcpserver

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nojussta/k8s-diag-agent/internal/tools"
)

type nsPod struct {
	Namespace string `json:"namespace" jsonschema:"Kubernetes namespace"`
	Pod       string `json:"pod" jsonschema:"pod name"`
}

type listPodsIn struct {
	Namespace     string `json:"namespace" jsonschema:"Kubernetes namespace"`
	LabelSelector string `json:"label_selector,omitempty" jsonschema:"optional label selector, e.g. app=backend"`
}

type eventsIn struct {
	Namespace string `json:"namespace" jsonschema:"Kubernetes namespace"`
	Object    string `json:"object" jsonschema:"name of the pod, deployment or other object"`
}

type logsIn struct {
	Namespace string `json:"namespace" jsonschema:"Kubernetes namespace"`
	Pod       string `json:"pod" jsonschema:"pod name"`
	Container string `json:"container,omitempty" jsonschema:"container name; default is the first container"`
	Previous  bool   `json:"previous,omitempty" jsonschema:"read the logs of the previous (crashed) container instance"`
	Tail      int    `json:"tail,omitempty" jsonschema:"number of last lines, at most 50"`
}

type nodeIn struct {
	Node string `json:"node" jsonschema:"node name"`
}

type queryLogsIn struct {
	LogQL   string `json:"logql" jsonschema:"LogQL query, e.g. {namespace=\"demo\"} |= \"error\""`
	Minutes int    `json:"minutes,omitempty" jsonschema:"look-back window, at most 60"`
	Limit   int    `json:"limit,omitempty" jsonschema:"number of lines, at most 50"`
}

type queryMetricsIn struct {
	PromQL string `json:"promql" jsonschema:"instant PromQL query"`
}

// Run is the common adapter: tool function -> MCP result carrying the JSON envelope.
func wrap[In any](f func(context.Context, In) tools.Result) func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
		r := f(ctx, in)
		b, _ := json.Marshal(r)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
			IsError: !r.OK,
		}, nil, nil
	}
}

// New builds the server with the ten tools.
func New(ts *tools.Toolset) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "k8s-diag-tools", Version: "0.1.0"}, nil)
	add := func(name, desc string, h any) {
		switch f := h.(type) {
		case func(context.Context, *mcp.CallToolRequest, nsPod) (*mcp.CallToolResult, any, error):
			mcp.AddTool(s, &mcp.Tool{Name: name, Description: desc}, f)
		case func(context.Context, *mcp.CallToolRequest, listPodsIn) (*mcp.CallToolResult, any, error):
			mcp.AddTool(s, &mcp.Tool{Name: name, Description: desc}, f)
		case func(context.Context, *mcp.CallToolRequest, eventsIn) (*mcp.CallToolResult, any, error):
			mcp.AddTool(s, &mcp.Tool{Name: name, Description: desc}, f)
		case func(context.Context, *mcp.CallToolRequest, logsIn) (*mcp.CallToolResult, any, error):
			mcp.AddTool(s, &mcp.Tool{Name: name, Description: desc}, f)
		case func(context.Context, *mcp.CallToolRequest, nodeIn) (*mcp.CallToolResult, any, error):
			mcp.AddTool(s, &mcp.Tool{Name: name, Description: desc}, f)
		case func(context.Context, *mcp.CallToolRequest, queryLogsIn) (*mcp.CallToolResult, any, error):
			mcp.AddTool(s, &mcp.Tool{Name: name, Description: desc}, f)
		case func(context.Context, *mcp.CallToolRequest, queryMetricsIn) (*mcp.CallToolResult, any, error):
			mcp.AddTool(s, &mcp.Tool{Name: name, Description: desc}, f)
		default:
			panic("mcpserver: unsupported handler for " + name)
		}
	}
	add("get_pod_status", "Phase, conditions, container states, restart counts and resource limits of one pod (A2).",
		wrap(func(ctx context.Context, in nsPod) tools.Result { return ts.GetPodStatus(ctx, in.Namespace, in.Pod) }))
	add("list_pods", "Pods in a namespace with phase and restart count, at most 50 (A1).",
		wrap(func(ctx context.Context, in listPodsIn) tools.Result {
			return ts.ListPods(ctx, in.Namespace, in.LabelSelector)
		}))
	add("get_events", "Latest events of an object, at most 20 (A3).",
		wrap(func(ctx context.Context, in eventsIn) tools.Result { return ts.GetEvents(ctx, in.Namespace, in.Object) }))
	add("get_container_logs", "Last log lines of a container, at most 50 lines and 8 KB (A4).",
		wrap(func(ctx context.Context, in logsIn) tools.Result {
			return ts.GetContainerLogs(ctx, in.Namespace, in.Pod, in.Container, in.Previous, in.Tail)
		}))
	add("get_rollout_history", "Owner workload, latest revisions, change cause and the diff between the two newest pod templates (A5).",
		wrap(func(ctx context.Context, in nsPod) tools.Result {
			return ts.GetRolloutHistory(ctx, in.Namespace, in.Pod)
		}))
	add("get_resource_usage", "Memory and CPU usage, memory limit, restarts and last termination reason of a pod from Prometheus (A6, A7).",
		wrap(func(ctx context.Context, in nsPod) tools.Result {
			return ts.GetResourceUsage(ctx, in.Namespace, in.Pod)
		}))
	add("get_service_endpoints", "Services selecting the pod and how many of their endpoints are ready (A8).",
		wrap(func(ctx context.Context, in nsPod) tools.Result {
			return ts.GetServiceEndpoints(ctx, in.Namespace, in.Pod)
		}))
	add("get_node_status", "Node conditions and allocatable resources (A9).",
		wrap(func(ctx context.Context, in nodeIn) tools.Result { return ts.GetNodeStatus(ctx, in.Node) }))
	add("query_logs", "LogQL query against Loki, at most 60 minutes and 50 lines (A10).",
		wrap(func(ctx context.Context, in queryLogsIn) tools.Result {
			return ts.QueryLogs(ctx, in.LogQL, in.Minutes, in.Limit)
		}))
	add("query_metrics", "Instant PromQL query against Prometheus, at most 20 rows (A11 and others).",
		wrap(func(ctx context.Context, in queryMetricsIn) tools.Result { return ts.QueryMetrics(ctx, in.PromQL) }))
	return s
}
