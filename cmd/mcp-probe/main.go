// Command mcp-probe is a tiny MCP client: it starts the mcp-server binary over
// stdio, lists the tools and walks a diagnosis of the unhealthiest pod in a
// namespace using only MCP calls. It is the live proof that the tool layer
// returns the decisive evidence (it is not the agent: no model is involved).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	server := flag.String("server", "build/mcp-server", "path to the mcp-server binary")
	ns := flag.String("ns", "demo", "namespace")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	sess, err := mcp.NewClient(&mcp.Implementation{Name: "mcp-probe", Version: "0.1.0"}, nil).
		Connect(ctx, &mcp.CommandTransport{Command: exec.Command(*server)}, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect:", err)
		os.Exit(1)
	}
	defer sess.Close()

	list, err := sess.ListTools(ctx, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	names := []string{}
	for _, t := range list.Tools {
		names = append(names, t.Name)
	}
	fmt.Printf("tools (%d): %s\n\n", len(names), strings.Join(names, ", "))

	calls := 0
	start := time.Now()
	do := func(name string, args map[string]any) map[string]any {
		calls++
		t0 := time.Now()
		res, err := sess.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
		if err != nil {
			fmt.Fprintln(os.Stderr, name, err)
			os.Exit(1)
		}
		var env map[string]any
		_ = json.Unmarshal([]byte(res.Content[0].(*mcp.TextContent).Text), &env)
		b, _ := json.Marshal(env["data"])
		fmt.Printf("[t%d] %-22s ok=%v truncated=%v %4d B  %4.0f ms\n", calls, name, env["ok"], env["truncated"], len(b), time.Since(t0).Seconds()*1000)
		return env
	}

	pods := do("list_pods", map[string]any{"namespace": *ns})
	worst, most := "", -1.0
	for _, p := range pods["data"].([]any) {
		m := p.(map[string]any)
		if r := m["restarts"].(float64); r > most {
			worst, most = m["name"].(string), r
		}
	}
	fmt.Printf("\nunhealthiest pod: %s (restarts=%.0f)\n\n", worst, most)
	args := map[string]any{"namespace": *ns, "pod": worst}
	st := do("get_pod_status", args)
	ro := do("get_rollout_history", args)
	do("get_events", map[string]any{"namespace": *ns, "object": worst})
	do("get_resource_usage", args)
	do("get_container_logs", map[string]any{"namespace": *ns, "pod": worst, "previous": true})

	fmt.Printf("\n%d MCP calls, %.2f s\n\n", calls, time.Since(start).Seconds())
	pretty := func(label string, v any) {
		b, _ := json.MarshalIndent(v, "", "  ")
		fmt.Printf("%s\n%s\n", label, b)
	}
	if d, ok := st["data"].(map[string]any); ok {
		for _, c := range d["containers"].([]any) {
			m := c.(map[string]any)
			pretty("get_pod_status -> "+m["name"].(string)+".lastState", m["lastState"])
		}
	}
	if d, ok := ro["data"].(map[string]any); ok {
		pretty("get_rollout_history -> template diff", d["template_diff_latest_vs_previous"])
	}
}
