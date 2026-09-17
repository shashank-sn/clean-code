package links

import "time"
func Authorize(now, expires time.Time) bool { return ValidAt(now, expires) }
