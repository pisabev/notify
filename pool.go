package notify

import (
	"fmt"
	"sync"
)

var ErrNoWorkers = fmt.Errorf("attempting to create worker pool with less than 1 worker")
var ErrNegativeChannelSize = fmt.Errorf("attempting to create worker pool with a negative channel size")

type Pool interface {
	// Stop stops the pool - stop or stop and flush
	Stop(bool)
	// AddTask adds a task - non-blocking
	AddTask(Task)
}

type Task interface {
	// Execute performs the work
	Execute()
	// OnDone called after execute
	OnDone()
}

type TaskPool struct {
	numWorkers int
	tasks      chan Task
	wg         *sync.WaitGroup

	// Ensure the pool can only be stopped once
	stop    sync.Once
	stopped bool

	// Close to signal the workers to stop working
	quit chan struct{}
}

func NewTaskPool(numWorkers int, channelSize int) (Pool, error) {
	if numWorkers <= 0 {
		return nil, ErrNoWorkers
	}
	if channelSize < 0 {
		return nil, ErrNegativeChannelSize
	}

	tasks := make(chan Task, channelSize)
	pool := &TaskPool{
		numWorkers: numWorkers,
		tasks:      tasks,
		wg:         &sync.WaitGroup{},

		stop: sync.Once{},

		quit: make(chan struct{}),
	}
	pool.startWorkers()

	return pool, nil
}

func (p *TaskPool) Stop(wait bool) {
	p.stopped = true
	if wait {
		p.wg.Wait()
	}
	p.stop.Do(func() {
		close(p.quit)
	})
}

func (p *TaskPool) AddTask(t Task) {
	if p.stopped {
		return
	}
	p.wg.Add(1)
	go func() {
		select {
		case p.tasks <- t:
		case <-p.quit:
		}
	}()
}

func (p *TaskPool) startWorkers() {
	for i := 0; i < p.numWorkers; i++ {
		go func(workerNum int) {
			for {
				select {
				case <-p.quit:
					return
				case task, ok := <-p.tasks:
					if !ok {
						return
					}

					task.Execute()
					task.OnDone()
					p.wg.Done()
				}
			}
		}(i)
	}
}
