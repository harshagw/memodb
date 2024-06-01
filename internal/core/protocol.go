package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

func readBulkString(buf *bytes.Buffer) (string, error) {
	length, err := readNum(buf)
	if err != nil {
		return "", err
	}

	str := make([]byte, length)
	_, err = buf.Read(str)
	if err != nil {
		return "", err
	}

	crlf := make([]byte, 2)
	_, err = buf.Read(crlf)
	if err != nil {
		return "", err
	}

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

func readArray(buf *bytes.Buffer) ([]interface{}, error) {
	len, err := readNum(buf)
	if err != nil {
		return nil, err
	}

	var values = make([]interface{}, len)
	for i := int64(0); i < len; i++ {
		val, err := getValue(buf)
		if err != nil {
			return nil, err
		}

		values[i] = val
	}

	return values, nil
}

func getValue(buf *bytes.Buffer) (interface{}, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}

	switch b {
	case '+':
		return readSimpleString(buf)
	case '-':
		return readError(buf)
	case ':':
		return readInt(buf)
	case '*':
		return readArray(buf)
	case '$':
		return readBulkString(buf)
	default:
		return nil, ErrInvalidProtocol
	}

}

func decode(data []byte) ([]interface{}, error) {
	var values []interface{} = make([]interface{}, 0)

	buf := bytes.NewBuffer(data)

	for {
		value, err := getValue(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		values = append(values, value)
		if len(data) == 0 {
			break
		}
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
