package links

import (
	"testing"
	"time"
)
func TestExpiryIsStrict(t *testing.T) {
	expires := time.Unix(100, 0)
	if ValidAt(expires, expires) { t.Fatal("link was valid at its expiry instant") }
	if !ValidAt(expires.Add(-time.Nanosecond), expires) { t.Fatal("link expired early") }
}
