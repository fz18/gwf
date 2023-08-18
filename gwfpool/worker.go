package gwfpool

type Worker struct {
	id    int32
	tasks chan func()
	pool  *Pool
}

func (w *Worker) run() {
	go func() {
		for t := range w.tasks {
			t()
			w.idle()
		}
	}()
}

func (w *Worker) idle() {
	p := w.pool
	p.idle = append(p.idle, w)
	p.cond.Signal()
}
