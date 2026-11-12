package scenario

import (
	"encoding/json"
	"strings"
)

// Verdict is the automatic comparison of one diagnosis with the ground truth.
// A diagnosis is correct only when all three parts are right:
//
//   - Category:  root_cause.category is one of the accepted categories (exact match);
//   - Component: root_cause.component names the workload that has to be fixed;
//   - Change:    root_cause.explanation or summary names the change that
//     triggered the fault (for example "--stress-memory=256" or the image tag).
//
// The category alone is not enough: "OOMKilled" without the release that
// raised the memory need is a symptom, not a root cause.
type Verdict struct {
	Category  bool   `json:"category"`
	Component bool   `json:"component"`
	Change    bool   `json:"change"`
	Correct   bool   `json:"correct"`
	Text      string `json:"text"`
}

// Score is the share of the three parts that are right (0, 1/3, 2/3 or 1).
func (v Verdict) Score() float64 {
	n := 0
	for _, ok := range []bool{v.Category, v.Component, v.Change} {
		if ok {
			n++
		}
	}
	return float64(n) / 3
}

// Judge compares a diagnosis.json with the scenario's ground truth.
func Judge(gt GroundTruth, diagnosis []byte) Verdict {
	var d struct {
		Summary   string `json:"summary"`
		RootCause struct {
			Category    string `json:"category"`
			Component   string `json:"component"`
			Explanation string `json:"explanation"`
		} `json:"root_cause"`
	}
	if err := json.Unmarshal(diagnosis, &d); err != nil {
		return Verdict{Text: "diagnosis.json unreadable"}
	}
	cats := gt.Match.CategoryAny
	if len(cats) == 0 {
		cats = []string{gt.Category}
	}
	v := Verdict{
		Category:  contains(cats, d.RootCause.Category),
		Component: anyIn(d.RootCause.Component, gt.Match.ComponentAny) && !anyIn(d.RootCause.Component, gt.Match.ComponentNot),
		Change:    anyIn(d.RootCause.Explanation+" "+d.Summary, gt.Match.ChangeAny),
	}
	v.Correct = v.Category && v.Component && v.Change
	var wrong []string
	for _, p := range []struct {
		ok   bool
		name string
	}{{v.Category, "category"}, {v.Component, "component"}, {v.Change, "change"}} {
		if !p.ok {
			wrong = append(wrong, p.name)
		}
	}
	if len(wrong) == 0 {
		v.Text = "category, component and change match"
	} else {
		v.Text = "wrong: " + strings.Join(wrong, ", ")
	}
	return v
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// anyIn reports whether text contains one of the words (case-insensitive).
// An empty word list never matches, so a scenario must state what it expects.
func anyIn(text string, words []string) bool {
	t := strings.ToLower(text)
	for _, w := range words {
		if w != "" && strings.Contains(t, strings.ToLower(w)) {
			return true
		}
	}
	return false
}
