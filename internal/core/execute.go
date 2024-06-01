package core

import (
	"errors"
	"log"
)

var (
	ErrCommandNotExists = errors.New("command not exists")
	ErrorInvalidArgs    = errors.New("invalid number of arguments")
)

var (
	RESP_OK  = []byte("+OK\r\n")
	RESP_NIL = []byte("$-1\r\n")
)

func Execute(cmd *Command, c *Client) ([]byte, error) {

	log.Println("Executing command: ", cmd.Cmd, " with args: ", cmd.Args)

	switch cmd.Cmd {
	case "SET":
		if len(cmd.Args) != 2 {
			return nil, ErrorInvalidArgs
		}

		key, value := cmd.Args[0], cmd.Args[1]

		obj := newObj(value)

		set(key, obj)

		return RESP_OK, nil
	case "GET":
		if len(cmd.Args) != 1 {
			return nil, ErrorInvalidArgs
		}

		obj := get(cmd.Args[0])
		if obj == nil {
			return nil, nil
		}

		return encode(obj.Value, false), nil
	case "DEL":
		if len(cmd.Args) == 0 {
			return nil, ErrorInvalidArgs
		}

		var cnt int = 0

		for _, key := range cmd.Args {
			if ok := del(key); ok {
				cnt++
			}
		}

		return encode(cnt, false), nil
	}

	return nil, ErrCommandNotExists
}
