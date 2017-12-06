package logging

import (
	"errors"
	"context"
	"io"
)

var ErrNotFound = errors.New("stream: not found")

// log entry
type Entry struct {
	ID string `json:"id,omitempty"`
	Data []byte `json:"data"`
	Tags map[string]string `json:"tags,omitempty"`
}

// callback func for handling log entries
type Handler func(...*Entry)

type Log interface {
	// open the log
	Open(c context.Context, path string) error

	// writes the entry to the log
	Write(c context.Context, path string, entry *Entry) error

	// tails the log
	Tail(c context.Context, path string, handler Handler) error

	// close the log.
	Close(c context.Context, path string) error

	// snapshots the stream to Writer w.
	Snapshot(c context.Context, path string, w io.Writer) error
}

