package notify

import (
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

	p.Stop(false)
	p.Stop(false)
}

func TestWorkerPool_Work(t *testing.T) {
	var tasks []*testTask

	for i := 0; i < 100; i++ {
		tasks = append(tasks, newTestTask())
	}

	p, err := NewTaskPool(10, 2)
	assert.NoError(t, err)

	for _, j := range tasks {
		p.AddTask(j)
	}

	p.Stop(true)

	for _, task := range tasks {
		assert.True(t, task.executed)
	}
}

type testTask struct {
	executed bool
}

func newTestTask() *testTask {
	return &testTask{}
}

func (t *testTask) Execute() {
	//time.Sleep(time.Second * 1)
}

func (t *testTask) OnDone() {
	t.executed = true
}
