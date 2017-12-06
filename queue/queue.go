package queue

import (
	"context"
	"errors"
)

var (
	// ErrCancel indicates the task was cancelled.
	ErrCancel = errors.New("queue: task cancelled")

	// ErrNotFound indicates the task was not found in the queue.
	ErrNotFound = errors.New("queue: task not found")
)

// 表示队列中的一项任务
type Task struct {
	ID string `json:"id,omitempty"`
	Data []byte `json:"data"`
	Labels map[string]string `json:"labels,omitempty"`
}

// InfoT 提供队列运行时信息
type InfoT struct {
	Pending []*Task `json:"pending"`
	Running []*Task `json:"running"`
	Stats   struct {
		Workers  int `json:"worker_count"`
		Pending  int `json:"pending_count"`
		Running  int `json:"running_count"`
		Complete int `json:"completed_count"`
	} `json:"stats"`
}

// 任务过滤函数， 如果返回 false 则该 Task 就被忽略
type Filter func(*Task) bool

type Queue interface {
	// 将一个 task 添加到队列末尾
	Push(c context.Context, task *Task) error

	// 从队列头取一个任务，并删除
	Poll(c context.Context, f Filter) (*Task, error)

	// Extend extends the deadline for a task.
	Extend(c context.Context, id string) error

	// Done signals the task is complete.
	Done(c context.Context, id string) error

	// Error signals the task is complete with errors.
	Error(c context.Context, id string, err error) error

	// Evict removes a pending task from the queue.
	Evict(c context.Context, id string) error

	// Wait waits until the task is complete.
	Wait(c context.Context, id string) error

	// Info returns internal queue information.
	Info(c context.Context) InfoT
}