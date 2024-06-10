package server

import (
	"context"
	"os"
	"sync"
	"sync/atomic"

	"github.com/harshagw/memodb/internal/config"
	"github.com/harshagw/memodb/internal/core"
)

const (
	ServerStatus_WAITING       int32 = 1 << 1
	ServerStatus_BUSY          int32 = 1 << 2
	ServerStatus_SHUTTING_DOWN int32 = 1 << 3
)

type Server struct {
	config  *config.Config
	wg      *sync.WaitGroup
	status  int32
	mu      *sync.Mutex
	clients map[int]core.Client
}

func NewServer(config *config.Config, wg *sync.WaitGroup) (*Server, error) {
	return &Server{
		config:  config,
		status:  ServerStatus_WAITING,
		wg:      wg,
		clients: make(map[int]core.Client),
		mu:      &sync.Mutex{},
	}, nil
}

func (s *Server) Run() error {
	defer s.wg.Done()
	defer func() {
		atomic.StoreInt32(&s.status, ServerStatus_SHUTTING_DOWN)
	}()

	atomic.StoreInt32(&s.status, ServerStatus_BUSY)

	if s.config.ServerType == config.ServerTypeAsync {
		return s.runAsync() // single threaded
	}

	return s.runSync() // multiple threaded
}

func (s *Server) WaitForShutdown(ctx context.Context) {
	defer s.wg.Done()
	<-ctx.Done()

	atomic.StoreInt32(&s.status, ServerStatus_SHUTTING_DOWN)

	// persist the data

	os.Exit(0)
}
