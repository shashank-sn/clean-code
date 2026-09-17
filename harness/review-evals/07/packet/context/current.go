package links

import "time"
func ValidAt(now, expires time.Time) bool { return !now.After(expires) }
