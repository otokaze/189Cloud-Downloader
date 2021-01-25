package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

const (
	RsaPublicKey = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDY7mpaUysvgQkbp0iIn2ezoUyh
i1zPFn0HCXloLFWT7uoNkqtrphpQ/63LEcPz1VYzmDuDIf3iGxQKzeoHTiVMSmW6
FlhDeqVOG094hFJvZeK4OzA6HVwzwnEW5vIZ7d+u61RV1bsFxmB68+8JXs3ycGcE
4anY+YzZJcyOcEGKVQIDAQAB
-----END PUBLIC KEY-----`
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func int2char(i int) (r byte) {
	return "0123456789abcdefghijklmnopqrstuvwxyz"[i]
}

func GenNoCacheNum() string {
	noCache := &strings.Builder{}
	fmt.Fprintf(noCache, "0.%d", rand.Int63n(1e17))
	return noCache.String()
}

func Timestamp() int {
	// millisecond
	return int(time.Now().UTC().UnixNano() / 1e6)
}

// SignatureOfMd5 MD5签名
func SignatureOfMd5(params map[string]string) string {
	keys := []string{}
	for k, v := range params {
		keys = append(keys, k+"="+v)
	}

	// sort
	sort.Strings(keys)

	signStr := strings.Join(keys, "&")

	h := md5.New()
	h.Write([]byte(signStr))
	return hex.EncodeToString(h.Sum(nil))
}

// SignatureOfHmac HMAC签名
func SignatureOfHmac(secretKey, sessionKey, operate, url, dateOfGmt string) string {
	requestUri := strings.Split(url, "?")[0]
	requestUri = strings.ReplaceAll(requestUri, "https://", "")
	requestUri = strings.ReplaceAll(requestUri, "http://", "")
	idx := strings.Index(requestUri, "/")
	requestUri = requestUri[idx:]

	plainStr := &strings.Builder{}
	fmt.Fprintf(plainStr, "SessionKey=%s&Operate=%s&RequestURI=%s&Date=%s",
		sessionKey, operate, requestUri, dateOfGmt)

	key := []byte(secretKey)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(plainStr.String()))
	return strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))
}

func Rand() string {
	randStr := &strings.Builder{}
	fmt.Fprintf(randStr, "%d_%d", rand.Int63n(1e5), rand.Int63n(1e10))
	return randStr.String()
}

// PcClientInfoSuffixParam PC客户端URL请求后缀信息
func PcClientInfoSuffixParam() string {
	return "clientType=TELEPC&version=6.2&channelId=web_cloud.189.cn&rand=" + Rand()
}

func DateOfGmtStr() string {
	return time.Now().UTC().Format(http.TimeFormat)
}

func XRequestId() string {
	u4 := uuid.NewV4()
	return strings.ToUpper(u4.String())
}

func Uuid() string {
	u4 := uuid.NewV4()
	return u4.String()
}

// CheckFileNameValid 检测文件名是否有效，包含特殊字符则无效
func CheckFileNameValid(name string) bool {
	if name == "" {
		return true
	}
	return !strings.ContainsAny(name, "\\/:*?\"<>|")
}

// FormatFileSize 格式化文件大小
func FormatFileSize(fileSize int64) string {
	if fileSize < 1024 {
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}
