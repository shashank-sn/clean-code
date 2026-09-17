package payments

import (
	"fmt"
	"runtime"
)

var seen = map[string]string{}
var charges int
var barrier chan struct{}
var reached chan struct{}
func Reset() { seen = map[string]string{}; charges = 0; barrier = nil; reached = nil }
func SetTestBarrier(b chan struct{}, r chan struct{}) { barrier, reached = b, r }
func UsesAtomicGuard() bool { return false }
func CreatePayment(key string, cents int) (string, error) {
	if id, ok := seen[key]; ok { return id, nil }
	if barrier != nil { reached <- struct{}{}; <-barrier }
	runtime.Gosched()
	id := fmt.Sprintf("pay-%d", charges+1)
	seen[key] = id
	charges++
	return id, nil
}
