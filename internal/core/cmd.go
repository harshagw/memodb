package core

import "fmt"

type Cmd string

const (
	CmdSet Cmd = "SET"
	CmdGet Cmd = "GET"
	CmdDel Cmd = "DEL"
)

func (c Cmd) String() string {
	return string(c)
}

func StringToCmd(s string) (Cmd, error) {
	switch s {
	case "SET":
		return CmdSet, nil
	case "GET":
		return CmdGet, nil
	case "DEL":
		return CmdDel, nil
	default:
		return "", fmt.Errorf("unrecognized command: %s", s)
	}
}

type Command struct {
	Cmd  Cmd
	Args []string
}

type Commands []*Command
