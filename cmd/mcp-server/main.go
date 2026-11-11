// Command mcp-server serves the ten read-only diagnostic tools over MCP (stdio).
// It uses the lab's isolated kubeconfig and reaches Prometheus and Loki through
// the API server proxy, so it needs nothing but read access to the cluster.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nojussta/k8s-diag-agent/internal/lab"
	"github.com/nojussta/k8s-diag-agent/internal/mcpserver"
	"github.com/nojussta/k8s-diag-agent/internal/tools"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	l, err := lab.New()
	if err != nil {
		log.Fatal(err)
	}
	srv := mcpserver.New(&tools.Toolset{CS: l.Cs, Obs: l})
	if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
