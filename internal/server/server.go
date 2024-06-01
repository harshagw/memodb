package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/harshagw/memodb/internal/core"
)

const (
	ServerStatus_WAITING       int32 = 1 << 1
	ServerStatus_BUSY          int32 = 1 << 2
	ServerStatus_SHUTTING_DOWN int32 = 1 << 3
)

type Server struct {
	wg       *sync.WaitGroup
	status   int32
	mu       *sync.Mutex
	clients  map[*net.Conn]*core.Client
	listener net.Listener
}

func NewServer(host string, port int, wg *sync.WaitGroup) (*Server, error) {

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	log.Println("Server started at", listener.Addr())

	return &Server{
		listener: listener,
		status:   ServerStatus_WAITING,
		wg:       wg,
		clients:  make(map[*net.Conn]*core.Client),
		mu:       &sync.Mutex{},
	}, nil
}

func (s *Server) Run() error {
	defer s.wg.Done()
	defer func() {
		atomic.StoreInt32(&s.status, ServerStatus_SHUTTING_DOWN)
	}()

	atomic.StoreInt32(&s.status, ServerStatus_BUSY)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}

		go s.handleConnection(conn)
	}

}

func (s *Server) handleConnection(conn net.Conn) {
	if conn == nil {
		log.Printf("Connection is nil")
		return
	}

	defer func() {
		log.Println("Closing connection with", conn.RemoteAddr(), "...")

		s.mu.Lock()
		delete(s.clients, &conn)
		s.mu.Unlock()

		conn.Close()
	}()

	client := core.NewClient(&conn)

	s.mu.Lock()
	s.clients[&conn] = client
	s.mu.Unlock()

	parser := core.NewParser()

	currentTime := time.Now()
	log.Printf("New connection from %s at %s", conn.RemoteAddr(), currentTime.Format("2006-01-02 15:04:05"))

	// conn.SetReadDeadline(currentTime.Add(30 * time.Second))

	for {
		n, err := conn.Read(parser.Tbuf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Read error: %v", err)
			}
			break
		}

		data := parser.Tbuf[:n]
		log.Printf("Received: %s", string(data))
		// log.Printf("Buffer: %v", data)
		parser.Write(data)

		commands, err := parser.GetCommands()
		if err != nil {
			log.Printf("Error parsing commands: %v", err)
			break
		}

		if commands == nil {
			continue
		}

		for _, command := range *commands {
			result, err := core.Execute(command, client)
			if err != nil {
				if errors.Is(err, core.ErrCommandNotExists) {
					log.Printf("Command not exists: %v", err)
					break
				}

				log.Printf("Error executing command: %v", err)
				break
			}

			if result != nil {
				_, err = conn.Write(result)
				if err != nil {
					log.Printf("Error writing response: %v", err)
					break
				}
			}
		}
	}
}

func (s *Server) WaitForShutdown(ctx context.Context) {
	defer s.wg.Done()
	<-ctx.Done()

	s.listener.Close()

	atomic.StoreInt32(&s.status, ServerStatus_SHUTTING_DOWN)

	// persist the data

	os.Exit(0)
}
