package rules

import "strings"
func Parse(raw string) []string {
	var parts []string; var current strings.Builder; escaped := false
	for _, r := range raw {
		if escaped { current.WriteRune(r); escaped = false; continue }
		if r == '\\' { escaped = true; continue }
		if r == ';' { if s := strings.TrimSpace(current.String()); s != "" { parts = append(parts, s) }; current.Reset(); continue }
		current.WriteRune(r)
	}
	if escaped { current.WriteByte('\\') }
	if s := strings.TrimSpace(current.String()); s != "" { parts = append(parts, s) }
	return parts
}
