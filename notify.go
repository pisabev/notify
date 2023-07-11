package notify

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type UrlRequestTask struct {
	url               string
	message           string
	timeout           time.Duration
	successStatusCode int
	wg                *sync.WaitGroup

	mFailure     *sync.Mutex
	FailureError error
}

func (t *UrlRequestTask) Execute() error {
	if t.wg != nil {
		defer t.wg.Done()
	}

	r, err := http.NewRequest("POST", t.url, bytes.NewBuffer([]byte(t.message)))
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}

	r.Header.Add("Content-Type", "text/plain")
	client := &http.Client{Timeout: t.timeout}
	res, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	defer res.Body.Close()
	if res.StatusCode != t.successStatusCode {
		return fmt.Errorf("status code: %v", res.StatusCode)
	}

	return nil
}

func (t *UrlRequestTask) OnDone() {
	//t.done <- true
}

func (t *UrlRequestTask) OnFailure(err error) {
	t.mFailure.Lock()
	defer t.mFailure.Unlock()

	t.FailureError = err
}

type MessageSender struct {
	url               string
	successStatusCode int
	wg                *sync.WaitGroup
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
	return &MessageSender{url: url, successStatusCode: sc, wg: &sync.WaitGroup{}, pool: pool}, nil
}

func (m *MessageSender) Send(message string, timeout time.Duration) *UrlRequestTask {
	task := &UrlRequestTask{
		url:               m.url,
		message:           message,
		timeout:           timeout,
		successStatusCode: m.successStatusCode,
		wg:                m.wg,
		mFailure:          &sync.Mutex{},
	}
	m.wg.Add(1)
	m.pool.AddTask(task)
	return task
}

func (m *MessageSender) StopAndWait() {
	m.pool.Stop()
	m.wg.Wait()
}
