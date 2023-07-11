package notify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageSender_NewMessageSender(t *testing.T) {
	_, err := NewMessageSender("", 0, 10)
	assert.NoError(t, err)
}

func TestMessageSender_Send(t *testing.T) {
	ms, err := NewMessageSender("", 0, 10)
	assert.NoError(t, err)

	var tasks []*UrlRequestTask

	tasks = append(tasks, ms.Send("test", 0))

	ms.wg.Wait()

	for _, task := range tasks {
		assert.NoError(t, task.failureError)
	}
}
