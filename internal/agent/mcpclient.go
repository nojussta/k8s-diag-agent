package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPTools is the orchestrator's MCP client (component 4 of the design).
type MCPTools struct{ sess *mcp.ClientSession }

// Connect starts the MCP server command over stdio.
func Connect(ctx context.Context, server *exec.Cmd) (*MCPTools, error) {
	return ConnectTransport(ctx, &mcp.CommandTransport{Command: server})
}

// ConnectTransport is Connect with any transport (in-memory in tests).
func ConnectTransport(ctx context.Context, t mcp.Transport) (*MCPTools, error) {
	s, err := mcp.NewClient(&mcp.Implementation{Name: "k8s-diag-agent", Version: "0.1.0"}, nil).Connect(ctx, t, nil)
	if err != nil {
		return nil, fmt.Errorf("mcp connect: %w", err)
	}
	return &MCPTools{sess: s}, nil
}

// Close ends the session (and the server process).
func (m *MCPTools) Close() error { return m.sess.Close() }

// Defs lists the server's tools and keeps only the allowed ones, converting
// their input schemas into function declarations for the model.
func (m *MCPTools) Defs(ctx context.Context, allowed map[string]bool) ([]ToolDef, error) {
	list, err := m.sess.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}
	var defs []ToolDef
	for _, t := range list.Tools {
		if !allowed[t.Name] {
			continue
		}
		b, err := json.Marshal(t.InputSchema)
		if err != nil {
			return nil, err
		}
		var schema map[string]any
		if err := json.Unmarshal(b, &schema); err != nil {
			return nil, err
		}
		delete(schema, "$schema")
		delete(schema, "additionalProperties")
		defs = append(defs, ToolDef{Name: t.Name, Description: t.Description, Parameters: schema})
	}
	if len(defs) == 0 {
		return nil, fmt.Errorf("mcp server offers none of the allowed tools")
	}
	return defs, nil
}

// Call implements Tools.
func (m *MCPTools) Call(ctx context.Context, name string, args map[string]any) (string, error) {
	res, err := m.sess.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, c := range res.Content {
		if t, ok := c.(*mcp.TextContent); ok {
			b.WriteString(t.Text)
		}
	}
	return b.String(), nil
}
