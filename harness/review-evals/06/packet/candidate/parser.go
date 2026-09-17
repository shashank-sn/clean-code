package rules

import "strings"
func Parse(raw string) []string {
	parts := strings.Split(raw, ";"); var out []string
	for _, s := range parts { if s = strings.TrimSpace(s); s != "" { out = append(out, s) } }
	return out
}
