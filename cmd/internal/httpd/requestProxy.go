package httpd

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

type rInst struct {
	request *http.Request
	size    int64
}

// NewRequestProxy は RequestProxy を作成します
func NewRequestProxy(request *http.Request) common.RequestProxy {
	return &rInst{request: request, size: 100 * 1024}
}

// r *http.Request
func (i *rInst) Request() *http.Request {
	return i.request
}

// (http.Request).Host
func (i *rInst) Host() string {
	return i.request.Host
}

func (i *rInst) Port() int {
	hosts := strings.Split(i.request.Host, ":")
	port, err := strconv.Atoi(hosts[1])
	if err != nil {
		port = 0
	}
	return port
}

// (http.Request).Method
func (i *rInst) Method() string {
	return i.request.Method
}

// (http.Request).URL.String() string
func (i *rInst) URLString() string {
	return i.request.URL.String()
}

// (http.Request).URL.Path() string
func (i *rInst) URLPath() string {
	return i.request.URL.Path
}

// RequestURI
func (i *rInst) RequestURI() string {
	return i.request.RequestURI
}

// (http.Request).RemoteAddr
func (i *rInst) RemoteAddr() string {
	return i.request.RemoteAddr
}

// Header().Get(key string) string
func (i *rInst) GetHeader(key string) string {
	return i.request.Header.Get(key)
}

// PostForm url.Values
func (i *rInst) PostForm() url.Values {
	return i.request.PostForm
}

// Form.Get(key string) string
func (i *rInst) GetForm(key string) string {
	return i.request.Form.Get(key)
}

// PostForm.Get(key string) string
func (i *rInst) GetPostForm(key string) string {
	return i.request.PostForm.Get(key)
}

// SetParseMemory は ParseMultipartForm で指定するメモリサイズを変更します。
func (i *rInst) SetParseMemory(size int64) {
	i.size = size
}

// ParseForm() error
func (i *rInst) ParseForm() error {
	v := i.request.Header.Get("Content-Type")
	if pos := strings.Index(v, ";"); pos != -1 {
		v = v[:pos]
	}
	if v == "multipart/form-data" {
		return i.request.ParseMultipartForm(i.size)
	}
	return i.request.ParseForm()
}
