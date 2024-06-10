package iomultiplexer

import (
	"syscall"
	"time"
)

// newTime converts the given time.Duration to Darwin's timespec struct
func newTime(t time.Duration) *syscall.Timespec {
	if t < 0 {
		return nil
	}

	return &syscall.Timespec{
		Nsec: int64(t),
	}
}

func newEvent(kEvent syscall.Kevent_t) Event {
	return Event{
		Fd: int(kEvent.Ident),
		Op: newOperations(kEvent.Filter),
	}
}

func (e Event) toNative(flags uint16) syscall.Kevent_t {
	return syscall.Kevent_t{
		Ident:  uint64(e.Fd),
		Filter: e.Op.toNative(),
		Flags:  flags,
	}
}

func newOperations(filter int16) Operations {
	op := Operations(0)

	if filter&syscall.EVFILT_READ != 0 {
		op |= OP_READ
	}
	if filter&syscall.EVFILT_WRITE != 0 {
		op |= OP_WRITE
	}

	return op
}

func (op Operations) toNative() int16 {
	native := int16(0)

	if op&OP_READ != 0 {
		native |= syscall.EVFILT_READ
	}
	if op&OP_WRITE != 0 {
		native |= syscall.EVFILT_WRITE
	}

	return native
}
