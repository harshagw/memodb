package main

import (
	"log"
	"net"
)

func main() {

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	_, err = conn.Write([]byte("Hello, server!\r\n"))
	if err != nil {
		log.Fatal(err)
	}

}
