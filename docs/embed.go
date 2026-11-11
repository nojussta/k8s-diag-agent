// Package docs embeds the agent output schema so the orchestrator and the
// model request use exactly the schema documented in docs/architecture.md.
package docs

import _ "embed"

// DiagnosisSchema is schemas/diagnosis.schema.json.
//
//go:embed schemas/diagnosis.schema.json
var DiagnosisSchema []byte
