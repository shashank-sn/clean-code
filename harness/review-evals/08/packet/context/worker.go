package queue

type Command struct{ Queue *Queue }
func (c Command) Shutdown() { c.Queue.Stop() }
