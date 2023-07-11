package notify

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMessageSender_NewMessageSender(t *testing.T) {
	_, err := NewMessageSender("", 0, 10)
	assert.NoError(t, err)
}

func TestMessageSender_Send(t *testing.T) {
	message := "test"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, message, string(b))
	}))
	defer ts.Close()

	ms, err := NewMessageSender(ts.URL, http.StatusOK, 10)
	assert.NoError(t, err)

	var tasks []*UrlRequestTask

	tasks = append(tasks, ms.Send(message, 0))

	ms.wg.Wait()

	for _, task := range tasks {
		assert.NoError(t, task.failureError)
	}
}

func TestMessageSender_SendTimeout(t *testing.T) {
	message := "test"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * 2)
		b, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, message, string(b))
	}))
	defer ts.Close()

	ms, err := NewMessageSender(ts.URL, http.StatusOK, 10)
	assert.NoError(t, err)

	var tasks []*UrlRequestTask

	tasks = append(tasks, ms.Send(message, time.Second*1))

	ms.wg.Wait()

	for _, task := range tasks {
		assert.Error(t, task.failureError)
	}
}
