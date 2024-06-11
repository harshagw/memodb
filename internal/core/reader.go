package core

import (
	"bufio"
	"io"
	"log"
	"strings"

	"github.com/harshagw/memodb/internal/config"
)

const (
	RESPTYPE_ARRAY         byte = '*'
	RESPTYPE_SIMPLE_STRING byte = '+'
	RESPTYPE_BULK_STRING   byte = '$'
	RESPTYPE_INTEGER       byte = ':'
	RESPTYPE_ERROR         byte = '-'
)

type Reader struct {
	r             *bufio.Reader
	temp          []byte
	stringBuilder strings.Builder
}

func NewReader(c io.Reader) *Reader {
	return &Reader{
		r:             bufio.NewReader(c),
		temp:          make([]byte, config.TEMP_BUFFER_SIZE),
		stringBuilder: strings.Builder{},
	}
}

func (m *Reader) ReadCommand() (*Command, error) {
	log.Println("Reading command")

	value, err := m.readValue()
	if err != nil {
		return nil, err
	}

	v, ok := value.([]interface{})
	if !ok {
		v = []interface{}{value}
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

	return &cmd, nil
}

func (m *Reader) readValue() (interface{}, error) {
	log.Println("Waiting to read first value")
	b, err := m.r.ReadByte()
	if err != nil {
		return nil, err
	}

	log.Println("Read first value : ", string(b))

	switch b {
	case RESPTYPE_SIMPLE_STRING:
		return m.readSimpleString()
	case RESPTYPE_BULK_STRING:
		return m.readBulkString()
	case RESPTYPE_INTEGER:
		return m.readInteger()
	case RESPTYPE_ERROR:
		return m.readError()
	case RESPTYPE_ARRAY:
		return m.readArray()
	default:
		return nil, ErrInvalidProtocol
	}

}

func (m *Reader) readArray() ([]interface{}, error) {
	length, err := m.readIntUntilCRLF()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, length)

	for i := 0; i < length; i++ {
		values[i], err = m.readValue()
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}

func (m *Reader) readInteger() (int, error) {
	num, err := m.readIntUntilCRLF()
	if err != nil {
		return 0, err
	}

	return num, nil
}

func (m *Reader) readSimpleString() (string, error) {
	m.stringBuilder.Reset()
	if err := m.readBytesUntilCRLF(&m.stringBuilder); err != nil {
		return "", err
	}
	return m.stringBuilder.String(), nil
}

func (m *Reader) readError() (string, error) {
	return m.readSimpleString()
}

func (m *Reader) readBulkString() (string, error) {
	length, err := m.readIntUntilCRLF()
	if err != nil {
		return "", err
	}

	m.stringBuilder.Reset()
	m.stringBuilder.Grow(length)
	if err := m.write(&m.stringBuilder, length); err != nil {
		return "", err
	}

	s := m.stringBuilder.String()
	if err := m.skipCRLF(); err != nil {
		return "", err
	}

	return s, nil
}

func (m *Reader) readIntUntilCRLF() (int, error) {
	b, err := m.r.ReadByte()
	if err != nil {
		return 0, err
	}

	if b < '0' || b > '9' {
		return 0, ErrInvalidProtocol
	}
	result := int(b - '0')

	for b != '\r' {
		b, err = m.r.ReadByte()
		if err != nil {
			return 0, err
		}

		if b != '\r' {
			if b < '0' || b > '9' {
				return 0, ErrInvalidProtocol
			}
			result = result*10 + int(b-'0')
		}
	}

	b, err = m.r.ReadByte()
	if err != nil {
		return 0, err
	}
	if b != '\n' {
		return 0, ErrInvalidProtocol
	}

	return result, nil
}

func (m *Reader) readBytesUntilCRLF(w io.Writer) error {
	b, err := m.r.ReadByte()
	if err != nil {
		return err
	}

	for b != '\r' {
		if _, err := w.Write([]byte{b}); err != nil {
			return err
		}

		b, err = m.r.ReadByte()
		if err != nil {
			return err
		}
	}

	b, err = m.r.ReadByte()
	if err != nil {
		return err
	}
	if b != '\n' {
		return ErrInvalidProtocol
	}

	return nil
}

func (m *Reader) skipCRLF() error {
	b, err := m.r.ReadByte()
	if err != nil {
		return err
	}
	if b != '\r' {
		return ErrInvalidProtocol
	}

	b, err = m.r.ReadByte()
	if err != nil {
		return err
	}
	if b != '\n' {
		return ErrInvalidProtocol
	}

	return nil
}

func (m *Reader) write(w io.Writer, length int) error {
	remaining := length
	for remaining > 0 {
		n := remaining
		if n > len(m.temp) {
			n = len(m.temp)
		}

		if _, err := io.ReadFull(m.r, m.temp[:n]); err != nil {
			return err
		}
		remaining -= n

		if _, err := w.Write(m.temp[:n]); err != nil {
			return err
		}
	}

	return nil
}
