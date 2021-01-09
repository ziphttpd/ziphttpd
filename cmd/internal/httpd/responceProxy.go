package httpd

import (
	"html/template"
	"io"
	"net/http"

	"github.com/xorvercom/ziphttpd/cmd/internal/common"
)

type wInst struct {
	writer http.ResponseWriter
}

// NewResponseProxy は ResponseProxy を生成します。
func NewResponseProxy(writer http.ResponseWriter) common.ResponseProxy {
	return &wInst{writer: writer}
}

// http.Redirect(w http.ResponseWriter, r *http.Request, url string, code int)
func (i *wInst) Redirect(r common.RequestProxy, url string, code int) {
	http.Redirect(i.writer, r.Request(), url, code)
}

// http.Error(w http.ResponseWriter, error string, code int)
func (i *wInst) Error(error string, code int) {
	http.Error(i.writer, error, code)
}

// (http.ResponseWriter).Head().Set(key string, value string)
func (i *wInst) SetHeader(key string, value string) {
	i.writer.Header().Set(key, value)
}

// (http.ResponseWriter).WriteHeader(statusCode int)
func (i *wInst) WriteHeader(statusCode int) {
	i.writer.WriteHeader(statusCode)
}

// io.Copy(dst io.Writer, src io.Reader) (written int64, err error)
func (i *wInst) WriteContents(src io.Reader) (written int64, err error) {
	return io.Copy(i.writer, src)
}

// (io.Writer).Write(bytes []byte) (written int64, err error)
func (i *wInst) WriteContentsByte(bytes []byte) (written int, err error) {
	return i.writer.Write(bytes)
}

// (*template.Template).Execute(wr io.Writer, data interface{}) error
func (i *wInst) ParseContents(template *template.Template, data interface{}) error {
	return template.Execute(i.writer, data)
}
