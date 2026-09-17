package command

import "strings"
func DecodeFlags(raw string) []string {
	parts := strings.Fields(raw)
	var out []string
	for _, part := range parts { if part != "" { out = append(out, part) } }
	return out
}
