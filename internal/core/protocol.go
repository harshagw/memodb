package core

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrInvalidProtocol = errors.New("invalid protocol")
)

func readString(buf *bytes.Buffer) (string, error) {
	str, err := buf.ReadString('\r')
	if err != nil {
		return "", err
	}

	buf.ReadByte()
	return str[:len(str)-1], nil
}

func readNum(buf *bytes.Buffer) (int64, error) {
	str, err := readString(buf)
	if err != nil {
		return 0, err
	}

	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}

	return num, nil
}

func readSimpleString(buf *bytes.Buffer) (string, error) {
	return readString(buf)
}

func readError(buf *bytes.Buffer) (string, error) {
	return readString(buf)
}

func readBulkString(buf *bytes.Buffer, p *Parser) (string, error) {
	length, err := readNum(buf)
	if err != nil {
		return "", err
	}

	var bytesRem int64 = length + 2
	bytesRem = bytesRem - int64(buf.Len())
	for bytesRem > 0 {
		tbuf := make([]byte, bytesRem)
		n, err := p.Read(tbuf)
		if err != nil {
			return "", nil
		}
		buf.Write(tbuf[:n])
		bytesRem = bytesRem - int64(n)
	}

	str := make([]byte, length)
	_, err = buf.Read(str)
	if err != nil {
		return "", err
	}

	buf.ReadByte()
	buf.ReadByte()

	return string(str), nil
}

func readInt(buf *bytes.Buffer) (int64, error) {
	sign, err := buf.ReadByte()
	if err != nil {
		return 0, err
	}

	if sign == '+' || sign == '-' {
		num, err := readNum(buf)
		if err != nil {
			return 0, err
		}

		if sign == '-' {
			num = -num
		}

		return num, nil
	} else {
		buf.Write([]byte{sign})

		num, err := readNum(buf)
		if err != nil {
			return 0, err
		}

		return num, nil
	}

}

func readArray(buf *bytes.Buffer, p *Parser) ([]interface{}, error) {
	len, err := readNum(buf)
	if err != nil {
		return nil, err
	}

	var values = make([]interface{}, len)
	for i := int64(0); i < len; i++ {
		val, err := p.decodeOne()
		if err != nil {
			return nil, err
		}

		values[i] = val
	}

	return values, nil
}

func encode(data interface{}, isSimple bool) []byte {
	switch v := data.(type) {
	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	case int:
		if v < 0 {
			return []byte(fmt.Sprintf(":-%d\r\n", v))
		}
		return []byte(fmt.Sprintf(":%d\r\n", v))
	case int64:
		return []byte(fmt.Sprintf(":%d\r\n", v))
	case []string:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, v := range data.([]string) {
			buf.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(v), buf.Bytes()))
	case error:
		return []byte(fmt.Sprintf("-%s\r\n", v))
	default:
		return RESP_NIL
	}
}
