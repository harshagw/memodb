package iomultiplexer

import (
	"errors"
	"time"
)

type Operations uint32

const (
	OP_READ  Operations = 1
	OP_WRITE Operations = 2
)

type Event struct {
	Fd int
	Op Operations
}

var (
	ErrInvalidMaxClients = errors.New("maxClients should be greater than 0")
)

type IOMultiplexer interface {
	Subscribe(event Event) error
	Poll(timeout time.Duration) ([]Event, error)
	Close() error
}
