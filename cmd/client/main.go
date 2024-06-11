package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
)

func main() {

	conn, err := net.Dial("tcp", "localhost:8090")
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

	// log.Println("running command 1")

	// _, err = conn.Write([]byte("*"))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("running command 2")

	// _, err = conn.Write([]byte("*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nAlice\r\n*3\r\n$3\r\nSET\r\n$3\r\nage\r\n$2\r\n30\r\n*2\r\n$3\r\nGET\r\n$4\r\nname\r\n"))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	log.Println("running command 3")

	_, err = conn.Write([]byte("*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$100\r\ny4bxFp8jCrH8x8u8mD0dBC9eYGvr6t1XEWsIkCZeFyz7jKtbwzU0nQlngODvudSt0vBpNZIDvVNMeDgnO0vOcRXQX1LG08NJmuYB\r\n"))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("running command 4")
	_, err = conn.Write([]byte("*2\r\n$3\r\nGET\r\n$4\r\nname\r\n"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	<-ctx.Done()
}
