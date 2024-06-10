package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"github.com/harshagw/memodb/internal/config"
	"github.com/harshagw/memodb/internal/server"
)

func main() {
	config, err := config.NewConfig(".")
	if err != nil {
		log.Fatal("cannot create config:", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)

	server, err := server.NewServer(config, &wg)
	if err != nil {
		log.Fatal(err)
	}

	go server.Run()

	go server.WaitForShutdown(ctx)

	wg.Wait()

	log.Println("Server stopped")

}
