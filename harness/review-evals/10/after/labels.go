package labels

import "sort"
func Normalize(values []string) []string {
	// Sorting a private copy keeps caller-owned input unchanged.
	copyOf := append([]string(nil), values...); sort.Strings(copyOf)
	out := make([]string, 0, len(copyOf))
	for _, v := range copyOf { if len(out) == 0 || out[len(out)-1] != v { out = append(out, v) } }
	return out
}
