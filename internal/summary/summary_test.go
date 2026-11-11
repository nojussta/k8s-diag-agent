package summary

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFmt(t *testing.T) {
	cases := []struct {
		in     []float64
		digits int
		want   string
	}{
		{nil, 1, "–"},
		{[]float64{71.4}, 1, "71.4"},
		{[]float64{71, 85}, 1, "78.0 ± 9.9 (71.0–85.0)"},
		{[]float64{20, 20, 20}, 0, "20 ± 0 (20–20)"},
	}
	for _, c := range cases {
		if got := Fmt(c.in, c.digits); got != c.want {
			t.Errorf("Fmt(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

// testdata holds the three recorded run.json files behind the expected table.
// The summary must reproduce the published numbers; the fixture is pinned so
// that newly recorded runs cannot break the test.
func TestTableReproducesPaperNumbers(t *testing.T) {
	root, _ := filepath.Abs("testdata")
	out, err := Table(root, []string{"s1-oomkilled"})
	if err != nil {
		t.Fatal(err)
	}
	want := "| s1-oomkilled | 3/3 | 92.9 ± 6.1 (86.7–98.9) | 20 ± 0 (20–20) | 0.07 ± 0.01 (0.06–0.08) | 3/3 ok |"
	if !strings.Contains(out, want) {
		t.Fatalf("table does not reproduce the recorded numbers:\n%s\nwant row: %s", out, want)
	}
}
