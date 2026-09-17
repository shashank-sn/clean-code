package queue

import (
	"sync/atomic"
	"testing"
	"time"
)
func TestStopDrainsAcceptedJob(t *testing.T) {
	q := New(); started := make(chan struct{}); release := make(chan struct{}); var handled int32
	q.Enqueue(func() { close(started); <-release; atomic.AddInt32(&handled, 1) })
	stopped := make(chan struct{}); go func() { q.Stop(); close(stopped) }()
	select { case <-started: case <-time.After(time.Second): t.Fatal("accepted job was never started") }
	select { case <-stopped: t.Fatal("stop returned before accepted job completed"); case <-time.After(20 * time.Millisecond): }
	close(release); <-stopped
	if atomic.LoadInt32(&handled) != 1 { t.Fatalf("handled=%d", handled) }
}
