package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func GetOffsetFromHeader(h http.Header) int64 {
	byteRange := h.Get("range")
	if len(byteRange) < 7 {
		return 0
	}
	if byteRange[:6] != "bytes=" {
		return 0
	}
	bytePos := strings.Split(byteRange[6:], "-")
	offset, _ := strconv.ParseInt(bytePos[0], 0, 64)
	return offset
}

func GetHashFromHeader(h http.Header) string {
	digest := h.Get("digest")
	if len(digest) < 9 {
		return ""
	}
	if digest[:8] != "SHA-256=" {
		return ""
	}
	return digest[8:]
}

func GetSizeFromHeader(h http.Header) int64 {
	// 解释：
	// 1. h.Get("content-length") 获取请求头中的 content-length 字段
	// 2. strconv.ParseInt() 将字符串转换为 int64 类型
	// 3. 0 表示自动选择进制，64 表示 int64 类型
	size, _ := strconv.ParseInt(h.Get("content-length"), 0, 64)
	return size
}

func GetFileNameFromRequest(r *http.Request) string {
	fileName := strings.Split(r.URL.EscapedPath(), "/")[2]
	// 如果文件名为空，则从请求头中获取文件名
	if fileName == "" {
		fileName = r.Header.Get("filename")
	}
	return fileName
}

func CalculateHash(r io.Reader) string {
	h := sha256.New()
	io.Copy(h, r)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
