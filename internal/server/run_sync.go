package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/harshagw/memodb/internal/core"
)

func (s *Server) runSync() error {

	return errors.New("not implemented yet")

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.Host, s.config.Port))
	if err != nil {
		return err
	}

	log.Println("Server started at", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			listener.Close()
			return err
		}

		go s.handleSyncConnection(conn)
	}
}

func (s *Server) handleSyncConnection(conn net.Conn) {
	if conn == nil {
		log.Printf("Connection is nil")
		return
	}

	file, err := conn.(*net.TCPConn).File()
	if err != nil {
		log.Printf("Error getting file descriptor: %v", err)
		return
	}
	id := int(file.Fd())

	defer func() {
		log.Println("Closing connection with", conn.RemoteAddr(), "...")

		s.mu.Lock()
		delete(s.clients, id)
		s.mu.Unlock()

		conn.Close()
	}()

	client := core.NewSyncClient(conn)

	s.mu.Lock()
	s.clients[id] = client
	s.mu.Unlock()

	currentTime := time.Now()
	log.Printf("New connection from %s at %s", conn.RemoteAddr(), currentTime.Format("2006-01-02 15:04:05"))
}
