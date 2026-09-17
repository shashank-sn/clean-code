package queue

import "sync"
type Queue struct{ input chan func(); stop chan struct{}; wg sync.WaitGroup }
func New() *Queue { q := &Queue{input: make(chan func(), 8), stop: make(chan struct{})}; q.wg.Add(1); go q.run(); return q }
func (q *Queue) run() { defer q.wg.Done(); for { select { case job := <-q.input: job(); case <-q.stop: return } } }
func (q *Queue) Enqueue(job func()) { q.input <- job }
func (q *Queue) Stop() { close(q.stop) }
