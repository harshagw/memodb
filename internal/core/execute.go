package core

import (
	"bytes"
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

func ExecuteCommands(cmds Commands, c Client) {
	var response []byte
	buf := bytes.NewBuffer(response)

	for _, command := range cmds {
		result, err := execute(command)
		if err != nil {
			if errors.Is(err, ErrCommandNotExists) {
				log.Printf("Command not exists: %v", err)
				break
			}

			log.Printf("Error executing command: %v", err)
			break
		}

		buf.Write(result)
	}

	_, err := c.Write(buf.Bytes())
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func DecodeCommands(values []interface{}) (Commands, error) {
	commands := make(Commands, 0)

	for _, val := range values {

		v, ok := val.([]interface{})
		if !ok {
			v = []interface{}{val}
		}

		tokens := make([]string, len(v))
		for i := range v {
			tokens[i] = v[i].(string)
		}

		cmdType, err := StringToCmd(tokens[0])
		if err != nil {
			return nil, err
		}

		cmd := Command{
			Cmd:  cmdType,
			Args: tokens[1:],
		}

		commands = append(commands, &cmd)
	}

	return commands, nil
}

func execute(cmd *Command) ([]byte, error) {

	log.Println("Executing command: ", cmd.Cmd, " with args: ", cmd.Args)

	switch cmd.Cmd {
	case CmdSet:
		if len(cmd.Args) != 2 {
			return nil, ErrorInvalidArgs
		}

		key, value := cmd.Args[0], cmd.Args[1]

		obj := newObj(value)

		set(key, obj)

		return RESP_OK, nil
	case CmdGet:
		if len(cmd.Args) != 1 {
			return nil, ErrorInvalidArgs
		}

		obj := get(cmd.Args[0])
		if obj == nil {
			return nil, nil
		}

		return encode(obj.Value, false), nil
	case CmdDel:
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
