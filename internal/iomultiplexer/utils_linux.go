package iomultiplexer

import (
	"syscall"
	"time"
)

// newTime converts the given time.Duration to Linux's ms in int
func newTime(t time.Duration) int {
	if t < 0 {
		return -1
	}

	return int(t / time.Millisecond)
}

func newEvent(ePEvent syscall.EpollEvent) Event {
	return Event{
		Fd: int(ePEvent.Fd),
		Op: newOperations(ePEvent.Events),
	}
}

func (e Event) toNative() syscall.EpollEvent {
	return syscall.EpollEvent{
		Fd:     int32(e.Fd),
		Events: e.Op.toNative(),
	}
}

func newOperations(events uint32) Operations {
	op := Operations(0)

	if events&syscall.EPOLLIN != 0 {
		op |= OP_READ
	}
	if events&syscall.EPOLLOUT != 0 {
		op |= OP_WRITE
	}

	return op
}

func (op Operations) toNative() uint32 {
	native := uint32(0)

	if op&OP_READ != 0 {
		native |= syscall.EPOLLIN
	}
	if op&OP_WRITE != 0 {
		native |= syscall.EPOLLOUT
	}

	return native
}
