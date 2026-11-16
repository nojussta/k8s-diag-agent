// Command labctl drives the experimental environment: environment checks,
// fault scenarios, the fixed read-only evidence collection (A1-A11) and result
// summaries.
//
//	labctl check                                  verify the environment (11 checks)
//	labctl scenario <id> [--repeat N] [--keep] [--method NAME=CMD]...
//	labctl evidence (--latest | --alert FILE) [--out DIR]
//	labctl summary [--since STAMP] [scenario ...] Markdown tables from results/
//	labctl alerts                                 notifications the receiver got
//	labctl judge <run-dir>...                     (re)judge the diagnoses of recorded runs
//	labctl rules <run-dir>...                     (re)apply the rule-based baseline to recorded runs
//	labctl rerun --method NAME=CMD <run-dir>...   re-run an offline method (e.g. pipeline) on recorded evidence
//	labctl watch --method CMD [--notify URL]      investigate every new alert at the receiver (queue + dedup)
//	labctl storm [N]                              send N synthetic alerts at once to the in-cluster agent
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
	"strings"
	"syscall"
	"time"

	"github.com/nojussta/k8s-diag-agent/internal/alertmsg"
	"github.com/nojussta/k8s-diag-agent/internal/envcheck"
	"github.com/nojussta/k8s-diag-agent/internal/evidence"
	"github.com/nojussta/k8s-diag-agent/internal/lab"
	"github.com/nojussta/k8s-diag-agent/internal/scenario"
	"github.com/nojussta/k8s-diag-agent/internal/summary"
	"github.com/nojussta/k8s-diag-agent/internal/watch"
)

