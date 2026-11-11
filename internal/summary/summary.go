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
			if e.IsDir() && e.Name() != "evidence" {
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
