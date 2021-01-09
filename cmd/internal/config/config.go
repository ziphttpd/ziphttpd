package config

import (
	"fmt"
	"io/ioutil"
	"os"
	fpath "path/filepath"
	"sort"
	"strings"

	"github.com/xorvercom/util/pkg/json"
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
	"github.com/xorvercom/ziphttpd/cmd/internal/model"
	"github.com/ziphttpd/zhsig/pkg/zhsig"
)

const (
	extConf               = ".json"
	fileConf              = "ziphttpd" + extConf
	portConf              = "portlockins" + extConf
	defaultDocument       = "docs"
	defaultAPIRootPath    = "api"
	defaultStaticRootPath = "static"
	systemHost            = "system"
	//commonHost            = "common"
	localHost = "localhost"
)

const (
	// ドキュメントの置き場
	docpathDocument = json.PathJSON("document")
	// Apiデータの置き場
	docpathAPIData = json.PathJSON("apidata")
	// 拡張子別の Content-Type の定義
	docpathContentType = json.PathJSON("contenttype")
	// バージョン表示するかの指定
	docpathShowVersion = json.PathJSON("showversion")
	// favicon.ico の指定
	docpathFavicon = json.PathJSON("favicon")
)

type conf struct {
	// 設定ファイルのディレクトリ
	configPath string
	// ドキュメント設定ファイルのディレクトリ
	docPath string
	// apiファイルの基準ディレクトリ
	apiRootPath string
	// ログ
	log common.Logger
	// 設定ファイルのエレメント
	element json.Element
	// ポート番号
	listenPort int
	// ドキュメントポート管理
	portMan common.PortMan
	// バージョン文言
	version string
	// 拡張子 - Content-Type 辞書
	contentTypes map[string]string
	// ホスト辞書
	hostDic map[common.HostName]common.DocHost
	// favicon.ico
	favicon []byte
	// トークン管理
	securityMan common.SecurityMan
	// タイトル情報
	titleMan *model.TitleMan
}

// newConf はコンストラクタです。
// わざわざOpenConfigと分離したのは単体テストのため。
func newConf(u common.ZipHttpdUtil) *conf {
	ret := &conf{
		hostDic:      map[common.HostName]common.DocHost{},
		contentTypes: map[string]string{},
		listenPort:   u.ListenPort(),
		version:      "",
		favicon:      nil,
		portMan:      model.NewPortMan(u.FirstDocPort()),
		securityMan:  model.NewSecurityMan(fpath.Join(u.ConfigDir(), "password.json")),
		titleMan:     model.NewTitleMan(),
	}
	return ret
}

// OpenConfig は設定を読みだします。
func OpenConfig(u common.ZipHttpdUtil) (common.Config, error) {
	var err error

	c := newConf(u)

	// 設定ファイルの置き場
	c.configPath = u.ConfigDir()
	// 無ければ作っておく
	err = os.MkdirAll(c.configPath, os.ModeDir)
	if err != nil {
		return nil, fmt.Errorf("error os.MkdirAll(%s) : %v", c.configPath, err)
	}

	logDir := u.LogDir()
	// 無ければ作っておく
	err = os.MkdirAll(logDir, os.ModeDir)
	if err != nil {
		return nil, fmt.Errorf("error os.MkdirAll(%s) : %v", logDir, err)
	}
	// ログの出力先
	c.log = common.NewLogger(logDir)

	// 設定ファイル
	configfile := fpath.Join(c.configPath, fileConf)
	_, err = os.Stat(configfile)
	if os.IsNotExist(err) {
		// 標準の設定ファイルを配置
		defaultZiphttpdconf := u.DefaultConfig()
		err := ioutil.WriteFile(configfile, []byte(defaultZiphttpdconf), os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("error write %s : %v", configfile, err)
		}
	}

	// 設定ファイルを読み込み
	c.element, err = json.LoadFromJSONFile(configfile)
	if err != nil {
		return nil, fmt.Errorf("error read %s : %v", configfile, err)
	}

	// ret.element -> conf
	c.setup()

	// デフォルトのポート
	err = c.portMan.Put(systemHost, c.listenPort)
	if err != nil {
		return nil, fmt.Errorf("error c.portMan.Put(%s, %d) : %v", systemHost, c.listenPort, err)
	}
	// システムのAPI用データは設定ファイルフォルダの下に作る
	c.hostDic[systemHost] = model.NewDocHost(c, systemHost)
	//c.groupsDic[systemHost] = model.NewDocGroup(c.PortMan(), systemHost, c.securityMan, apiPath)

	// ドキュメント設定ディレクトリが無ければ作っておく
	err = os.MkdirAll(c.docPath, os.ModeDir)
	if err != nil {
		return nil, fmt.Errorf("error os.MkdirAll(%s) : %v", c.docPath, err)
	}

	// ポートロックインファイル読み込み
	portsfile := fpath.Join(c.configPath, portConf)
	c.portMan.Load(portsfile)

	// ドキュメントのファイル jar, zip, zhd を全てチェックして対応する設定ファイルが無い場合には作成する
	// store から設定ファイルを作る
	c.readStore()
	// docpath から設定ファイルを作る
	c.readDocs()

	// タイトルのコピー
	c.titleFit()

	// ポートロックインファイル書き出し
	c.portMan.Save(portsfile)

	return c, nil
}

