package common

import (
	"fmt"
	"log"
	"os"
	fpath "path/filepath"
)

const (
	hops = 2
)

// LoggerInst は Logger の実体です。
type LoggerInst struct {
	log *log.Logger
}

// NewLogger は Loger を生成します。
func NewLogger(logPath string) *LoggerInst {
	var file *os.File
	var err error
	if file, err = os.OpenFile(fpath.Join(logPath, "ziphttpd.log"), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666); err != nil {
		panic(err)
	}
	return &LoggerInst{log: log.New(file, "", log.Ldate|log.Ltime|log.Llongfile)}
}

// Info は [INFO] ログを出力します。
func (l *LoggerInst) Info(msg string) {
	l.log.Output(hops, fmt.Sprintf("[INFO] %s", msg))
}

// Infof は [INFO] ログを出力します。
func (l *LoggerInst) Infof(format string, arg ...interface{}) {
	l.log.Output(hops, fmt.Sprintf("[INFO] "+format, arg...))
}

// Warn は [WARN] ログを出力します。
func (l *LoggerInst) Warn(msg string) {
	l.log.Output(hops, fmt.Sprintf("[WARN] %s", msg))
}

// Warnf は [WARN] ログを出力します。
func (l *LoggerInst) Warnf(format string, arg ...interface{}) {
	l.log.Output(hops, fmt.Sprintf("[WARN] "+format, arg...))
}
