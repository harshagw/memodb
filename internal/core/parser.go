package core

import (
	"bytes"
	"errors"
	"log"
)

type Parser struct {
	buf  *bytes.Buffer
	Tbuf []byte
}

func NewParser() *Parser {
	var b []byte
	var buf *bytes.Buffer = bytes.NewBuffer(b)
	return &Parser{
		buf:  buf,
		Tbuf: make([]byte, 1024),
	}
}

func (p *Parser) Write(b []byte) {
	p.buf.Write(b)
}

func toArrayString(ai []interface{}) ([]string, error) {
	as := make([]string, len(ai))
	for i := range ai {
		as[i] = ai[i].(string)
	}
	return as, nil
}

func (p *Parser) GetCommands() (*Commands, error) {

	if !bytes.Contains(p.Tbuf, []byte{'\r', '\n'}) {
		return nil, errors.New("no CRLF found")
	}

	values, err := decode(p.buf.Bytes())
	if err != nil {
		log.Println("Error decoding values: ", err)
		return nil, err
	}

	p.buf.Reset()

	commands := make(Commands, 0)

	for _, val := range values {

		v, ok := val.([]interface{})
		if !ok {
			v = []interface{}{val}
		}

		tokens, err := toArrayString(v)
		if err != nil {
			return nil, err
		}

		cmd := Command{
			Cmd:  tokens[0],
			Args: tokens[1:],
		}

		commands = append(commands, &cmd)
	}

	return &commands, nil
}
