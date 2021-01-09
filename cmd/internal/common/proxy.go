package common

import (
	"html/template"
	"io"
	"net/http"
	"net/url"
)

// ResponseProxy はテストのため http.ResponseWriter をラップします
type ResponseProxy interface {
	// http.Redirect(w http.ResponseWriter, r *http.Request, url string, code int)
	Redirect(r RequestProxy, url string, code int)
	// http.Error(w http.ResponseWriter, error string, code int)
	Error(error string, code int)
	// (http.ResponseWriter).Head().Set(key string, value string)
	SetHeader(key string, value string)
	// (http.ResponseWriter).WriteHeader(statusCode int)
	WriteHeader(statusCode int)
	// io.Copy(dst io.Writer, src io.Reader) (written int64, err error)
	WriteContents(src io.Reader) (written int64, err error)
	// (io.Writer).Write(bytes []byte) (written int, err error)
	WriteContentsByte(bytes []byte) (written int, err error)
	// (*template.Template).Execute(wr io.Writer, data interface{}) error
	ParseContents(template *template.Template, data interface{}) error
}

// RequestProxy はテストのため http.Request をラップします
type RequestProxy interface {
	// r *http.Request
	Request() *http.Request
	// (http.Request).Host
	Host() string
	// Port
	Port() int
	// (http.Request).Method
	Method() string
	// (http.Request).URL.String() string
	URLString() string
	// (http.Request).URL.Path() string
	URLPath() string
	// RequestURI
	RequestURI() string
	// (http.Request).RemoteAddr
	RemoteAddr() string
	// Header().Get(key string) string
	GetHeader(key string) string
	// PostForm url.Values
	PostForm() url.Values
	// Form.Get(key string) string
	GetForm(key string) string
	// PostForm.Get(key string) string
	GetPostForm(key string) string
	// ParseForm() error
	ParseForm() error
}
