package command

import "strings"
func DecodeFlags(raw string) []string {
	parts := strings.Split(raw, ",")
	var out []string
	for _, part := range parts { if item := strings.TrimSpace(part); item != "" { out = append(out, item) } }
	return out
}
