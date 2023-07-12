package notify

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type UrlRequestTask struct {
	url               string
	message           string
	timeout           time.Duration
	successStatusCode int

	// Set on response - use in doneFunc
	RError   error
	RCode    int
	RMessage string

	doneFunc func(*UrlRequestTask)
}

func (t *UrlRequestTask) Execute() {
	r, err := http.NewRequest("POST", t.url, bytes.NewBuffer([]byte(t.message)))
	if err != nil {
		t.RError = fmt.Errorf("request: %w", err)
		return
	}

	r.Header.Add("Content-Type", "text/plain")
	client := &http.Client{Timeout: t.timeout}
	res, err := client.Do(r)
	if err != nil {
		t.RError = fmt.Errorf("do request: %w", err)
		return
	}

	defer res.Body.Close()
	if res.StatusCode != t.successStatusCode {
		t.RError = fmt.Errorf("status code: %v", res.StatusCode)
		return
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.RError = fmt.Errorf("response read: %w", err)
		return
	}

	t.RMessage = string(b)
	t.RCode = res.StatusCode
}

func (t *UrlRequestTask) OnDone() {
	if t.doneFunc != nil {
		t.doneFunc(t)
	}
}

type MessageSender struct {
	url               string
	successStatusCode int
	pool              Pool
}

// NewMessageSender construct a new *MessageSender
func NewMessageSender(url string, successStatusCode int, workers int) (*MessageSender, error) {
	pool, err := NewTaskPool(workers, 1)
	if err != nil {
		return nil, err
	}
	sc := successStatusCode
	if sc == 0 {
		sc = http.StatusOK
	}
	return &MessageSender{url: url, successStatusCode: sc, pool: pool}, nil
}

func (m *MessageSender) Send(message string, timeout time.Duration, doneFunc func(*UrlRequestTask)) *UrlRequestTask {
	task := &UrlRequestTask{
		url:               m.url,
		message:           message,
		timeout:           timeout,
		successStatusCode: m.successStatusCode,
		doneFunc:          doneFunc,
	}
	m.pool.AddTask(task)
	return task
}

func (m *MessageSender) Stop(wait bool) {
	m.pool.Stop(wait)
}
