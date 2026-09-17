package queue

import "sync"
type Queue struct{ input chan func(); wg sync.WaitGroup }
func New() *Queue { q := &Queue{input: make(chan func(), 8)}; q.wg.Add(1); go q.run(); return q }
func (q *Queue) run() { defer q.wg.Done(); for job := range q.input { job() } }
func (q *Queue) Enqueue(job func()) { q.input <- job }
func (q *Queue) Stop() { close(q.input); q.wg.Wait() }