// readStore は zhget でダウンロードした ./store 以下のドキュメントのファイルから設定ファイルを作成します
func (c *conf) readStore() {
	store := fpath.Join(c.ConfigPath(), "store")
	// ホスト別ディレクトリ ./store/{ホスト}/ の検索
	hostdirs, err := ioutil.ReadDir(store)
	if err != nil {
		return
	}
	for _, hostdir := range hostdirs {
		if false == hostdir.IsDir() {
			// ディレクトリでない
			continue
		}
		hostname := hostdir.Name()
		host := zhsig.NewHost(c.ConfigPath(), hostname)

		// ホストのカタログを読む
		cat, err := zhsig.ReadCatalog(host.CatalogFile())
		if err != nil {
			continue
		}

		// ホスト追加
		docHost, ok := c.hostDic[hostname]
		if false == ok {
			docHost = model.NewDocHost(c, hostname)
			c.hostDic[hostname] = docHost
		}
		// ホストの書誌情報　(証明書、表示情報)
		hostTitle := c.titleMan.AddHost(hostname, cat.Peer)
		// ドキュメントグループ
		for groupname, group := range cat.Groups {
			// グループのドキュメント情報を登録
			docHost.Put(groupname, model.NewDocGroup(hostname, groupname))
			// グループのタイトル情報を収集
			groupTitle := hostTitle.AddGroup(groupname, group.Title, group.Description)
			for docname, doc := range group.Docs {
				// 署名ファイル読み出し
				sig, err := zhsig.ReadSig(host, docname)
				if err != nil {
					continue
				}
				// 実ファイル名
				// TODO: フォルダ階層変更 zipFileName := host.DocFile(groupname, docname, sig.File())
				zipFileName := host.File(docname, sig.File())
				checkext := strings.ToLower(fpath.Ext(zipFileName))
				if checkext != ".zip" && checkext != ".jar" && checkext != ".zhd" {
					continue
				}

				// ドキュメントのタイトル情報を収集
				groupTitle.AddDoc(docname, doc.Title, doc.Description)

				// 設定ファイルを作る
				basename := common.BaseName(zipFileName)
				basePath := fpath.Dir(zipFileName)
				confName, _ := fpath.Abs(fpath.Join(basePath, basename+extConf))
				if false == common.FileExists(confName) {
					defelem, err := model.NewDocConfig(c, zipFileName, basename)
					if err != nil {
						//return nil, fmt.Errorf("error NewDocData(%s,...) : %v", zipName, err)
						continue
					}
					// 保存する
					err = json.SaveToJSONFile(confName, defelem, true)
					if err != nil {
						//return nil, fmt.Errorf("error json.SaveToJSONFile(%s, %+v, true) : %v", docConfName, defelem, err)
					}
				}

				// 設定ファイル読み出し
				c.readConf(confName, hostname, groupname, docname)
			}
		}
	}
}

