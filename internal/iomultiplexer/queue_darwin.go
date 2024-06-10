package iomultiplexer

import (
	"fmt"
	"syscall"
	"time"
)

type KQueue struct {
	fd       int
	kqevents []syscall.Kevent_t
	events   []Event
}

func New(maxClients int) (*KQueue, error) {
	if maxClients < 0 {
		return nil, ErrInvalidMaxClients
	}

	fd, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}

	return &KQueue{
		fd:       fd,
		kqevents: make([]syscall.Kevent_t, maxClients),
		events:   make([]Event, maxClients),
	}, nil
}

func (kq *KQueue) Subscribe(event Event) error {
	if subscribed, err := syscall.Kevent(kq.fd, []syscall.Kevent_t{event.toNative(syscall.EV_ADD)}, nil, nil); err != nil || subscribed == -1 {
		return fmt.Errorf("kqueue subscribe: %w", err)
	}
	return nil
}

func (kq *KQueue) Poll(timeout time.Duration) ([]Event, error) {

	nEvents, err := syscall.Kevent(kq.fd, nil, kq.kqevents, newTime(timeout))
	if err != nil {
		return nil, fmt.Errorf("kqueue poll: %w", err)
	}

	for i := 0; i < nEvents; i++ {
		kq.events[i] = newEvent(kq.kqevents[i])
	}

	return kq.events[:nEvents], nil
}

func (kq *KQueue) Close() error {
	return syscall.Close(kq.fd)
}
