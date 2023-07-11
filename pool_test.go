package notify

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPool_NewTaskPool(t *testing.T) {
	_, err := NewTaskPool(0, 0)
	assert.Equal(t, ErrNoWorkers, err)

	_, err = NewTaskPool(-1, 0)
	assert.Equal(t, ErrNoWorkers, err)

	_, err = NewTaskPool(1, -1)
	assert.Equal(t, ErrNegativeChannelSize, err)

	p, err := NewTaskPool(5, 0)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	p.Stop()
	p.Stop()
}

func TestWorkerPool_Work(t *testing.T) {
	var tasks []*testTask
	wg := &sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		tasks = append(tasks, newTestTask(wg))
	}

	p, err := NewTaskPool(5, 1)
	assert.NoError(t, err)

	for _, j := range tasks {
		p.AddTask(j)
	}

	// timeout if not processed
	wg.Wait()

	for _, task := range tasks {
		assert.NoError(t, task.hitFailureCase())
	}
}

type testTask struct {
	wg           *sync.WaitGroup
	mFailure     *sync.Mutex
	done         chan bool
	failureError error
}

func newTestTask(wg *sync.WaitGroup) *testTask {
	return &testTask{
		wg:       wg,
		mFailure: &sync.Mutex{},
		done:     make(chan bool),
	}
}

func (t *testTask) Execute() error {
	if t.wg != nil {
		defer t.wg.Done()
	}

	return nil
}

func (t *testTask) OnDone() {
	t.done <- true
}

func (t *testTask) OnFailure(err error) {
	t.mFailure.Lock()
	defer t.mFailure.Unlock()

	t.failureError = err
}

func (t *testTask) hitFailureCase() error {
	t.mFailure.Lock()
	defer t.mFailure.Unlock()

	return t.failureError
}
