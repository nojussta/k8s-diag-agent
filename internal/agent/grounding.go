package agent

import (
	"fmt"
	"regexp"
	"strings"
)

// Grounding is the content check of a diagnosis. ParseDiagnosis only proves
// that every evidence item cites a call that ran; it does not prove that the
// cited result says what the item claims. This check closes part of that gap
// with a rule that needs no model: every number an evidence item quotes (a
// port, an exit code, a limit) must occur in the result it cites, and every
// number in the explanation must occur in some result or in the alert.
//
// The check is deliberately not a reason to reject an answer: numbers can be
// rephrased ("128Mi" for 134217728 bytes), so a miss is a warning shown to the
// engineer and a metric, not an error.
type Grounding struct {
	// Facts is the number of evidence items that quote at least one number;
	// FactsGrounded is how many of them have all their numbers in the cited result.
	Facts         int `json:"facts"`
	FactsGrounded int `json:"facts_grounded"`
	// Numbers counts every checked number (evidence and explanation).
	Numbers         int `json:"numbers"`
	NumbersGrounded int `json:"numbers_grounded"`
	// Ungrounded lists what was not found, e.g. `evidence[1] (t3): 9899`.
	Ungrounded []string `json:"ungrounded,omitempty"`
}

// numberToken matches digit runs that do not continue a word, so "p95" or
// "t12" are not numbers while "exitCode=137" and "128Mi" are. One-digit
// numbers ("revision 2", "1 of 2") are too common to carry evidence.
var numberToken = regexp.MustCompile(`(?:^|[^A-Za-z0-9_])(\d{2,})`)

func numbers(s string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range numberToken.FindAllStringSubmatch(s, -1) {
		n := m[1]
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out
}

// CheckGrounding compares the numbers of a diagnosis with the tool results.
// results maps a call id (tN or AN) to the text the model saw; context is any
// other text the model was given (the alert).
func CheckGrounding(d *Diagnosis, results map[string]string, context string) Grounding {
	var g Grounding
	if d == nil {
		return g
	}
	var all strings.Builder
	all.WriteString(context)
	for _, r := range results {
		all.WriteString("\n")
		all.WriteString(r)
	}
	everything := all.String()

	for i, e := range d.Evidence {
		nums := numbers(e.Fact)
		if len(nums) == 0 {
			continue
		}
		g.Facts++
		ok := true
		for _, n := range nums {
			g.Numbers++
			if strings.Contains(results[e.ToolCall], n) {
				g.NumbersGrounded++
				continue
			}
			ok = false
			g.Ungrounded = append(g.Ungrounded, fmt.Sprintf("evidence[%d] (%s): %s", i, e.ToolCall, n))
		}
		if ok {
			g.FactsGrounded++
		}
	}
	for _, n := range numbers(d.RootCause.Explanation + " " + d.Summary) {
		g.Numbers++
		if strings.Contains(everything, n) {
			g.NumbersGrounded++
			continue
		}
		g.Ungrounded = append(g.Ungrounded, "explanation: "+n)
	}
	return g
}

// truncatedResult reports whether a tool result carries the tool layer's
// "truncated": true flag (the 8 KB cap removed part of the data).
func truncatedResult(result string) bool {
	return strings.Contains(result, `"truncated":true`)
}
