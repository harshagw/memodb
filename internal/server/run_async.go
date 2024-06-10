package server

import (
	"io"
	"log"
	"net"
	"sync/atomic"
	"syscall"

	"github.com/harshagw/memodb/internal/core"
	"github.com/harshagw/memodb/internal/iomultiplexer"
)

func (s *Server) runAsync() error {

	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(serverFD)

	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}

	ip4 := net.ParseIP(s.config.Host)

	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: s.config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		return err
	}

	if err = syscall.Listen(serverFD, s.config.MaxClients); err != nil {
		return err
	}

	var multiplexer iomultiplexer.IOMultiplexer
	multiplexer, err = iomultiplexer.New(s.config.MaxClients)
	if err != nil {
		log.Fatal(err)
	}
	defer multiplexer.Close()

	if err := multiplexer.Subscribe(iomultiplexer.Event{
		Fd: serverFD,
		Op: iomultiplexer.OP_READ,
	}); err != nil {
		return err
	}

	for atomic.LoadInt32(&s.status) != ServerStatus_SHUTTING_DOWN {

		events, err := multiplexer.Poll(-1)
		if err != nil {
			continue
		}

		for _, event := range events {
			if event.Fd == serverFD {
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("err", err)
					continue
				}

				syscall.SetNonblock(fd, true)

				s.clients[fd] = core.NewAsyncClient(fd)

				if err := multiplexer.Subscribe(iomultiplexer.Event{
					Fd: fd,
					Op: iomultiplexer.OP_READ,
				}); err != nil {
					return err
				}

			} else {
				comm := s.clients[event.Fd]
				if comm == nil {
					continue
				}

				cmds, err := handleAsyncConnection(comm)

				if err != nil {
					syscall.Close(event.Fd)
					delete(s.clients, event.Fd)
					continue
				}

				core.ExecuteCommands(cmds, comm)
			}
		}
	}

	return nil
}

func handleAsyncConnection(c io.ReadWriter) (core.Commands, error) {
	var parser = core.NewParser(c)

	values, err := parser.GetMultiple()
	if err != nil {
		log.Println("Error getting values: ", err)
		return nil, err
	}

	cmds, err := core.DecodeCommands(values)
	if err != nil {
		log.Println("Error getting commands: ", err)
		return nil, err
	}

	return cmds, nil
}
