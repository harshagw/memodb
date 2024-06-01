package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
)

func main() {

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	go func() {
		_, err := io.Copy(os.Stdout, conn)
		if err != nil {
			log.Fatal(err)
		}
	}()

	_, err = conn.Write([]byte("*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nAlice\r\n"))
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Write([]byte("*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nAlice\r\n*3\r\n$3\r\nSET\r\n$3\r\nage\r\n$2\r\n30\r\n*2\r\n$3\r\nGET\r\n$4\r\nname\r\n"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	<-ctx.Done()

}
