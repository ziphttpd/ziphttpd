package httpd

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/xorvercom/util/pkg/easywork"
	"github.com/xorvercom/ziphttpd/cmd/internal/common"
	"github.com/xorvercom/ziphttpd/cmd/internal/handler"
)

type serv struct {
	mu sync.Mutex
	// WaitGroup
	activeConWg sync.WaitGroup
	// connection count
	numberOfActive int
	// 設定
	conf common.Config
	// 待ち受けリスナ
	listener *net.TCPListener
	// ポート
	port int
	// ホスト名
	hostName common.HostName
}

// NewServer はサーバを作成します。
func NewServer(conf common.Config, hostName common.HostName) easywork.Runnable /* *serv */ {
	log := conf.Logger()
	log.Infof("create server:%s", hostName)
	return &serv{
		conf:           conf,
		numberOfActive: 0,
		//docGroupName:   hostName,
		listener: conf.PortMan().Listener(hostName),
		//docGroup:       conf.DocGroup(hostName),
		port:     conf.PortMan().Port(hostName),
		hostName: hostName,
	}
}

// サーバを開始します。
func (s *serv) Run() {
	s.conf.Logger().Infof("run server:%s, port:%d", s.hostName, s.port)

	// HTTPサーバ
	srv := &http.Server{Handler: s, ConnState: func(c net.Conn, st http.ConnState) {
		pre := s.numberOfActive
		if st == http.StateActive {
			s.ConnAdd()
		} else if st == http.StateIdle || st == http.StateHijacked {
			s.ConnDone()
		}
		s.conf.Logger().Infof("ConState: %+v :  %+v : %d -> %d", s, st, pre, s.numberOfActive)
	}}

	// ループ開始
	var err error
	err = srv.Serve(s.listener)
	if err != nil {
		s.conf.Logger().Warnf("server:%s : %+v", s.hostName, err)
		//		panic(fmt.Errorf("web server error : %+v", err))
	}

	// 抜けてきたら接続数が無くなるまで待つ
	s.activeConWg.Wait()
}

// 接続数を増やします
func (s *serv) ConnAdd() {
	s.activeConWg.Add(1)
	s.numberOfActive++
}

// 接続数を減らします
func (s *serv) ConnDone() {
	s.activeConWg.Done()
	s.numberOfActive--
}

// ポート番号を返します
func (s *serv) Port() int {
	return s.port
}

// ServeHTTP は http.Handler の実装メソッド。
func (s *serv) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.ServeHTTPinner(NewResponseProxy(writer), NewRequestProxy(request))
}

// テストの利便性から http.ResponseWriter などは隠ぺいした
func (s *serv) ServeHTTPinner(writer common.ResponseProxy, request common.RequestProxy) {
	conf := s.conf
	log := conf.Logger()
	log.Infof("%+v", request.Request())

	request.ParseForm()
	switch request.Method() {
	case "POST":
		log.Infof("body:%+v", request.PostForm())
	case "OPTIONS":
		// CORS プリフライト禁止、つまりクロスオリジンのアクセスは禁止
		// 400 Bad Request
		p := &param{conf: conf, paths: nil, server: s}
		handler.ErrorHandler(writer, request, p, http.StatusBadRequest)
		return
	}

	// パスのurlエンコーディング(%xxとか)の解除
	urlpath, _ := url.QueryUnescape(request.URLPath())
	log.Infof("access %s : %s", s.hostName, urlpath)
	// パラメータ
	p := &param{
		conf:   conf,
		paths:  strings.Split(strings.TrimSpace(urlpath), "/"),
		server: s,
	}

	// 特殊なホスト
	switch strings.ToLower(p.paths[1]) {
	case "":
		// リクエストされたのはトップディレクトリだった
		handler.TopHandler(writer, request, p)
		return
	case "ad":
		// TODO: 広告のiframeテスト用
		// リクエストされたのは広告だった
		handler.AdHandler(writer, request, p)
		return
	case "api":
		// リクエストされたのはwebapiだった
		handler.APIHandler(writer, request, p)
		return
	case "files":
		// リクエストされたのはファイル一覧
		handler.FilesHandler(writer, request, p)
		return
	case "login":
		// LoginHandler は、パスワード検査の関係上、同時には一個しか処理しない
		s.mu.Lock()
		defer s.mu.Unlock()
		// リクエストされたのはloginだった
		handler.LoginHandler(writer, request, p)
		return
	case "favicon.ico":
		// リクエストされたのはfavicon.icoだった
		handler.FaviconHandler(writer, request, p)
		return
	}

	// ドキュメントを特定
	pathlen := len(p.paths)
	// ホスト
	p.docHost = conf.DocHost(p.paths[1])
	if p.docHost == nil {
		handler.ErrorHandler(writer, request, p, http.StatusNotFound)
		return
	}
	// ドキュメントグループ名
	if pathlen > 2 {
		p.docGroup = p.docHost.Get(p.paths[2])
		// ドキュメント名
		if pathlen > 3 {
			p.docData = p.docGroup.Get(p.paths[3])
		}
	}

	// フォルダ判定。/で終わっているならpathsの最後の要素は空
	isDir := p.docData == nil || p.paths[pathlen-1] == ""
	if isDir {
		// リクエストされたのはディレクトリだった
		handler.DirHandler(writer, request, p)
	} else {
		// リクエストされたのはファイルだった
		handler.DocHandler(writer, request, p)
	}
}