// readDocs は docpath に存在しているドキュメントのファイルから設定ファイルを作成します
func (c *conf) readDocs() {
	// グループの書誌情報を収集
	hostTitle := c.titleMan.AddHost(localHost, nil)
	groupTitle := hostTitle.AddGroup(localHost, "COMMON", "localhost document")
	// ./docs のファイルを検索
	files, _ := ioutil.ReadDir(c.docPath)
	for _, f := range files {
		// ディレクトリは除く
		if f.IsDir() {
			continue
		}
		zipFileName := f.Name()

		// アーカイブ以外は除く
		checkext := strings.ToLower(fpath.Ext(zipFileName))
		if checkext != ".zip" && checkext != ".jar" && checkext != ".zhd" {
			continue
		}

		basename := common.BaseName(zipFileName)

		// ドキュメントのタイトル情報を収集
		groupTitle.AddDoc(basename, basename, "")

		// 設定ファイルを作る
		docConfName := basename + extConf
		confPath := fpath.Join(c.docPath, docConfName)
		if false == common.FileExists(confPath) {
			zipFilePath := fpath.Join(c.docPath, zipFileName)
			defelem, err := model.NewDocConfig(c, zipFilePath, basename)
			if err != nil {
				//return nil, fmt.Errorf("error NewDocData(%s,...) : %v", zipName, err)
				continue
			}
			// 保存する
			err = json.SaveToJSONFile(confPath, defelem, true)
			if err != nil {
				//return nil, fmt.Errorf("error json.SaveToJSONFile(%s, %+v, true) : %v", docConfName, defelem, err)
			}
		}

		// 設定ファイル読み出し
		c.readConf(confPath, localHost, "", "")
	}
}

// readConf はドキュメントの設定ファイルを読みます
func (c *conf) readConf(confFileName, hostname, groupname, docname string) {
	// ドキュメントの設定ファイルを読む
	docdata, err := model.OpenDocConfig(c, confFileName, hostname, groupname, docname)
	if err != nil {
		// ドキュメントが読み込めない
		c.log.Warn(fmt.Sprintf("read error docname:%s : %+v", confFileName, err))
		return
	}
	docid := docdata.DocID()
	// ホスト追加
	docHost, ok := c.hostDic[hostname]
	if false == ok {
		docHost = model.NewDocHost(c, hostname)
		c.hostDic[hostname] = docHost
	}

	// ドキュメントポートグループの決定
	docGroupName := docdata.DocGroupName()
	docGroup := docHost.Get(docGroupName)
	if nil == docGroup { // nolint:gosimple
		// 未登録
		// ドキュメントグループ情報を登録
		docGroup = model.NewDocGroup(hostname, docGroupName)
		docHost.Put(docGroupName, docGroup)
	}
	docGroup.Put(docid, docdata)
	// 静的ファイル
	if docdata.UseStaticFiles() {
		folder := fpath.Join(c.configPath, "static", hostname, docGroupName, docid)
		os.MkdirAll(folder, os.ModeDir)
	}
}

// setup は conf.element の内容を読みだします。
func (c *conf) setup() {
	// 拡張子 - Content-Type 辞書
	if elem, ok := json.QueryElemObject(c.element, docpathContentType); ok {
		keys := elem.Keys()
		for _, exts := range keys {
			// Content-Type
			typestr := elem.Child(exts).Text()
			for _, ext := range strings.Split(strings.ToLower(exts), ",") {
				// 拡張子
				nmlExt := strings.TrimSpace(ext)
				c.contentTypes[nmlExt] = typestr
			}
		}
	}

	// ドキュメント設定ファイルのディレクトリ
	var docPath string
	if elem, ok := json.QueryElemString(c.element, docpathDocument); ok {
		docPath = elem.Text()
		if false == fpath.IsAbs(docPath) { // nolint:gosimple
			docPath = fpath.Join(c.configPath, docPath)
		}
	} else {
		docPath = fpath.Join(c.configPath, defaultDocument)
	}
	c.docPath = docPath

	// apiデータファイルの基準ディレクトリ
	var apiRootPath string
	if elem, ok := json.QueryElemString(c.element, docpathAPIData); ok {
		apiRootPath = elem.Text()
		if false == fpath.IsAbs(apiRootPath) { // nolint:gosimple
			apiRootPath = fpath.Join(c.configPath, apiRootPath)
		}
	} else {
		apiRootPath = fpath.Join(c.configPath, defaultAPIRootPath)
	}
	c.apiRootPath = apiRootPath

	// バージョン
	if elem, ok := json.QueryElemBool(c.element, docpathShowVersion); ok {
		if elem.Bool() {
			c.version = " (ver." + version() + ")"
		}
	}

	// favicon
	if elem, ok := json.QueryElemString(c.element, docpathFavicon); ok {
		fav := elem.Text()
		if false == fpath.IsAbs(fav) { // nolint:gosimple
			fav = fpath.Join(c.configPath, fav)
		}
		f, e := os.OpenFile(fav, os.O_RDONLY, os.ModePerm)
		if e == nil {
			buf, e := ioutil.ReadAll(f)
			if e == nil {
				c.favicon = buf
			}
		}
	}
}

