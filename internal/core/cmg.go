package core

type Command struct {
	Cmd  string
	Args []string
}

type Commands []*Command