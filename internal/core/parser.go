package core

import (
	"bytes"
	"errors"
	"io"
	"log"

	"github.com/harshagw/memodb/internal/config"
)

type Parser struct {
	c    io.ReadWriter
	tbuf []byte
	buf  *bytes.Buffer
}

func NewParser(c io.ReadWriter) *Parser {
	var b []byte
	var buf *bytes.Buffer = bytes.NewBuffer(b)
	return &Parser{
		tbuf: make([]byte, config.TEMP_BUFFER_SIZE),
		c:    c,
		buf:  buf,
	}
}

func (p *Parser) Write(b []byte) (int, error) {
	n, err := p.buf.Write(b)
	if err != nil {
		log.Println("Error writing to buffer: ", err)
		return 0, err
	}

	if p.buf.Len() > config.MAX_BUFFER_SIZE {
		log.Println("Max buffer size reached")
		return 0, errors.New("max buffer size reached")
	}

	return n, nil
}

func (p *Parser) Read(b []byte) (int, error) {
	return p.c.Read(b)
}

func (p *Parser) decodeOne() (interface{}, error) {
	for {
		n, err := p.c.Read(p.tbuf)
		if n <= 0 {
			log.Println("No data read")
			break
		}

		data := p.tbuf[:n]
		log.Println("read the data : ", string(data))

		_, err2 := p.Write(data)
		if err2 != nil {
			return nil, err
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if bytes.Contains(p.tbuf, []byte{'\r', '\n'}) {
			break
		}
	}

	b, err := p.buf.ReadByte()
	if err != nil {
		return nil, err
	}

	switch b {
	case '+':
		return readSimpleString(p.buf)
	case '-':
		return readError(p.buf)
	case ':':
		return readInt(p.buf)
	case '*':
		return readArray(p.buf, p)
	case '$':
		return readBulkString(p.buf, p)
	default:
		return nil, ErrInvalidProtocol
	}
}

func (p *Parser) GetMultiple() ([]interface{}, error) {
	var values []interface{} = make([]interface{}, 0)

	for {
		value, err := p.decodeOne()
		if err != nil {
			return nil, err
		}
		values = append(values, value)
		if p.buf.Len() == 0 {
			break
		}
	}

	return values, nil
}