// titleFit は TitleMan に集めたタイトル情報をドキュメントツリーに設定します。
func (c *conf) titleFit() {
	// TODO: nil エラーハンドリング
	logger := c.Logger()
	hosts := json.NewElemObject()
	for id, host := range c.hostDic {
		hosts.Put(id, host.JSON())
	}
	logger.Info(json.ToJSON(hosts, true))
	tm := c.titleMan
	logger.Info(json.ToJSON(tm.JSON(), true))
	for hostName, host := range c.hostDic {
		thost := tm.Host(hostName)
		host.SetTitleInfo(hostName, "")
		for _, groupName := range host.Ids() {
			group := host.Get(groupName)
			tgroup := thost.Group(group.Name())
			group.SetTitleInfo(tgroup.Title(), tgroup.Description())
			for _, docID := range group.Ids() {
				doc := group.Get(docID)
				tdoc := tgroup.Doc(docID)
				doc.SetTitleInfo(tdoc.Title(), tdoc.Description())
			}
		}
	}
}

func (c *conf) Logger() common.Logger {
	return c.log
}

// DocPath はドキュメントの基準フォルダを取得します。
func (c *conf) DocPath() string {
	return c.docPath
}

// ListenPort は標準のポート番号を取得します。
func (c *conf) ListenPort() int {
	return c.listenPort
}

// Favicon はfavicon.icoを取得します。
func (c *conf) Favicon() []byte {
	return c.favicon
}

// DocHost はホスト名称から DocHost を返却します。
func (c *conf) DocHost(hostname common.HostName) common.DocHost {
	return c.hostDic[hostname]
}

// ContentType はファイルの拡張子から Content-Type を取得します。
func (c *conf) ContentType(filepath string) string {
	ext := strings.ToLower(strings.TrimPrefix(fpath.Ext(filepath), "."))
	if val, ok := c.contentTypes[ext]; ok {
		return val
	}
	// 設定が無かったので空文字列
	return ""
}

// Version はバージョンの文言を取得します。
func (c *conf) Version() string {
	return c.version
}

// Close はドキュメントをクローズします。
func (c *conf) Close() {
	for _, hs := range c.hostDic {
		hs.Close()
	}
	c.hostDic = map[string]common.DocHost{}
}

// PortMan はポートマネージャを取得します。
func (c *conf) PortMan() common.PortMan {
	return c.portMan
}

// ConfigPath は設定ファイルのフォルダを取得します。
func (c *conf) ConfigPath() string {
	return c.configPath
}

// トークン管理
func (c *conf) SecurityMan() common.SecurityMan {
	return c.securityMan
}

// APIパス
func (c *conf) APIPath(host common.HostName) string {
	return fpath.Join(c.apiRootPath, host)
}

// タイトル管理
func (c *conf) HostTitle(name string) common.HostTitle {
	return c.titleMan.Host(name)
}

// ホスト名一覧
func (c *conf) HostNames() []common.HostName {
	keys := []common.HostName{}
	for k := range c.hostDic {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (c *conf) String() string {
	return fmt.Sprintf("directory : %s", c.configPath)
}
