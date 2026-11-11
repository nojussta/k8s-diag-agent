// Command labctl drives the experimental environment: environment checks,
// fault scenarios, the fixed read-only evidence collection (A1-A11) and result
// summaries.
//
//	labctl check                                  verify the environment (11 checks)
//	labctl scenario <id> [--repeat N] [--keep] [--agent-cmd CMD]
//	labctl evidence (--latest | --alert FILE) [--out DIR]
//	labctl summary [scenario ...]                 Markdown table from results/
//	labctl alerts                                 notifications the receiver got
//	labctl judge <run-dir>                        compare diagnosis.json with the scenario ground truth
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
	"github.com/nojussta/k8s-diag-agent/internal/envcheck"
	"github.com/nojussta/k8s-diag-agent/internal/evidence"
	"github.com/nojussta/k8s-diag-agent/internal/lab"
	"github.com/nojussta/k8s-diag-agent/internal/scenario"
	"github.com/nojussta/k8s-diag-agent/internal/summary"
)

const usage = `usage: labctl <command>

  check                                   verify the environment
  scenario <id> [--repeat N] [--keep] [--agent-cmd CMD]
  evidence (--latest | --alert FILE) [--out DIR]
  summary [scenario ...]
  alerts
  judge <run-dir>`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var err error
	switch cmd, args := os.Args[1], os.Args[2:]; cmd {
	case "check":
		err = cmdCheck(ctx)
	case "scenario":
		err = cmdScenario(ctx, args)
	case "evidence":
		err = cmdEvidence(ctx, args)
	case "summary":
		err = cmdSummary(args)
	case "alerts":
		err = cmdAlerts(ctx)
	case "judge":
		err = cmdJudge(args)
	case "-h", "--help", "help":
		fmt.Println(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n%s\n", cmd, usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func logf(format string, args ...any) {
	fmt.Printf("%s %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}

func cmdCheck(ctx context.Context) error {
	l, err := lab.New()
	if err != nil {
		return err
	}
	results := envcheck.Run(ctx, l, func(r envcheck.Result) {
		status, detail := "PASS", ""
		if !r.OK {
			status = "FAIL"
		}
		if r.Detail != "" {
			detail = " - " + r.Detail
		}
		fmt.Printf("[%s] %s%s\n", status, r.Name, detail)
	})
	failed := 0
	for _, r := range results {
		if !r.OK {
			failed++
		}
	}
	fmt.Printf("\n%d/%d checks passed\n", len(results)-failed, len(results))
	if failed > 0 {
		os.Exit(1)
	}
	return nil
}

func cmdScenario(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("scenario", flag.ExitOnError)
	keep := fs.Bool("keep", false, "do not reset after the run")
	repeat := fs.Int("repeat", 1, "number of runs")
	agentCmd := fs.String("agent-cmd", "", "command to run the agent; {alert} and {out} are substituted")
	// Allow the scenario id before or after the flags.
	var id string
	if len(args) > 0 && args[0] != "" && args[0][0] != '-' {
		id, args = args[0], args[1:]
	}
	_ = fs.Parse(args)
	if id == "" {
		id = fs.Arg(0)
	}
	if id == "" {
		return errors.New("scenario id required, e.g. labctl scenario s1-oomkilled")
	}
	if *keep && *repeat > 1 {
		return errors.New("--keep cannot be combined with --repeat")
	}

	l, err := lab.New()
	if err != nil {
		return err
	}
	sc, err := scenario.Load(l.Root, id)
	if err != nil {
		return err
	}
	if sc.Status != "implemented" {
		logf("WARNING: scenario status is '%s', commands are not validated yet", sc.Status)
	}
	runner := &scenario.Runner{Lab: l, Log: logf}
	var latencies []float64
	for i := 0; i < *repeat; i++ {
		out := filepath.Join(l.Root, "results", sc.ID, time.Now().Format("20060102-150405"))
		rel, _ := filepath.Rel(l.Root, out)
		logf("run %d/%d -> %s", i+1, *repeat, rel)
		rec, err := runner.RunOnce(ctx, sc, scenario.Options{Keep: *keep, AgentCmd: *agentCmd}, out)
		if rec != nil && rec.DetectionLatencyS > 0 {
			latencies = append(latencies, rec.DetectionLatencyS)
		}
		if err != nil {
			return err
		}
	}
	if len(latencies) > 0 {
		lo, hi, sum := latencies[0], latencies[0], 0.0
		for _, v := range latencies {
			sum += v
			lo, hi = min(lo, v), max(hi, v)
		}
		logf("detection latency: min %v s, max %v s, mean %.1f s over %d run(s)", lo, hi, sum/float64(len(latencies)), len(latencies))
	}
	return nil
}

func cmdEvidence(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("evidence", flag.ExitOnError)
	latest := fs.Bool("latest", false, "last firing alert at the receiver")
	alertFile := fs.String("alert", "", "alert from a JSON file")
	out := fs.String("out", "", "output directory (default results/evidence)")
	_ = fs.Parse(args)
	if *latest == (*alertFile != "") {
		return errors.New("use exactly one of --latest or --alert FILE")
	}
	l, err := lab.New()
	if err != nil {
		return err
	}
	var alert alertmsg.Alert
	if *latest {
		ns, err := l.Notifications(ctx, 0)
		if err != nil {
			return err
		}
		found := false
		for i := len(ns) - 1; i >= 0 && !found; i-- {
			for _, a := range ns[i].Payload.Alerts {
				if a.Status == "firing" && a.Labels["alertname"] != "LabEnvSelfTest" {
					alert, found = a, true
					break
				}
			}
		}
		if !found {
			return errors.New("no firing alert at the receiver")
		}
	} else {
		data, err := os.ReadFile(*alertFile)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &alert); err != nil {
			return err
		}
	}
	dir := *out
	if dir == "" {
		dir = filepath.Join(l.Root, "results", "evidence")
	}
	ev := evidence.Collect(ctx, l.Cs, l, alert)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(ev, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "evidence.json"), data, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "evidence.md"), []byte(ev.Markdown()), 0o644); err != nil {
		return err
	}
	fmt.Printf("%d read-only calls in %v s -> %s\n", ev.TotalCalls, ev.TotalDurationS, dir)
	return nil
}

