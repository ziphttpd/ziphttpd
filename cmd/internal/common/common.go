package common

import (
	"net"
	"time"

	"github.com/xorvercom/util/pkg/json"
	"github.com/xorvercom/util/pkg/zip"
	"github.com/ziphttpd/zhsig/pkg/zhsig"
)

// DocID は zip ドキュメントを特定する識別です。
type DocID = string

// HostName はホストの名称です。
type HostName = string

// DocGroupName はドキュメントグループの名称です。
type DocGroupName = string

// Token はトークンです。
type Token = string

// Logger はログ出力を管理します。
type Logger interface {
	Info(msg string)
	Infof(format string, arg ...interface{})
	Warn(msg string)
	Warnf(format string, arg ...interface{})
}

// Server はサーバです。
type Server interface {
	// サーバを開始します。
	Run()
	// 接続数を減らします
	ConnDone()
	// ポート番号を返します
	Port() int
}

// Config は ziphttpd.json で記述された設定内容を提供します。
type Config interface {
	// ログ
	Logger() Logger
	// Close はドキュメントをクローズします。
	Close()
	// DocPath はドキュメントの基準フォルダを取得します。
	DocPath() string
	// ListenPort はシステム管理用のポート番号を取得します。
	ListenPort() int
	// Favicon はfavicon.icoを取得します。
	Favicon() []byte
	// DocHost はホスト名称から DocHost を返却します。
	DocHost(hostname HostName) DocHost
	// ContentType はファイルの拡張子から Content-Type を取得します。
	ContentType(filepath string) string
	// Version はバージョンの文言を取得します。
	Version() string
	// APIPath はAPIのストレージを返します。
	APIPath(host HostName) string
	// PortMan はポートマネージャを取得します。
	PortMan() PortMan
	// ConfigPath は設定ファイルのフォルダを取得します。
	ConfigPath() string
	// トークン管理
	SecurityMan() SecurityMan
	// タイトル管理
	//HostTitle(name string) HostTitle
	// ホスト名一覧
	HostNames() []HostName
}

// PortMan はポートを管理します。ポートはドキュメントグループの名称で管理します。
type PortMan interface {
	// Port はドキュメントグループ名のポートを返します。未登録ならば空いているポートを探して確保します。
	Port(host HostName) int
	// DocGroupName はポート番号のドキュメントグループ名称を返します。
	HostName(port int) HostName
	// PutLockIn は以前に利用していたポート番号を予約します。
	PutLockIn(host HostName, port int)
	// HostNames はグループ名を返します。
	HostNames() []HostName
	// Listener はリスナーを返します。
	Listener(host HostName) *net.TCPListener
	// Put はポートを登録します。固定ポートの登録時に使用します。
	Put(host HostName, port int) error
	// Close は全てのリスナをクローズします。
	Close()
	// Load はグループで使用するポートをポートロックインファイルから読みだします。
	Load(portsfile string)
	// Save はポートグループで使用しているポートをポートロックインファイルに書き出します。
	Save(portsfile string)
}

// SecurityMan はセキュリティを管理します。
type SecurityMan interface {
	// Token はドキュメントグループにトークンを振り出します。
	Token(hostName HostName) Token
	// IsValid はドキュメントグループのパスワードをチェックします。
	IsValid(hostName HostName, password string) bool
	// LoadPassword はパスワードを読み込みます。
	LoadPassword(passwordfile string)
	// UseLocalStorage はドキュメントグループでのCSRFトークンをlocalStorageで行うかを返します。
	UseLocalStorage(hostName HostName) bool
}

// ContentTypeer はファイルの拡張子から Content-Type を取得します。
type ContentTypeer interface {
	// ContentType はファイルの拡張子から Content-Type を取得します。
	ContentType(filepath string) string
}

// DocFileInfo はファイルの情報です。
type DocFileInfo interface {
	// IsDir は対象がディレクトリの時に真を返します。
	IsDir() bool
	// ModTime は更新時刻を返します。
	ModTime() time.Time
	// Size はファイルのサイズを返します。
	Size() string
}

