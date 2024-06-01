package core

import (
	"net"
)

type Client struct {
	conn *net.Conn
}

func NewClient(conn *net.Conn) *Client {
	return &Client{
		conn: conn,
	}
}
