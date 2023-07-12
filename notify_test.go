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
	type tc struct {
		Name          string
		Messages      []string
		Workers       int
		ResponseCode  int
		TimeOut       time.Duration
		ResponseDelay time.Duration
	}
	cases := []tc{
		{
			Name:          "test send message",
			Messages:      []string{"test1", "test2", "test3", "test4"},
			Workers:       4,
			ResponseCode:  http.StatusOK,
			TimeOut:       time.Millisecond * 1,
			ResponseDelay: time.Millisecond * 0,
		},
		{
			Name:          "test send message with timeout",
			Messages:      []string{"test1", "test2", "test3", "test4"},
			Workers:       4,
			ResponseCode:  http.StatusOK,
			TimeOut:       time.Millisecond * 1,
			ResponseDelay: time.Millisecond * 2,
		},
		{
			Name:          "test send message with status created",
			Messages:      []string{"test1", "test2", "test3", "test4"},
			Workers:       1,
			ResponseCode:  http.StatusCreated,
			TimeOut:       time.Millisecond * 1,
			ResponseDelay: time.Millisecond * 0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				w.WriteHeader(tc.ResponseCode)
				_, err = w.Write(b)
				assert.NoError(t, err)
				time.Sleep(tc.ResponseDelay)
			}))
			defer ts.Close()

			ms, err := NewMessageSender(ts.URL, tc.ResponseCode, tc.Workers)
			assert.NoError(t, err)

			var tasks []*UrlRequestTask
			for _, m := range tc.Messages {
				func(sendMessage string, response int, delay, timeout time.Duration) {
					tasks = append(tasks, ms.Send(sendMessage, timeout, func(task *UrlRequestTask) {
						// If we hit timeout
						if delay > timeout {
							assert.Error(t, task.RError)
						} else { // Normal execution
							assert.Equal(t, tc.ResponseCode, task.RCode)
							assert.Equal(t, sendMessage, task.RMessage)
						}
					}))
				}(m, tc.ResponseCode, tc.ResponseDelay, tc.TimeOut)
			}

			// Wait to finish
			ms.Stop(true)
		})
	}
}