func cmdSummary(args []string) error {
	root, err := lab.FindRoot()
	if err != nil {
		return err
	}
	table, err := summary.Table(root, args)
	if err != nil {
		return err
	}
	fmt.Print(table)
	return nil
}

func cmdAlerts(ctx context.Context) error {
	l, err := lab.New()
	if err != nil {
		return err
	}
	ns, err := l.Notifications(ctx, 0)
	if err != nil {
		return err
	}
	for _, n := range ns {
		ts := time.Unix(0, int64(n.ReceivedAt*1e9)).Format("15:04:05")
		for _, a := range n.Payload.Alerts {
			fmt.Println(ts, a.Status, a.Labels["alertname"], a.Labels["namespace"], a.Labels["pod"])
		}
	}
	return nil
}

// cmdJudge compares a recorded run's diagnosis.json with its scenario's ground truth
// and writes the verdict into run.json.
func cmdJudge(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: labctl judge <run-dir>")
	}
	dir := args[0]
	root, err := lab.FindRoot()
	if err != nil {
		return err
	}
	var rec scenario.Record
	b, err := os.ReadFile(filepath.Join(dir, "run.json"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &rec); err != nil {
		return err
	}
	sc, err := scenario.Load(root, rec.Scenario)
	if err != nil {
		return err
	}
	diag, err := os.ReadFile(filepath.Join(dir, "diagnosis.json"))
	if err != nil {
		return err
	}
	if rec.Agent == nil {
		rec.Agent = &scenario.AgentRun{}
	}
	rec.Agent.Diagnosed = true
	rec.Agent.Correct, rec.Agent.Verdict = scenario.Judge(sc.GroundTruth, diag)
	out, _ := json.MarshalIndent(rec, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "run.json"), out, 0o644); err != nil {
		return err
	}
	fmt.Printf("%s: correct=%v (%s)\n", dir, rec.Agent.Correct, rec.Agent.Verdict)
	return nil
}
