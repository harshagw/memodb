package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"github.com/harshagw/memodb/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)

	server, err := server.NewServer("0.0.0.0", 8080, &wg)
	if err != nil {
		log.Fatal(err)
	}

	go server.Run()

	go server.WaitForShutdown(ctx)

	wg.Wait()

	log.Println("Server stopped")

}
