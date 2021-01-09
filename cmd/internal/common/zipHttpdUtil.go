package common

import (
	"os"
	fpath "path/filepath"
)

// ZipHttpdUtil は設定のユーティリティです。
type ZipHttpdUtil interface {
	// SetConfigDir は設定ファイルの置き場を設定します。
	SetConfigDir(configPath string)
	// ConfigDir は設定ファイルの置き場を取得します。
	ConfigDir() string
	// SetLogDir はログファイルの置き場を設定します。
	SetLogDir(logPath string)
	// LogDir はログファイルの置き場を取得します。
	LogDir() string
	// SetListenPort は代表ポートを設定します。
	SetListenPort(port int)
	// ListenPort は代表ポートを取得します。
	ListenPort() int
	// SetFirstDocPort はドキュメントグループのポートの開始番号を設定します。
	SetFirstDocPort(port int)
	// FirstDocPort はドキュメントグループのポートの開始番号を取得します。
	FirstDocPort() int
	// DefaultConfig は標準の設定ファイルの内容を取得します。
	DefaultConfig() string
}
type util struct {
	configPath   string
	logPath      string
	listenPort   int
	firstDocPort int
}

const (
	// DefaultListenPort はデフォルトのポート番号
	DefaultListenPort = 8823
	// DefaultFirstDocPort はグループの先頭ポート番号
	DefaultFirstDocPort = 58823
)

// NewUtil はコンストラクタです。
func NewUtil() ZipHttpdUtil {
	return &util{listenPort: 0, firstDocPort: 0}
}

// SetConfigDir は設定ファイルの置き場を設定します。
func (u *util) SetConfigDir(configPath string) {
	u.configPath = os.ExpandEnv(configPath)
}

// ConfigDir は設定ファイルの置き場を取得します。
func (u *util) ConfigDir() string {
	if u.configPath != "" {
		return u.configPath
	}
	// 無ければ実行ファイルのディレクトリ
	exe, _ := os.Executable()
	u.configPath = fpath.Dir(exe)
	return u.configPath
}

// SetLogDir はログファイルの置き場を設定します。
func (u *util) SetLogDir(logPath string) {
	u.logPath = logPath
}

// LogDir はログファイルの置き場を取得します。
func (u *util) LogDir() string {
	if u.logPath != "" {
		return u.logPath
	}
	// 無ければ設定ファイルの置き場/log
	u.logPath = fpath.Join(u.ConfigDir(), "log")
	return u.logPath
}

func (u *util) SetListenPort(port int) {
	u.listenPort = port
}

func (u *util) ListenPort() int {
	if u.listenPort != 0 {
		return u.listenPort
	}
	return DefaultListenPort
}

func (u *util) SetFirstDocPort(port int) {
	u.firstDocPort = port
}

func (u *util) FirstDocPort() int {
	if u.firstDocPort != 0 {
		return u.firstDocPort
	}
	return DefaultFirstDocPort
}

// DefaultConfig は標準の設定ファイルの内容を取得します。
func (u *util) DefaultConfig() string {
	// TODO: 標準の設定ファイルは Config から生成するように検討する。
	return `{
	"contentType": {
		"html,htm": "text/html",
		"js": "text/javascript",
		"json": "text/json",
		"css": "text/css",
		"txt": "text/plain",
		"md": "text/markdown",
		"bmp": "image/bmp",
		"gif": "image/gif",
		"ico": "image/ico",
		"jpg,jpg,jpeg,jfif,pjpeg,pjp": "image/jpeg",
		"png": "image/png",
		"svg": "image/svg+xml",
		"endterminator": null
	},
	"endterminator": null
}
`
}
