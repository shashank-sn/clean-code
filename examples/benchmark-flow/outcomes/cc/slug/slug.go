package slug

import (
	"strings"
	"unicode"
)

// NormalizeSlug returns a lowercase URL slug with single hyphens between words.
func NormalizeSlug(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	return trimHyphens(collapseHyphens(filterSlugRunes(trimmed)))
}

func filterSlugRunes(input string) []byte {
	lowered := strings.ToLower(input)
	out := make([]byte, 0, len(lowered))
	for _, r := range lowered {
		switch {
		case r == ' ' || r == '_':
			out = append(out, '-')
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			out = append(out, byte(r))
		case r == '-':
			out = append(out, '-')
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			// drop non-ascii alphanumeric without transliteration
		}
	}
	return out
}

func collapseHyphens(runes []byte) []byte {
	if len(runes) == 0 {
		return runes
	}
	out := make([]byte, 0, len(runes))
	lastHyphen := false
	for _, b := range runes {
		if b == '-' {
			if !lastHyphen && len(out) > 0 {
				out = append(out, '-')
				lastHyphen = true
			}
			continue
		}
		out = append(out, b)
		lastHyphen = false
	}
	return out
}

func trimHyphens(runes []byte) string {
	start := 0
	for start < len(runes) && runes[start] == '-' {
		start++
	}
	end := len(runes)
	for end > start && runes[end-1] == '-' {
		end--
	}
	if start >= end {
		return ""
	}
	return string(runes[start:end])
}
