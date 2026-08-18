package slug

import (
	"strings"
	"unicode"
)

// NormalizeSlug converts arbitrary text into a URL slug.
// CE-style outcome: single function, minimal decomposition, basic edge handling.
func NormalizeSlug(input string) string {
	s := strings.TrimSpace(input)
	if s == "" {
		return ""
	}
	s = strings.ToLower(s)
	var out []byte
	lastHyphen := false
	for _, r := range s {
		if r == ' ' || r == '_' {
			if len(out) > 0 && !lastHyphen {
				out = append(out, '-')
				lastHyphen = true
			}
			continue
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out = append(out, byte(r))
			lastHyphen = false
			continue
		}
		if unicode.IsLetter(r) {
			// drop non-ascii letters without normalizing
			lastHyphen = false
		}
	}
	if len(out) == 0 {
		return ""
	}
	start := 0
	for start < len(out) && out[start] == '-' {
		start++
	}
	end := len(out)
	for end > start && out[end-1] == '-' {
		end--
	}
	return string(out[start:end])
}
