package core

import (
	"bytes"
	"fmt"
)

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
