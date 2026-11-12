// Package summary turns scenario runs into a Markdown table. It reads
// results/<scenario>/<timestamp>/run.json and reports, per scenario: number of runs, detection latency, and the fixed action set's call
// count and collection time.
package summary

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nojussta/k8s-diag-agent/internal/scenario"
)

// Fmt formats values as "mean ± stdev (min–max)", like a report table.
// The stdev is the sample standard deviation (n-1).
func Fmt(values []float64, digits int) string {
	if len(values) == 0 {
		return "–"
	}
	var sum float64
	lo, hi := values[0], values[0]
	for _, v := range values {
		sum += v
		lo, hi = math.Min(lo, v), math.Max(hi, v)
	}
	mean := sum / float64(len(values))
	if len(values) == 1 {
		return fmt.Sprintf("%.*f", digits, mean)
	}
	var ss float64
	for _, v := range values {
		ss += (v - mean) * (v - mean)
	}
	sd := math.Sqrt(ss / float64(len(values)-1))
	return fmt.Sprintf("%.*f ± %.*f (%.*f–%.*f)", digits, mean, digits, sd, digits, lo, digits, hi)
}

// Table renders the Markdown table for the given scenarios (all when empty).
func Table(root string, wanted []string) (string, error) {
	resultsDir := filepath.Join(root, "results")
	if len(wanted) == 0 {
		entries, err := os.ReadDir(resultsDir)
		if err != nil {
			return "", err
		}
		for _, e := range entries {
			if e.IsDir() && strings.HasPrefix(e.Name(), "s") && e.Name() != "evidence" {
				wanted = append(wanted, e.Name())
			}
		}
		sort.Strings(wanted)
	}
	var b strings.Builder
	b.WriteString("| Scenario | Runs | Detection latency, s | Fixed-sequence calls | Fixed-sequence duration, s | Reset |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, id := range wanted {
		paths, _ := filepath.Glob(filepath.Join(resultsDir, id, "*", "run.json"))
		sort.Strings(paths)
		var runs []scenario.Record
		for _, p := range paths {
			data, err := os.ReadFile(p)
			if err != nil {
				return "", err
			}
			var r scenario.Record
			if err := json.Unmarshal(data, &r); err != nil {
				return "", fmt.Errorf("%s: %w", p, err)
			}
			runs = append(runs, r)
		}
		var lat, calls, dur []float64
		detected, reset := 0, 0
		for _, r := range runs {
			if r.Steps["detect"] == "ok" {
				detected++
				lat = append(lat, r.DetectionLatencyS)
				if r.BaselineCollection != nil {
					calls = append(calls, float64(r.BaselineCollection.Calls))
					dur = append(dur, r.BaselineCollection.DurationS)
				}
			}
			if r.Steps["reset"] == "ok" {
				reset++
			}
		}
		fmt.Fprintf(&b, "| %s | %d/%d | %s | %s | %s | %d/%d ok |\n",
			id, detected, len(runs), Fmt(lat, 1), Fmt(calls, 0), Fmt(dur, 2), reset, len(runs))
	}
	return b.String(), nil
}

// MethodRow aggregates one method on one scenario.
type MethodRow struct {
	Scenario, Method       string
	Runs, Diagnosed        int
	Correct                int // category, component and change all right
	Category, Component    int
	Change, Notified       int
	Calls, Steps, Duration []float64
	TokensIn, TokensOut    []float64
}

type methodMetrics struct {
	ToolCalls    int     `json:"tool_calls"`
	ModelSteps   int     `json:"model_steps"`
	DurationS    float64 `json:"duration_s"`
	WaitS        float64 `json:"rate_limit_wait_s"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	Model        string  `json:"model"`
}

// Methods reads every run of the given scenarios (all when empty) and
// aggregates each diagnosis method. Runs recorded before methods existed are
// reported as method "agent-v0" (their single agent, judged by labctl judge).
// since filters run directories by name (timestamp prefix, "" = all).
func Methods(root string, wanted []string, since string) ([]MethodRow, error) {
	resultsDir := filepath.Join(root, "results")
	if len(wanted) == 0 {
		entries, err := os.ReadDir(resultsDir)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() && strings.HasPrefix(e.Name(), "s") {
				wanted = append(wanted, e.Name())
			}
		}
		sort.Strings(wanted)
	}
	var rows []MethodRow
	for _, id := range wanted {
		paths, _ := filepath.Glob(filepath.Join(resultsDir, id, "*", "run.json"))
		sort.Strings(paths)
		byMethod := map[string]*MethodRow{}
		var order []string
		get := func(name string) *MethodRow {
			if byMethod[name] == nil {
				byMethod[name] = &MethodRow{Scenario: id, Method: name}
				order = append(order, name)
			}
			return byMethod[name]
		}
		for _, p := range paths {
			if filepath.Base(filepath.Dir(p)) < since {
				continue
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return nil, err
			}
			var r scenario.Record
			if err := json.Unmarshal(data, &r); err != nil {
				return nil, fmt.Errorf("%s: %w", p, err)
			}
			if r.Steps["detect"] != "ok" {
				continue
			}
			add := func(name string, diagnosed bool, v *scenario.Verdict, metrics []byte, notified bool) {
				row := get(name)
				row.Runs++
				if notified {
					row.Notified++
				}
				if diagnosed && v != nil {
					row.Diagnosed++
					row.Correct += b2i(v.Correct)
					row.Category += b2i(v.Category)
					row.Component += b2i(v.Component)
					row.Change += b2i(v.Change)
				}
				var m methodMetrics
				if json.Unmarshal(metrics, &m) == nil && len(metrics) > 0 {
					row.Calls = append(row.Calls, float64(m.ToolCalls))
					row.Steps = append(row.Steps, float64(m.ModelSteps))
					row.Duration = append(row.Duration, m.DurationS-m.WaitS) // provider rate-limit waits excluded
					row.TokensIn = append(row.TokensIn, float64(m.InputTokens))
					row.TokensOut = append(row.TokensOut, float64(m.OutputTokens))
				}
			}
			names := make([]string, 0, len(r.Methods))
			for n := range r.Methods {
				names = append(names, n)
			}
			sort.Strings(names)
			for _, n := range names {
				m := r.Methods[n]
				add(n, m.Diagnosed, m.Judgement, m.Metrics, m.Notified)
			}
			if r.Agent != nil && len(r.Methods) == 0 {
				metrics, _ := os.ReadFile(filepath.Join(filepath.Dir(p), "metrics.json"))
				add("agent-v0", r.Agent.Diagnosed, r.Agent.Judgement, metrics, false)
			}
		}
		for _, n := range order {
			rows = append(rows, *byMethod[n])
		}
	}
	return rows, nil
}

// MethodTable renders Methods as Markdown.
func MethodTable(rows []MethodRow) string {
	var b strings.Builder
	b.WriteString("| Scenario | Method | Runs | Correct (all three) | Category | Component | Change | Read-only calls | Model steps | Duration, s | Notified |\n")
	b.WriteString("|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s | %s | %d | %d/%d | %d | %d | %d | %s | %s | %s | %d/%d |\n",
			r.Scenario, r.Method, r.Runs, r.Correct, r.Runs, r.Category, r.Component, r.Change,
			Fmt(r.Calls, 1), Fmt(r.Steps, 1), Fmt(r.Duration, 1), r.Notified, r.Runs)
	}
	return b.String()
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
