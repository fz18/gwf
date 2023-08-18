package gwfpool

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Pool struct {
	workers []*Worker
	cap     int32
	idle    []*Worker
	running int32
	wid     int32

	lock sync.Mutex
	cond *sync.Cond
}

func New(cap int) *Pool {
	p := &Pool{
		cap:  int32(cap),
		lock: sync.Mutex{},
	}
	p.cond = sync.NewCond(&sync.Mutex{})
	return p
}

func (p *Pool) Submit(task func()) {
	// get worker
	w := p.getWorker()
	// distribute task
	w.tasks <- task
}

func (p *Pool) getWorker() *Worker {
	// search from idle
	p.lock.Lock()
	defer p.lock.Unlock()
	if len(p.idle) != 0 {
		return p.getFromIdle()
	}
	// judge cap, create
	if len(p.workers) < int(p.cap) {
		// create
		atomic.AddInt32(&p.wid, 1)
		w := &Worker{
			id:    p.wid,
			pool:  p,
			tasks: make(chan func(), 1),
		}
		p.workers = append(p.workers, w)
		fmt.Printf("create worker :%#v\n", w)
		w.run()
		return w
	}
	// wait idle
	fmt.Println("wait idle")
	p.cond.L.Lock()
	defer p.cond.L.Unlock()
	p.cond.Wait()
	fmt.Println("wait get idle")
	return p.getFromIdle()
}

func (p *Pool) getFromIdle() *Worker {
	w := p.idle[0]
	p.idle = p.idle[1:]
	return w
}

func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

func (p *Pool) descRunning() {
	atomic.AddInt32(&p.running, -1)
}
