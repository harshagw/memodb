package core

import (
	"net"
)

type Client struct {
	conn   *net.Conn
	cqueue Commands
}

func NewClient(conn *net.Conn) *Client {
	return &Client{
		conn:   conn,
		cqueue: make(Commands, 0),
	}
}