// DocData はドキュメント（ホストしているzip）を管理します。
type DocData interface {
	ContentTypeer
	// DocID は zip ドキュメントの識別を返します。
	DocID() DocID
	// DocRoot はドキュメントルートのパス文字列を返します。
	DocRoot() string
	// DocGroup は所属しているグループを取得します。
	DocGroupName() DocGroupName
	// Encoding は text/xxx の Content-Encoding　に設定する文字列を返します。
	Encoding() string
	// UseStaticFiles は静的ファイルを利用するかを取得します。
	UseStaticFiles() bool
	// ZipDic はZipファイル辞書を返します。
	ZipDic() zip.Dictionary
	// ZipPath はzipファイルのパスを返します。
	ZipPath() string
	// FilePaths はドキュメント内のファイルパスの一覧(zip, static)を返します。
	FilePaths() []string
	// FileInfo はファイルパスの DocFileInfo を返します。
	FileInfo(filepath string) (DocFileInfo, error)
	// Close はドキュメントをクローズします。
	Close()
	// SetTitleInfo はタイトル情報を設定します。
	SetTitleInfo(title, description string)
	// Title はタイトルを返します。
	Title() string
	// Description は説明を返します。
	Description() string
	JSON() json.ElemObject
}

// DocGroup はドキュメントのグループです。
type DocGroup interface {
	// Name はドキュメントグループ名称を取得します。
	Name() DocGroupName
	// Put は zip ドキュメントを追加します。
	Put(docid DocID, doc DocData)
	// Get は zip ドキュメントを取得します。
	Get(docid DocID) DocData
	// Ids はホストしている zip ドキュメントの名前の一覧を取得します。
	Ids() []DocID
	// Close はホストしている zip ドキュメントをクローズします。
	Close()
	// SetTitleInfo はタイトル情報を設定します。
	SetTitleInfo(title, description string)
	// Title はタイトルを返します。
	Title() string
	// Description は説明を返します。
	Description() string
	JSON() json.ElemObject
}

// DocHost はドキュメントホストです。
type DocHost interface {
	// Name はホスト名称を取得します。
	Name() HostName
	// Port はポート番号を取得します。
	Port() string
	// Put はドキュメントグループを追加します。
	Put(groupid DocGroupName, group DocGroup)
	// Get はドキュメントグループを取得します。
	Get(groupid DocGroupName) DocGroup
	// Ids はホストしているドキュメントグループの名前の一覧を取得します。
	Ids() []DocGroupName
	// Close はホストしている zip ドキュメントをクローズします。
	Close()
	// Token はCSRFトークンを返します
	Token() Token
	// GetApi は WebAPI ロジックで使用するフォルダを返します。
	GetAPIPath() string
	GetAPI() API
	SetAPI(API)
	// SetTitleInfo はタイトル情報を設定します。
	SetTitleInfo(title, description string)
	// Title はタイトルを返します。
	Title() string
	// Description は説明を返します。
	Description() string
	JSON() json.ElemObject
}

// API は API のロジック
type API interface {
	// Execute は API ロジックを同期実行する
	Execute(text string) (string, error)
	// Attach は非同期の情報取得を行います（未実装）
	// TODO: デタッチできるように再検討
	Attach() <-chan json.ElemObject
	// 強制終了
	Terminate()
}

// DocTitle はドキュメントのタイトル情報
type DocTitle interface {
	// タイトル
	Title() string
	// 注釈
	Description() string
	JSON() json.ElemObject
}

// GroupTitle はドキュメントグループのタイトル情報
type GroupTitle interface {
	// タイトル
	Title() string
	// 注釈
	Description() string
	// ドキュメントのタイトル情報
	Doc(name string) DocTitle
	// AddDoc はドキュメントを追加します。
	AddDoc(name, title, description string) DocTitle
	JSON() json.ElemObject
}

// HostTitle はホストのタイトル情報
type HostTitle interface {
	// ピア情報
	Peer() *zhsig.PeerInfo
	// ドキュメントグループのタイトル情報
	Group(name string) GroupTitle
	// AddGroup はドキュメントグループを追加します。
	AddGroup(name, title, description string) GroupTitle
	JSON() json.ElemObject
}

// Param パラメータ
type Param interface {
	Config() Config
	PortMan() PortMan
	//Port2DocGroup(int) DocGroup
	//Name2DocGroup(string, string) DocGroup
	DocHost() DocHost
	DocGroup() DocGroup
	DocData() DocData
	Logger() Logger
	SecurityMan() SecurityMan
	Server() Server
	ListenPort() int
	Version() string
	ConfigPath() string
	Favicon() []byte
	Paths() []string
	AdURL() string
}
