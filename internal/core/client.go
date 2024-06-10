package core

import (
	"io"
	"net"
	"syscall"
)

type Client interface {
	io.ReadWriter
}

type SyncClient struct {
	conn net.Conn
}

func NewSyncClient(conn net.Conn) *SyncClient {
	return &SyncClient{
		conn: conn,
	}
}

func (c SyncClient) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c SyncClient) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

type AsyncClient struct {
	Fd int
}

func NewAsyncClient(fd int) *AsyncClient {
	return &AsyncClient{
		Fd: fd,
	}
}

func (c AsyncClient) Write(b []byte) (int, error) {
	return syscall.Write(c.Fd, b)
}

func (c AsyncClient) Read(b []byte) (int, error) {
	return syscall.Read(c.Fd, b)
}
