package utils

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"strings"
)

// Base64Encode base64加密
func Base64Encode(raw []byte) []byte {
	var encoded bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &encoded)
	encoder.Write(raw)
	encoder.Close()
	return encoded.Bytes()
}

// Base64EncodeStr base64字符串加密
func Base64EncodeStr(raw string) string {
	return string(Base64Encode([]byte(raw)))
}

// Base64Decode base64解密
func Base64Decode(raw []byte) []byte {
	var buf bytes.Buffer
	buf.Write(raw)
	decoder := base64.NewDecoder(base64.StdEncoding, &buf)
	decoded, _ := ioutil.ReadAll(decoder)
	return decoded
}

// Base64DecodeStr base64字符串解密
func Base64DecodeStr(bs64str string) string {
	return string(Base64Decode([]byte(bs64str)))
}

// B64toHex 将base64字符串转换成HEX十六进制字符串
func Base64toHex(b64str string) (hexstr string) {
	b64map := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	sb := strings.Builder{}
	e := 0
	c := 0
	for _, r := range b64str {
		if r != '=' {
			v := strings.Index(b64map, string(r))
			if 0 == e {
				e = 1
				sb.WriteByte(int2char(v >> 2))
				c = 3 & v
			} else if 1 == e {
				e = 2
				sb.WriteByte(int2char(c<<2 | v>>4))
				c = 15 & v
			} else if 2 == e {
				e = 3
				sb.WriteByte(int2char(c))
				sb.WriteByte(int2char(v >> 2))
				c = 3 & v
			} else {
				e = 0
				sb.WriteByte(int2char(c<<2 | v>>4))
				sb.WriteByte(int2char(15 & v))
			}
		}
	}
	if e == 1 {
		sb.WriteByte(int2char(c << 2))
	}
	return sb.String()
}
