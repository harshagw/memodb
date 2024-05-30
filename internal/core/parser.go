package core

import (
	"bytes"
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

func (p *Parser) GetCommand() ([]*Command, error) {

	if bytes.Contains(p.Tbuf, []byte{'\r', '\n'}) {
		parts := bytes.SplitN(p.buf.Bytes(), []byte{'\r', '\n'}, 2)
		log.Printf("Whole command: %s", string(parts[0]))
		p.buf.Reset()
		p.buf.Write(parts[1])
	}

	return []*Command{}, nil
}
