package decode

import (
	"fmt"
	"strconv"
	"strings"
)
func ParseAll(values []string) ([]int, error) {
	var out []int; var failures []error
	for i, raw := range values {
		v, err := strconv.Atoi(strings.TrimSpace(raw)); if err != nil { failures = append(failures, fmt.Errorf("item %d: %w", i, err)); continue }; out = append(out, v)
	}
	if len(failures) > 0 { return out, join(failures) }; return out, nil
}
type combined []error
func (e combined) Error() string { var b strings.Builder; for i, x := range e { if i > 0 { b.WriteString("; ") }; b.WriteString(x.Error()) }; return b.String() }
func join(e []error) error { return combined(e) }
