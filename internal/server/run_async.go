package server

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"sync/atomic"
	"syscall"

	"github.com/harshagw/memodb/internal/config"
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

				cmd, err := handleAsyncConnection(comm)

				if err != nil {
					syscall.Close(event.Fd)
					delete(s.clients, event.Fd)
					continue
				}

				core.ExecuteCommands(cmd, comm)
			}
		}
	}

	return nil
}

func handleAsyncConnection(c io.ReadWriter) (core.Commands, error) {
	var b []byte
	var buf *bytes.Buffer = bytes.NewBuffer(b)

	tbuf := make([]byte, config.TEMP_BUFFER_SIZE)

	totalN := 0

	for {
		n, err := c.Read(tbuf)
		if n <= 0 {
			log.Println("No data read")
			break
		}

		totalN += n

		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		data := tbuf[:n]
		log.Println("read the data : ", string(data))

		_, err = buf.Write(data)
		if err != nil {
			log.Println("Error writing to buffer: ", err)
			return nil, err
		}

		if buf.Len() > config.MAX_BUFFER_SIZE {
			log.Println("Max buffer size reached")
			return nil, errors.New("max buffer size reached")
		}
	}

	if totalN == 0 {
		return nil, errors.New("no data read")
	}

	var reader = core.NewReader(buf)

	cmds := make(core.Commands, 0)

	for {
		cmd, err := reader.ReadCommand()
		if err != nil {
			log.Printf("Error reading command: %v", err)
			break
		}

		cmds = append(cmds, cmd)
	}

	return cmds, nil
}
