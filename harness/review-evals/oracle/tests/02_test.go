package payments

import (
	"sync"
	"testing"
)

func TestConcurrentIdempotency(t *testing.T) {
	Reset(); const n = 24
	if !UsesAtomicGuard() {
		gate := make(chan struct{}); reached := make(chan struct{}, n); SetTestBarrier(gate, reached)
		var ready sync.WaitGroup; ready.Add(n); var wg sync.WaitGroup; wg.Add(n)
		for i := 0; i < n; i++ { go func() { defer wg.Done(); ready.Done(); _, _ = CreatePayment("retry-key", 900) }() }
		ready.Wait(); for i := 0; i < n; i++ { <-reached }; close(gate); wg.Wait()
	} else {
		var wg sync.WaitGroup; wg.Add(n)
		for i := 0; i < n; i++ { go func() { defer wg.Done(); _, _ = CreatePayment("retry-key", 900) }() }; wg.Wait()
	}
	if charges != 1 { t.Fatalf("charges=%d, want one", charges) }
}
