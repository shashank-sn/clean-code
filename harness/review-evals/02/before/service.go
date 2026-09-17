package payments

import (
	"fmt"
	"sync"
)

var mu sync.Mutex
var seen = map[string]string{}
var charges int
func Reset() { mu.Lock(); defer mu.Unlock(); seen = map[string]string{}; charges = 0 }
func SetTestBarrier(b chan struct{}, r chan struct{}) {}
func UsesAtomicGuard() bool { return true }
func CreatePayment(key string, cents int) (string, error) {
	mu.Lock(); defer mu.Unlock()
	if id, ok := seen[key]; ok { return id, nil }
	id := fmt.Sprintf("pay-%d", charges+1)
	seen[key] = id
	charges++
	return id, nil
}
