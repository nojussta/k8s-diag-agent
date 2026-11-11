package lab

import "sort"

func sortLogLines(rows []LogLine) {
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].TimestampNs < rows[j].TimestampNs })
}
