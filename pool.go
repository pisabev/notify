package notify

import (
	"fmt"
	"log"
	"sync"
)

var ErrNoWorkers = fmt.Errorf("attempting to create worker pool with less than 1 worker")
var ErrNegativeChannelSize = fmt.Errorf("attempting to create worker pool with a negative channel size")

type Pool interface {
	// Stop stops the workerpool, tears down any required resources,
	// and should only be called once
	Stop()
	// AddTask adds a task for the worker pool to process. It is only valid after
	// Start() has been called and before Stop() has been called.
	AddTask(Task)
}

type Task interface {
	// Execute performs the work
	Execute() error
	// OnFailure handles any error returned from Execute()
	OnFailure(error)
}

type TaskPool struct {
	numWorkers int
	tasks      chan Task

	// Ensure the pool can only be stopped once
	stop sync.Once

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

		stop: sync.Once{},

		quit: make(chan struct{}),
	}
	pool.startWorkers()

	return pool, nil
}

func (p *TaskPool) Stop() {
	p.stop.Do(func() {
		log.Print("stopping simple worker pool")
		close(p.quit)
	})
}

func (p *TaskPool) AddTask(t Task) {
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
			log.Printf("starting worker %d", workerNum)

			for {
				select {
				case <-p.quit:
					log.Printf("stopping worker %d with quit channel\n", workerNum)
					return
				case task, ok := <-p.tasks:
					if !ok {
						log.Printf("stopping worker %d with closed tasks channel\n", workerNum)
						return
					}

					if err := task.Execute(); err != nil {
						task.OnFailure(err)
					}
				}
			}
		}(i)
	}
}