const usage = `usage: labctl <command>

  check                                   verify the environment
  scenario <id> [--repeat N] [--keep] [--incluster] [--method NAME=CMD]...
  evidence (--latest | --alert FILE) [--out DIR]
  summary [--since STAMP] [scenario ...]
  alerts
  judge <run-dir>...
  rules <run-dir>...
  rerun --method NAME=CMD <run-dir>...
  watch --method CMD [--notify URL] [--once]
  storm [N]`

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
	case "rules":
		err = cmdRules(args)
	case "rerun":
		err = cmdRerun(ctx, args)
	case "watch":
		err = cmdWatch(ctx, args)
	case "storm":
		err = cmdStorm(ctx, args)
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
	var methods methodFlags
	fs.Var(&methods, "method", "diagnosis method NAME=CMD (repeatable); {alert}, {out}, {evidence}, {notify} are substituted. The rule-based baseline always runs.")
	agentCmd := fs.String("agent-cmd", "", "shorthand for --method agent=CMD")
	inCluster := fs.Bool("incluster", false, "let the agent deployed in the cluster diagnose the alert on its own (needs make agent-on)")
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
	if *agentCmd != "" {
		methods = append(methods, scenario.Method{Name: "agent", Cmd: *agentCmd})
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
		rec, err := runner.RunOnce(ctx, sc, scenario.Options{Keep: *keep, Methods: methods, InCluster: *inCluster}, out)
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
	fs := flag.NewFlagSet("summary", flag.ExitOnError)
	since := fs.String("since", "", "only runs whose directory name (timestamp) is >= this, e.g. 20261112")
	_ = fs.Parse(args)
	root, err := lab.FindRoot()
	if err != nil {
		return err
	}
	table, err := summary.Table(root, fs.Args())
	if err != nil {
		return err
	}
	fmt.Print("Detection and fixed-sequence collection\n\n" + table)
	rows, err := summary.Methods(root, fs.Args(), *since)
	if err != nil {
		return err
	}
	fmt.Print("\nDiagnosis methods (correct = category, component and triggering change)\n\n" + summary.MethodTable(rows))
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

// cmdJudge (re)judges recorded runs with the current ground truth and judge:
// every method directory with a diagnosis.json, and the diagnosis.json of runs
// made before methods existed. The verdicts are written into run.json.
func cmdJudge(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: labctl judge <run-dir>...")
	}
	root, err := lab.FindRoot()
	if err != nil {
		return err
	}
	for _, dir := range args {
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
		if diag, err := os.ReadFile(filepath.Join(dir, "diagnosis.json")); err == nil {
			if rec.Agent == nil {
				rec.Agent = &scenario.AgentRun{}
			}
			v := scenario.Judge(sc.GroundTruth, diag)
			rec.Agent.Diagnosed, rec.Agent.Correct, rec.Agent.Verdict, rec.Agent.Judgement = true, v.Correct, v.Text, &v
			fmt.Printf("%s agent: %s\n", dir, v.Text)
		}
		for name, m := range rec.Methods {
			diag, err := os.ReadFile(filepath.Join(dir, name, "diagnosis.json"))
			if err != nil {
				continue
			}
			v := scenario.Judge(sc.GroundTruth, diag)
			m.Diagnosed, m.Judgement = true, &v
			fmt.Printf("%s %s: %s\n", dir, name, v.Text)
		}
		out, _ := json.MarshalIndent(rec, "", "  ")
		if err := os.WriteFile(filepath.Join(dir, "run.json"), out, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// cmdRules re-applies the rule-based baseline to recorded runs (after the
// rules change) and updates run.json.
func cmdRules(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: labctl rules <run-dir>...")
	}
	root, err := lab.FindRoot()
	if err != nil {
		return err
	}
	for _, dir := range args {
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
		m, err := scenario.ApplyRules(sc, dir)
		if err != nil {
			fmt.Printf("%s: skipped (%v)\n", dir, err)
			continue
		}
		if rec.Methods == nil {
			rec.Methods = map[string]*scenario.MethodRun{}
		}
		rec.Methods["rules"] = m
		out, _ := json.MarshalIndent(rec, "", "  ")
		if err := os.WriteFile(filepath.Join(dir, "run.json"), out, 0o644); err != nil {
			return err
		}
		fmt.Printf("%s rules (%s): %s\n", dir, m.Rule, m.Judgement.Text)
	}
	return nil
}

// cmdRerun re-runs one method that needs only recorded data (the pipeline
// reads alert.json and evidence.json) on recorded runs, so a method change
// can be evaluated on exactly the same evidence without new faults.
func cmdRerun(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("rerun", flag.ExitOnError)
	var methods methodFlags
	fs.Var(&methods, "method", "NAME=CMD, as for scenario")
	_ = fs.Parse(args)
	if len(methods) != 1 || fs.NArg() == 0 {
		return errors.New("usage: labctl rerun --method NAME=CMD <run-dir>...")
	}
	l, err := lab.New()
	if err != nil {
		return err
	}
	for _, dir := range fs.Args() {
		var rec scenario.Record
		b, err := os.ReadFile(filepath.Join(dir, "run.json"))
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &rec); err != nil {
			return err
		}
		if _, err := os.Stat(filepath.Join(dir, "evidence.json")); err != nil {
			fmt.Printf("%s: skipped (no evidence.json)\n", dir)
			continue
		}
		sc, err := scenario.Load(l.Root, rec.Scenario)
		if err != nil {
			return err
		}
		m, err := scenario.RerunMethod(ctx, l, sc, methods[0], dir)
		if err != nil {
			return err
		}
		if rec.Methods == nil {
			rec.Methods = map[string]*scenario.MethodRun{}
		}
		rec.Methods[methods[0].Name] = m
		out, _ := json.MarshalIndent(rec, "", "  ")
		if err := os.WriteFile(filepath.Join(dir, "run.json"), out, 0o644); err != nil {
			return err
		}
		verdict := "no diagnosis"
		if m.Judgement != nil {
			verdict = m.Judgement.Text
		}
		fmt.Printf("%s %s: %s\n", dir, methods[0].Name, verdict)
	}
	return nil
}

// cmdWatch is the automatic path of the design: it polls the alert receiver,
// picks firing lab alerts that were not investigated in the last 30 min, and
// runs the agent for each, one at a time. The agent posts its report to
// --notify (default $NOTIFY_WEBHOOK_URL); without one, a local stand-in
// channel prints the messages here.
func cmdWatch(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("watch", flag.ExitOnError)
	method := fs.String("method", "", "agent command; {alert}, {out} and {notify} are substituted")
	notifyURL := fs.String("notify", os.Getenv("NOTIFY_WEBHOOK_URL"), "chat webhook for the reports")
	once := fs.Bool("once", false, "exit after the first investigation")
	_ = fs.Parse(args)
	if *method == "" {
		return errors.New("usage: labctl watch --method CMD [--notify URL] [--once]")
	}
	l, err := lab.New()
	if err != nil {
		return err
	}
	var local *scenario.Sink
	if *notifyURL == "" {
		if local, err = scenario.StartSink(); err != nil {
			return err
		}
		defer local.Close()
		*notifyURL = local.URL("watch")
		logf("no NOTIFY_WEBHOOK_URL: reports go to a local stand-in channel and are printed here")
	}
	since := float64(time.Now().UnixNano()) / 1e9
	dedup := &watch.Dedup{}
	logf("watching the alert receiver (Ctrl-C to stop)")
	for {
		ns, err := l.Notifications(ctx, since)
		if err != nil {
			logf("receiver: %v", err)
		}
		for _, n := range ns {
			since = max(since, n.ReceivedAt)
		}
		for _, a := range watch.Pick(ns, dedup, time.Now()) {
			out := filepath.Join(l.Root, "results", "watch", time.Now().Format("20060102-150405")+"-"+a.Labels["alertname"])
			if err := os.MkdirAll(out, 0o755); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(a, "", "  ")
			if err := os.WriteFile(filepath.Join(out, "alert.json"), b, 0o644); err != nil {
				return err
			}
			cmd := strings.NewReplacer("{alert}", filepath.Join(out, "alert.json"), "{out}", out, "{notify}", *notifyURL).Replace(*method)
			logf("investigating %s %s/%s -> %s", a.Labels["alertname"], a.Labels["namespace"], firstNonEmpty(a.Labels["pod"], a.Labels["service"], a.Labels["deployment"]), out)
			code, output := l.Sh(ctx, cmd, 10*time.Minute)
			_ = os.WriteFile(filepath.Join(out, "output.txt"), []byte(output), 0o644)
			logf("agent finished (exit %d)", code)
			if local != nil {
				for _, m := range local.Take("watch") {
					fmt.Printf("----- message delivered to the channel -----\n%s-------------------------------------------\n", m.Text)
				}
			}
			if *once {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(5 * time.Second):
		}
	}
}

// cmdStorm sends N synthetic firing alerts in one notification to the
// in-cluster agent and reports what its queue did with them: how many were
// investigated, how many were refused by the budget, and how long the queue
// took to drain. The alerts name pods that do not exist, so every
// investigation is short.
func cmdStorm(ctx context.Context, args []string) error {
	n := 20
	if len(args) > 0 {
		if _, err := fmt.Sscanf(args[0], "%d", &n); err != nil || n < 1 || n > 200 {
			return errors.New("usage: labctl storm [N] (1-200)")
		}
	}
	l, err := lab.New()
	if err != nil {
		return err
	}
	start := time.Now()
	stamp := start.Format("20060102-150405")
	payload := alertmsg.Payload{Version: "4", Status: "firing", Receiver: "diag-agent"}
	for i := 0; i < n; i++ {
		payload.Alerts = append(payload.Alerts, alertmsg.Alert{
			Status: "firing", StartsAt: start.UTC().Format(time.RFC3339), Fingerprint: fmt.Sprintf("storm-%s-%02d", stamp, i),
			Labels: map[string]string{"alertname": "LabPodCrashLooping", "lab": "kursinis", "severity": "warning",
				"namespace": "demo", "pod": fmt.Sprintf("storm-%s-%02d", stamp, i), "container": "app"},
			Annotations: map[string]string{"summary": "synthetic alert of labctl storm"},
		})
	}
	if _, err := l.ProxyPostJSON(ctx, lab.Receiver, "alerts", payload); err != nil {
		return err
	}
	logf("sent %d alerts in one notification; waiting for the queue to drain", n)
	mine := func(list []scenario.Investigation) (out []scenario.Investigation) {
		for _, it := range list {
			if strings.HasPrefix(it.Alert.Fingerprint, "storm-"+stamp) {
				out = append(out, it)
			}
		}
		return out
	}
	var items []scenario.Investigation
	deadline := start.Add(20 * time.Minute)
	for time.Now().Before(deadline) {
		list, err := scenario.Investigations(ctx, l)
		if err == nil {
			items = mine(list)
			open := 0
			for _, it := range items {
				if it.Status == "queued" || it.Status == "running" {
					open++
				}
			}
			if len(items) > 0 && open == 0 {
				break
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
	if len(items) == 0 {
		return errors.New("the in-cluster agent queued nothing (run `make agent-on` first)")
	}
	count := map[string]int{}
	var first, last time.Time
	var total time.Duration
	for _, it := range items {
		count[it.Status]++
		if it.Started != nil && it.Finished != nil {
			total += it.Finished.Sub(*it.Started)
			if first.IsZero() || it.Started.Before(first) {
				first = *it.Started
			}
			if it.Finished.After(last) {
				last = *it.Finished
			}
		}
	}
	msgs, _ := scenario.Messages(ctx, l)
	storm := 0
	for _, m := range msgs {
		if strings.HasPrefix(m.Text, "Alert storm:") && m.Received >= start.UTC().Format(time.RFC3339) {
			storm++
		}
	}
	investigated := count["done"] + count["failed"]
	res := map[string]any{"sent": n, "investigated": investigated, "done": count["done"], "failed": count["failed"],
		"skipped": count["skipped"], "storm_messages": storm, "queue_drained_s": last.Sub(start).Seconds(),
		"sum_investigation_s": total.Seconds(), "items": items}
	out := filepath.Join(l.Root, "results", "storm", stamp)
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(res, "", "  ")
	if err := os.WriteFile(filepath.Join(out, "storm.json"), b, 0o644); err != nil {
		return err
	}
	logf("%d sent: %d investigated one at a time (%d with a diagnosis), %d refused by the budget, %d summary message(s); queue drained %.0f s after the notification -> %s",
		n, investigated, count["done"], count["skipped"], storm, last.Sub(start).Seconds(), out)
	return nil
}

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}
	return ""
}

// methodFlags collects repeated --method NAME=CMD flags.
type methodFlags []scenario.Method

func (m *methodFlags) String() string { return fmt.Sprint(*m) }

func (m *methodFlags) Set(v string) error {
	name, cmd, ok := strings.Cut(v, "=")
	if !ok || name == "" || cmd == "" || name == "rules" || strings.ContainsAny(name, "/ .") {
		return fmt.Errorf("--method wants NAME=CMD with a simple name other than \"rules\", got %q", v)
	}
	*m = append(*m, scenario.Method{Name: name, Cmd: cmd})
	return nil
}
